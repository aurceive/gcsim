package covenantoffrostandsnow

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.CovenantOfFrostAndSnow, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Using an Elemental Skill grants the equipping character 120-240 EM for 12s.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	emBonus := make([]float64, attributes.EndStatType)
	emBonus[attributes.EM] = float64(90 + 30*p.Refine)

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("covenant-of-frost-and-snow-em", 12*60),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return emBonus
			},
		})
	}, "covenant-of-frost-and-snow-on-skill-"+char.Base.Key.String())

	return &Weapon{}, nil
}
