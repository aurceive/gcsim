package servermode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/simulator"
	"github.com/go-chi/chi"
)

func (s *Server) ready() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		running := s.isRunning(id)
		if running {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) running() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		running := s.isRunning(id)
		if running {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) validate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				w.Write([]byte(errorRecover(r).Error()))
				w.WriteHeader(http.StatusBadRequest)
			}
		}()
		id := chi.URLParam(r, "id")
		s.Log.Info("request to run sample", "id", id)
		var payload struct {
			Config string `json:"config"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			s.Log.Info("body did not decode to json", "id", id, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		file := ast.NewFile()
		cfg, _, err := simulator.Parse(file, payload.Config)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		marshalled, err := json.Marshal(cfg)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) sample() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				w.Write([]byte(errorRecover(r).Error()))
				w.WriteHeader(http.StatusBadRequest)
			}
		}()
		id := chi.URLParam(r, "id")
		s.Log.Info("request to run sample", "id", id)
		var payload struct {
			Config string `json:"config"`
			Seed   uint64 `json:"seed"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			s.Log.Info("body did not decode to json", "id", id, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		opts := simulator.Options{
			GZIPResult:       false,
			ResultSaveToPath: "",
			ConfigPath:       "",
		}
		data, err := simulator.GenerateSampleWithSeed(payload.Config, payload.Seed, opts)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		marshalled, err := data.MarshalJSON()
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(marshalled)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) run() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		s.Log.Info("request to run sim", "id", id)
		var payload struct {
			Config string `json:"config"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			s.Log.Info("body did not decode to json", "id", id, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		wk := &worker{
			id:     id,
			cfg:    payload.Config,
			log:    s.Log,
			cancel: make(chan bool),
			ready:  make(chan struct{}),
		}
		s.Lock()
		if _, ok := s.pool[id]; ok {
			s.Unlock()
			s.Log.Info("run request failed; already running", "id", id)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("already running!!"))
			return
		}
		s.pool[id] = wk
		s.Unlock()
		s.Log.Info("config decoded ok; running", "id", id)

		// start worker
		go wk.run(s.WorkerCount, s.FlushInterval)

		// add a timeout
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
			defer cancel()
			select {
			case <-wk.cancel:
				return
			case <-ctx.Done():
				wk.mu.Lock()
				if !wk.done {
					wk.done = true
					if wk.err == nil {
						wk.err = fmt.Errorf("execution timed out after %s", s.Timeout)
					}
				}
				wk.mu.Unlock()
				wk.markReady()
				wk.stop()
			}
		}()

		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) latest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		s.Log.Info("request for latest", "id", id)

		wk, ok := s.getWorker(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		select {
		case <-wk.ready:
		case <-r.Context().Done():
			return
		}
		var res struct {
			Result string `json:"result"`
			Hash   string `json:"hash"`
			Done   bool   `json:"done"`
			Error  string `json:"error"`
		}
		var responseErr error
		func() {
			wk.mu.Lock()
			defer wk.mu.Unlock()
			res.Done = wk.done
			if wk.err != nil {
				res.Error = wk.err.Error()
				return
			}
			if wk.result == nil {
				responseErr = errors.New("unexpected result is blank")
				return
			}
			b, err := wk.result.MarshalJSON()
			if err != nil {
				responseErr = fmt.Errorf("error marshalling result to json: %w", err)
				return
			}
			res.Result = string(b)
			res.Hash, err = wk.result.Sign(s.ShareKey)
			if err != nil {
				responseErr = fmt.Errorf("error signing result: %w", err)
			}
		}()

		defer func() {
			if res.Done {
				s.Lock()
				if s.pool[id] == wk {
					delete(s.pool, id)
				}
				s.Unlock()
			}
		}()

		s.Log.Info("found data", "id", id, "res.Error", res.Error, "res.Done", res.Done)
		if responseErr != nil {
			s.Log.Info("error preparing result", "id", id, "err", responseErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(responseErr.Error()))
			return
		}

		msg, err := json.Marshal(res)
		if err != nil {
			s.Log.Info("error marshalling final response to json", "id", id, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(msg)
	}
}

func (s *Server) cancel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		wk, ok := s.getWorker(id)
		if !ok {
			s.Log.Info("cancel request received; worker does not exist", "id", id)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		wk.mu.Lock()
		if wk.done {
			wk.mu.Unlock()
			s.Log.Info("cancel request received; already done", "id", id)
			w.WriteHeader(http.StatusOK)
			return
		}
		wk.done = true
		if wk.result == nil && wk.err == nil {
			wk.err = errors.New("simulation cancelled")
		}
		wk.mu.Unlock()
		s.Log.Info("cancelling run", "id", id)
		wk.markReady()
		wk.stop()
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) info() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
