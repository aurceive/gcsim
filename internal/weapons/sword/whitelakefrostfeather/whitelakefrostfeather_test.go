package whitelakefrostfeather

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func TestWeaponCritRateScaling(t *testing.T) {
	for _, tt := range []struct {
		level int
		want  float64
	}{
		{level: 1, want: 0.048},
		{level: 90, want: 0.2205},
	} {
		stats, err := character.WeaponBaseStat(info.WeaponProfile{
			Level:    tt.level,
			MaxLevel: 90,
		}, base)
		if err != nil {
			t.Fatalf("WeaponBaseStat(level %d): %v", tt.level, err)
		}
		if got := stats[attributes.CR]; math.Abs(got-tt.want) > 0.001 {
			t.Errorf("weapon CR at level %d = %.4f, want %.4f", tt.level, got, tt.want)
		}
	}
}

func TestIsStellarGlimmerDamage(t *testing.T) {
	for _, tt := range []struct {
		tag  attacks.AttackTag
		want bool
	}{
		{tag: attacks.AttackTagDirectStellarConduct, want: true},
		{tag: attacks.AttackTagDirectStellarSwirl, want: true},
		{tag: attacks.AttackTagReactionStellarSwirl, want: true},
		{tag: attacks.AttackTagNormal, want: false},
		{tag: attacks.AttackTagElementalArt, want: false},
	} {
		if got := isStellarGlimmerDamage(tt.tag); got != tt.want {
			t.Errorf("isStellarGlimmerDamage(%v) = %t, want %t", tt.tag, got, tt.want)
		}
	}
}
