package kagura

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
)

func TestKaguraDamageBonus(t *testing.T) {
	tests := []struct {
		refine int
		want   float64
	}{
		{refine: 1, want: 0.12},
		{refine: 5, want: 0.24},
	}

	for _, tt := range tests {
		if got := kaguraDamageBonus(tt.refine); got != tt.want {
			t.Errorf("kaguraDamageBonus(%d) = %v, want %v", tt.refine, got, tt.want)
		}
	}
}

func TestStellarConductBonus(t *testing.T) {
	tests := []struct {
		name   string
		tag    attacks.AttackTag
		stacks int
		want   float64
	}{
		{name: "one stack", tag: attacks.AttackTagDirectStellarConduct, stacks: 1, want: 0.12},
		{name: "three stacks", tag: attacks.AttackTagDirectStellarConduct, stacks: 3, want: 0.36},
		{name: "stellar swirl", tag: attacks.AttackTagDirectStellarSwirl, stacks: 3},
		{name: "non-stellar hit", tag: attacks.AttackTagElementalArt, stacks: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stellarConductBonus(tt.tag, tt.stacks, 0.12); got != tt.want {
				t.Errorf("stellarConductBonus(%v, %d, 0.12) = %v, want %v",
					tt.tag, tt.stacks, got, tt.want)
			}
		})
	}
}
