package clashofkings

import "testing"

func TestPassiveBonuses(t *testing.T) {
	tests := []struct {
		refine  int
		wantATK float64
		wantEM  float64
	}{
		{refine: 1, wantATK: 0.20, wantEM: 100},
		{refine: 5, wantATK: 0.40, wantEM: 300},
	}

	for _, tt := range tests {
		atk, em := passiveBonuses(tt.refine)
		if atk != tt.wantATK || em != tt.wantEM {
			t.Errorf("passiveBonuses(%d) = (%v, %v), want (%v, %v)", tt.refine, atk, em, tt.wantATK, tt.wantEM)
		}
	}
}
