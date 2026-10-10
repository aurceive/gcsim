package newbough

import "testing"

func TestStackBonuses(t *testing.T) {
	tests := []struct {
		name              string
		refine            int
		stellarGlimmer    bool
		wantATK, wantEM   float64
		wantReactionBonus float64
	}{
		{name: "standard R1", refine: 1, wantATK: 0.04, wantEM: 20},
		{name: "standard R5", refine: 5, wantATK: 0.08, wantEM: 40},
		{name: "stellar glimmer R1", refine: 1, stellarGlimmer: true, wantATK: 0.06, wantReactionBonus: 0.08},
		{name: "stellar glimmer R5", refine: 5, stellarGlimmer: true, wantATK: 0.12, wantReactionBonus: 0.16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atk, em, reaction := stackBonuses(tt.refine, tt.stellarGlimmer)
			if atk != tt.wantATK || em != tt.wantEM || reaction != tt.wantReactionBonus {
				t.Errorf("stackBonuses(%d, %t) = (%v, %v, %v), want (%v, %v, %v)",
					tt.refine, tt.stellarGlimmer, atk, em, reaction,
					tt.wantATK, tt.wantEM, tt.wantReactionBonus)
			}
		})
	}
}
