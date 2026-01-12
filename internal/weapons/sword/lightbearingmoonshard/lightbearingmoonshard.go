package lightbearingmoonshard

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.LightbearingMoonshard, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	defKey = "lightbearingmoonshard-def"
	lcrKey = "lightbearingmoonshard-lcr"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	defBonus := 0.15 + 0.05*float64(r) // 20/25/30/35/40%
	lcrBonus := 0.48 + 0.16*float64(r) // 64/80/96/112/128%
	lcrDur := 5 * 60                   // 5s

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(defKey, -1),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return []float64{defBonus}, true
		},
	})

	c.Events.Subscribe(event.OnSkill, func(args ...any) bool {
		if c.Player.Active() != char.Index() {
			return false
		}

		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(lcrKey, lcrDur),
			Amount: func(atk info.AttackInfo) (float64, bool) {
				switch atk.AttackTag {
				case attacks.AttackTagReactionLunarCrystallize, attacks.AttackTagDirectLunarCrystallize:
					return lcrBonus, false
				default:
					return 0, false
				}
			},
		})

		return false
	}, "lightbearingmoonshard-skill")

	return w, nil
}
