package ateaspoonoftranscendence

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
)

func TestPassiveBonuses(t *testing.T) {
	tests := []struct {
		refine         int
		wantATK        float64
		wantStellarDMG float64
	}{
		{refine: 1, wantATK: 0.28, wantStellarDMG: 0.16},
		{refine: 5, wantATK: 0.56, wantStellarDMG: 0.32},
	}

	for _, tt := range tests {
		atk, stellar := passiveBonuses(tt.refine)
		if atk != tt.wantATK || stellar != tt.wantStellarDMG {
			t.Errorf("passiveBonuses(%d) = (%v, %v), want (%v, %v)",
				tt.refine, atk, stellar, tt.wantATK, tt.wantStellarDMG)
		}
	}
}

func TestIsStellarReaction(t *testing.T) {
	for _, tt := range []struct {
		tag  attacks.AttackTag
		want bool
	}{
		{tag: attacks.AttackTagDirectStellarConduct, want: true},
		{tag: attacks.AttackTagDirectStellarSwirl, want: true},
		{tag: attacks.AttackTagReactionStellarSwirl, want: true},
		{tag: attacks.AttackTagSwirlCryo, want: false},
	} {
		if got := isStellarReaction(tt.tag); got != tt.want {
			t.Errorf("isStellarReaction(%v) = %t, want %t", tt.tag, got, tt.want)
		}
	}
}
