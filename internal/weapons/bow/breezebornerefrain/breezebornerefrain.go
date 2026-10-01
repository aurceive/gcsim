package breezebornerefrain

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

const (
	stackICDKey = "breezeborne-refrain-stack-icd"
	buffKey     = "breezeborne-refrain-buff"
)

func init() {
	core.RegisterWeaponFunc(keys.BreezeborneRefrain, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	energyRecharge := 0.15 + 0.05*r
	stellarBonus := 0.18 + 0.06*r
	stacks := 0

	er := make([]float64, attributes.EndStatType)
	er[attributes.ER] = energyRecharge
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("breezeborne-refrain-er", -1),
		AffectedStat: attributes.ER,
		Amount: func() []float64 {
			return er
		},
	})

	for _, partyChar := range c.Player.Chars() {
		partyChar.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase("breezeborne-refrain-stellar", -1),
			Amount: func(ai info.AttackInfo) float64 {
				if !char.StatusIsActive(buffKey) || !attacks.AttackTagIsStellar(ai.AttackTag) {
					return 0
				}
				return stellarBonus
			},
		})
	}

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() ||
			char.StatusIsActive(buffKey) ||
			char.StatusIsActive(stackICDKey) {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagElementalBurst:
		default:
			return
		}

		char.AddStatus(stackICDKey, 2, true)
		stacks++
		if stacks < 3 {
			return
		}
		stacks = 0
		char.AddStatus(buffKey, 12*60, true)
	}, fmt.Sprintf("breezeborne-refrain-%v", char.Base.Key.String()))

	return w, nil
}
