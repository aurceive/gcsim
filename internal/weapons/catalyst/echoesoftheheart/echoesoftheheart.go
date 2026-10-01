package echoesoftheheart

import (
	"fmt"

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
	core.RegisterWeaponFunc(keys.EchoesOfTheHeart, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// When the equipping character's Elemental Skill hits an opponent, they gain
// a buff that increases ATK by 16%/20%/24%/28%/32% for 8s. Also increases the
// reaction damage bonus for Stellar Glimmer-related reactions.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	em := make([]float64, attributes.EndStatType)
	em[attributes.EM] = 30 + float64(r)*15
	const (
		emKey    = "echoes-of-the-heart-em"
		reactKey = "echoes-of-the-heart-react-status"
	)
	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		char.AddStatus(emKey, 12*60, true)
		char.AddStatus(reactKey, 12*60, true)
	}
	for e := event.ReactionEventStartDelim + 1; e < event.ReactionEventEndDelim; e++ {
		c.Events.Subscribe(e, onReaction, fmt.Sprintf("echoes-of-the-heart-%v-%d", char.Base.Key.String(), e))
	}

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(emKey, -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			if char.StatusIsActive(emKey) {
				return em
			}
			return nil
		},
	})
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("echoes-of-the-heart-react", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if !char.StatusIsActive(reactKey) {
				return 0
			}
			switch ai.AttackTag {
			case attacks.AttackTagDirectStellarConduct,
				attacks.AttackTagDirectStellarSwirl,
				attacks.AttackTagReactionStellarSwirl:
				return 0.12 + float64(r)*0.04
			default:
				return 0
			}
		},
	})
	return w, nil
}
