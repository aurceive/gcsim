package clashofkings

import (
	"math"
	"testing"
)

func TestPassiveBonuses(t *testing.T) {
	tests := []struct {
		refine  int
		wantATK float64
		wantEM  float64
	}{
		{refine: 1, wantATK: 0.20, wantEM: 100},
		{refine: 2, wantATK: 0.25, wantEM: 125},
		{refine: 3, wantATK: 0.30, wantEM: 150},
		{refine: 4, wantATK: 0.35, wantEM: 175},
		{refine: 5, wantATK: 0.40, wantEM: 200},
	}

	for _, tt := range tests {
		atk, em := passiveBonuses(tt.refine)
		if math.Abs(atk-tt.wantATK) > 1e-9 || em != tt.wantEM {
			t.Errorf("passiveBonuses(%d) = (%v, %v), want (%v, %v)", tt.refine, atk, em, tt.wantATK, tt.wantEM)
		}
	}
}
