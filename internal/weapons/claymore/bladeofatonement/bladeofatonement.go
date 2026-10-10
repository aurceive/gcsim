package bladeofatonement

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

func init() {
	core.RegisterWeaponFunc(keys.BladeOfAtonement, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Triggering an Elemental Reaction grants 64-128 EM for 12s; a Stellar Glimmer
// reaction also grants 16%-32% ATK for 12s. Both effects refresh, not stack,
// and can trigger while the equipping character is off-field.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	emBonus := make([]float64, attributes.EndStatType)
	emBonus[attributes.EM] = float64(48 + 16*p.Refine)

	atkBonus := make([]float64, attributes.EndStatType)
	atkBonus[attributes.ATKP] = 0.12 + float64(p.Refine)*0.04

	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("blade-of-atonement-em", 12*60),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return emBonus
			},
		})
	}

	onStellar := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("blade-of-atonement-atk", 12*60),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return atkBonus
			},
		})
	}

	for evt := event.ReactionEventStartDelim + 1; evt < event.ReactionEventEndDelim; evt++ {
		c.Events.Subscribe(evt, onReaction, fmt.Sprintf("blade-of-atonement-on-reaction-%s", char.Base.Key.String()))
	}

	c.Events.Subscribe(event.OnStellarConduct, onStellar, "blade-of-atonement-on-stellar-conduct-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnStellarSwirl, onStellar, "blade-of-atonement-on-stellar-swirl-"+char.Base.Key.String())

	return &Weapon{}, nil
}
