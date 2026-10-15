package clashofkings

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

const (
	chargeExtensionKey = "clash-of-kings-ca-extension"
	buffKey            = "clash-of-kings"
	buffICDKey         = "clash-of-kings-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.ClashOfKings, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Using an Elemental Skill grants 20%-40% ATK and 100-200 EM for 6s, once
// every 12s. Hitting an opponent with a Charged Attack during the effect
// extends its duration by 6s, up to one extension.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	atkGain, emGain := passiveBonuses(p.Refine)

	bonus := make([]float64, attributes.EndStatType)
	bonus[attributes.ATKP] = atkGain
	bonus[attributes.EM] = emGain

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if char.Index() != c.Player.Active() || char.StatusIsActive(buffICDKey) {
			return
		}

		char.AddStatus(buffICDKey, 12*60, true)
		char.AddStatus(chargeExtensionKey, 12*60, true)
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag(buffKey, 6*60),
			Amount: func() []float64 {
				return bonus
			},
		})
	}, "clash-of-kings-on-skill-"+char.Base.Key.String())

	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || atk.Info.ActorIndex != char.Index() || char.Index() != c.Player.Active() {
			return
		}
		if atk.Info.AttackTag != attacks.AttackTagExtra ||
			!char.StatusIsActive(buffKey) ||
			!char.StatusIsActive(chargeExtensionKey) {
			return
		}

		char.DeleteStatus(chargeExtensionKey)
		char.ExtendStatus(buffKey, 6*60)
	}, "clash-of-kings-on-charge-hit-"+char.Base.Key.String())

	return w, nil
}

func passiveBonuses(refine int) (float64, float64) {
	return 0.15 + float64(refine)*0.05, float64(75 + refine*25)
}
