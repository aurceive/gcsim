package servermode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/genshinsim/gcsim/pkg/model"
)

type resultResponse struct {
	Result string `json:"result"`
	Done   bool   `json:"done"`
	Error  string `json:"error"`
}

func testServer(t *testing.T) *Server {
	t.Helper()
	s, err := New(WithDefaults(), WithWorkers(2), WithTimeout(10*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func request(s *Server, method, path string, body []byte) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(body)))
	return w
}

func decodeResult(t *testing.T, w *httptest.ResponseRecorder) resultResponse {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("results status = %d, body = %q", w.Code, w.Body.String())
	}
	var result resultResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestResultsWaitForInitialization(t *testing.T) {
	s := testServer(t)
	wk := &worker{
		cancel: make(chan bool),
		ready:  make(chan struct{}),
	}
	s.pool["waiting"] = wk
	defer wk.markReady()
	response := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		response <- request(s, http.MethodGet, "/results/waiting", nil)
	}()
	select {
	case w := <-response:
		t.Fatalf("result returned before initialization: %d %q", w.Code, w.Body.String())
	case <-time.After(50 * time.Millisecond):
	}

	wk.mu.Lock()
	wk.result = &model.SimulationResult{Config: "initialized"}
	wk.mu.Unlock()
	wk.markReady()
	select {
	case w := <-response:
		result := decodeResult(t, w)
		if result.Done || result.Error != "" || !strings.Contains(result.Result, "initialized") {
			t.Fatalf("unexpected result: %+v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("result request did not resume after initialization")
	}
}

func TestCancelBeforeInitializationUnblocksResults(t *testing.T) {
	s := testServer(t)
	wk := &worker{
		cancel: make(chan bool),
		ready:  make(chan struct{}),
	}
	s.pool["cancel"] = wk
	w := request(s, http.MethodPost, "/cancel/cancel", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("cancel status = %d", w.Code)
	}
	result := decodeResult(t, request(s, http.MethodGet, "/results/cancel", nil))
	if !result.Done || result.Error != "simulation cancelled" {
		t.Fatalf("unexpected cancellation result: %+v", result)
	}
}

func TestCancelAfterInitializationKeepsPartialResult(t *testing.T) {
	s := testServer(t)
	wk := &worker{
		cancel: make(chan bool),
		ready:  make(chan struct{}),
		result: &model.SimulationResult{Config: "partial"},
	}
	wk.markReady()
	s.pool["cancel"] = wk
	w := request(s, http.MethodPost, "/cancel/cancel", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("cancel status = %d", w.Code)
	}
	result := decodeResult(t, request(s, http.MethodGet, "/results/cancel", nil))
	if !result.Done || result.Error != "" || !strings.Contains(result.Result, "partial") {
		t.Fatalf("unexpected cancellation result: %+v", result)
	}
}

func TestImmediateResultsForWeaponSimulation(t *testing.T) {
	for _, weapon := range []string{"starsealededict", "sacfragments"} {
		t.Run(weapon, func(t *testing.T) {
			s := testServer(t)
			config := fmt.Sprintf(`
sucrose char lvl=90/90 cons=6 talent=9,9,9;
sucrose add weapon=%q refine=1 lvl=90/90;
sucrose add set="viridescentvenerer" count=4;
sucrose add stats hp=4780 atk=311 em=187 em=187 em=187;
options swap_delay=12 iteration=3;
active sucrose;
target lvl=100 resist=0.1 radius=2 pos=0,2.4 hp=999999999;
sucrose burst, skill, attack;
`, weapon)
			body, err := json.Marshal(map[string]string{"config": config})
			if err != nil {
				t.Fatal(err)
			}
			w := request(s, http.MethodPost, "/run/weapon", body)
			if w.Code != http.StatusOK {
				t.Fatalf("run status = %d, body = %q", w.Code, w.Body.String())
			}

			deadline := time.Now().Add(10 * time.Second)
			for {
				result := decodeResult(t, request(s, http.MethodGet, "/results/weapon", nil))
				if result.Error != "" {
					t.Fatalf("simulation error: %s", result.Error)
				}
				if result.Result == "" || !json.Valid([]byte(result.Result)) {
					t.Fatalf("invalid simulation result: %q", result.Result)
				}
				if result.Done {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("simulation did not finish")
				}
				time.Sleep(time.Millisecond)
			}
		})
	}
}

func TestInitializationErrorIsReported(t *testing.T) {
	s := testServer(t)
	body := []byte(`{"config":"invalid config"}`)
	w := request(s, http.MethodPost, "/run/invalid", body)
	if w.Code != http.StatusOK {
		t.Fatalf("run status = %d, body = %q", w.Code, w.Body.String())
	}
	result := decodeResult(t, request(s, http.MethodGet, "/results/invalid", nil))
	if !result.Done || result.Error == "" {
		t.Fatalf("expected initialization error, got %+v", result)
	}
}
