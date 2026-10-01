package silverlight

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const stackGainICDKey = "silver-light-stack-icd"

func init() {
	core.RegisterWeaponFunc(keys.SilverLight, NewWeapon)
}

type Weapon struct {
	Index        int
	stackCounter int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	em := make([]float64, attributes.EndStatType)
	em[attributes.EM] = 39 + 13*float64(p.Refine)

	onSkill := func(args ...any) {
		if char.Index() != c.Player.Active() || char.StatusIsActive(stackGainICDKey) {
			return
		}

		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag(fmt.Sprintf("silver-light-em-%v", w.stackCounter+1), 12*60),
			Amount: func() []float64 {
				return em
			},
		})
		char.AddStatus(stackGainICDKey, 12, true)
		w.stackCounter = (w.stackCounter + 1) % 2
	}

	c.Events.Subscribe(event.OnSkill, onSkill, "silver-light-on-skill-"+char.Base.Key.String())
	return w, nil
}
