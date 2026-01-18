package blackmarrowlantern

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.BlackmarrowLantern, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const reactModKey = "blackmarrowlantern-react"

// Blackmarrow Lantern (i_n14433):
//   - Bloom DMG +48/60/72/84/96%
//   - Lunar-Bloom DMG +12/15/18/21/24%
//   - Moonsign: Ascendant Gleam: Lunar-Bloom DMG +additional 12/15/18/21/24%
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)

	bloomBonus := 0.36 + 0.12*r // 0.48/0.60/0.72/0.84/0.96
	lunarBonus := 0.09 + 0.03*r // 0.12/0.15/0.18/0.21/0.24

	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase(reactModKey, -1),
		Amount: func(ai info.AttackInfo) (float64, bool) {
			switch ai.AttackTag {
			case attacks.AttackTagBloom:
				return bloomBonus, false
			case attacks.AttackTagDirectLunarBloom:
				return lunarBonus * getBonus(c), false
			default:
				return 0, false
			}
		},
	})

	return w, nil
}

func getBonus(c *core.Core) float64 {
	if c.Player.GetMoonsignCount() < 2 {
		return 1.0
	}
	return 2.0
}
