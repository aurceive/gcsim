package starsealededict

import "testing"

func TestPassiveBonuses(t *testing.T) {
	tests := []struct {
		refine           int
		wantBaseEM       float64
		wantStackEM      float64
		wantStackCritDMG float64
	}{
		{refine: 1, wantBaseEM: 48, wantStackEM: 24, wantStackCritDMG: 0.12},
		{refine: 5, wantBaseEM: 96, wantStackEM: 48, wantStackCritDMG: 0.24},
	}

	for _, tt := range tests {
		baseEM, stackEM, stackCritDMG := passiveBonuses(tt.refine)
		if baseEM != tt.wantBaseEM ||
			stackEM != tt.wantStackEM ||
			stackCritDMG != tt.wantStackCritDMG {
			t.Errorf("passiveBonuses(%d) = (%v, %v, %v), want (%v, %v, %v)",
				tt.refine, baseEM, stackEM, stackCritDMG,
				tt.wantBaseEM, tt.wantStackEM, tt.wantStackCritDMG)
		}
	}
}
