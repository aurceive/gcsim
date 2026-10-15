package clashofkings

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/shortcut"
)

func TestPassiveBonuses(t *testing.T) {
	for _, tt := range []struct {
		refine  int
		wantATK float64
		wantEM  float64
	}{
		{refine: 1, wantATK: 0.20, wantEM: 100},
		{refine: 2, wantATK: 0.25, wantEM: 125},
		{refine: 3, wantATK: 0.30, wantEM: 150},
		{refine: 4, wantATK: 0.35, wantEM: 175},
		{refine: 5, wantATK: 0.40, wantEM: 200},
	} {
		atk, em := passiveBonuses(tt.refine)
		if math.Abs(atk-tt.wantATK) > 1e-9 || em != tt.wantEM {
			t.Errorf("passiveBonuses(%d) = (%v, %v), want (%v, %v)",
				tt.refine, atk, em, tt.wantATK, tt.wantEM)
		}
	}
}

func TestWeaponKeyAndData(t *testing.T) {
	var key keys.Weapon
	if err := json.Unmarshal([]byte(`"clashofkings"`), &key); err != nil {
		t.Fatalf("unmarshal weapon key: %v", err)
	}
	if key != keys.ClashOfKings {
		t.Fatalf("weapon key = %v, want %v", key, keys.ClashOfKings)
	}
	for _, alias := range []string{"clashofkings", "Игры королей"} {
		if got := shortcut.WeaponNameToKey[alias]; got != keys.ClashOfKings {
			t.Errorf("weapon alias %q = %v, want %v", alias, got, keys.ClashOfKings)
		}
	}

	data := (&Weapon{}).Data()
	if data.GetId() != 14435 || data.GetKey() != "clashofkings" {
		t.Errorf("weapon data = (id %d, key %q), want (id 14435, key %q)",
			data.GetId(), data.GetKey(), "clashofkings")
	}
}
