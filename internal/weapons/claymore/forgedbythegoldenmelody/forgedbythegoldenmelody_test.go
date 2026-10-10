package forgedbythegoldenmelody

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
)

func TestIsStellarDamageTag(t *testing.T) {
	tests := []struct {
		tag  attacks.AttackTag
		want bool
	}{
		{tag: attacks.AttackTagDirectStellarConduct, want: true},
		{tag: attacks.AttackTagDirectStellarSwirl, want: true},
		{tag: attacks.AttackTagReactionStellarSwirl, want: true},
		{tag: attacks.AttackTagSwirlCryo, want: false},
	}

	for _, tt := range tests {
		if got := isStellarDamageTag(tt.tag); got != tt.want {
			t.Errorf("isStellarDamageTag(%v) = %t, want %t", tt.tag, got, tt.want)
		}
	}
}
