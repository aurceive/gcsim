package emberwell

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
	core.RegisterWeaponFunc(keys.Emberwell, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	atkBuff := make([]float64, attributes.EndStatType)
	atkBuff[attributes.ATKP] = 0.12 + float64(p.Refine)*0.04
	stellarBuff := 0.12 + float64(p.Refine)*0.04

	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}

		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag("emberwell-atk", 12*60),
			Amount: func() []float64 {
				return atkBuff
			},
		})
	}

	onStellar := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}

		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag("emberwell-stellar", 12*60),
			Amount: func(ai info.AttackInfo) float64 {
				switch ai.AttackTag {
				case attacks.AttackTagDirectStellarConduct,
					attacks.AttackTagDirectStellarSwirl,
					attacks.AttackTagReactionStellarSwirl:
					return stellarBuff
				default:
					return 0
				}
			},
		})
	}

	for evt := event.ReactionEventStartDelim + 1; evt < event.ReactionEventEndDelim; evt++ {
		c.Events.Subscribe(evt, onReaction, "emberwell-on-reaction-"+char.Base.Key.String())
	}
	c.Events.Subscribe(event.OnStellarConduct, onStellar, "emberwell-stellar-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnStellarSwirl, onStellar, "emberwell-stellar-swirl-"+char.Base.Key.String())

	return w, nil
}
