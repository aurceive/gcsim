package echoesoftheheart

import "testing"

func TestPassiveBonuses(t *testing.T) {
	tests := []struct {
		refine       int
		wantEM       float64
		wantReaction float64
	}{
		{refine: 1, wantEM: 60, wantReaction: 0.16},
		{refine: 5, wantEM: 120, wantReaction: 0.32},
	}

	for _, tt := range tests {
		em, reaction := passiveBonuses(tt.refine)
		if em != tt.wantEM || reaction != tt.wantReaction {
			t.Errorf("passiveBonuses(%d) = (%v, %v), want (%v, %v)", tt.refine, em, reaction, tt.wantEM, tt.wantReaction)
		}
	}
}
