package breezebornerefrain

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	buffKey  = "breezeborne-refrain-viper"
	teamKey  = "breezeborne-refrain-team-bonus"
	stackICD = "breezeborne-refrain-stack-icd"
)

type Weapon struct {
	Index  int
	core   *core.Core
	char   *character.CharWrapper
	refine int
	stacks int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	for _, ch := range w.core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(teamKey, -1),
			Amount: func(ai info.AttackInfo) float64 {
				if !ai.AttackTag.IsStellar() {
					return 0
				}
				// The strongest active wielder is an assumption, not a verified game mechanic.
				best := 0.0
				for _, member := range w.core.Player.Chars() {
					other, ok := member.Equip.Weapon.(*Weapon)
					if ok && other.char.StatusDuration(buffKey) > 0 {
						best = max(best, 0.18+0.06*float64(other.refine))
					}
				}
				return best
			},
		})
	}
	return nil
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{core: c, char: char, refine: p.Refine}
	er := make([]float64, attributes.EndStatType)
	er[attributes.ER] = 0.15 + 0.05*float64(p.Refine)
	char.AddStatMod(character.StatMod{
		Base:   modifier.NewBase("breezeborne-refrain-er", -1),
		Amount: func() []float64 { return er },
	})

	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || atk.Info.ActorIndex != char.Index() {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagElementalBurst:
		default:
			return
		}
		if char.StatusDuration(buffKey) > 0 || char.StatusDuration(stackICD) > 0 {
			return
		}
		w.stacks++
		char.AddStatus(stackICD, 2, true)
		if w.stacks == 3 {
			w.stacks = 0
			char.AddStatus(buffKey, 12*60, true)
		}
	}, "breezeborne-refrain-on-hit-"+char.Base.Key.String())
	return w, nil
}
