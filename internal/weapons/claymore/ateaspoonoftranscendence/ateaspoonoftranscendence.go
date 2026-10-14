package ateaspoonoftranscendence

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// Grants 28%-56% ATK. Charged Attack hits grant a stack for 5s, once every
// 0.2s, up to 3 stacks. Each stack increases Stellar Conduct and Stellar Swirl
// reaction DMG by 16%-32%.
const (
	stackICDKey = "a-teaspoon-of-transcendence-stack-icd"
	stellarKey  = "a-teaspoon-of-transcendence-stellar"
)

func init() {
	core.RegisterWeaponFunc(keys.ATeaspoonOfTranscendence, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	atkBonus, stellarPerStack := passiveBonuses(p.Refine)

	atk := make([]float64, attributes.EndStatType)
	atk[attributes.ATKP] = atkBonus
	char.AddStatMod(character.StatMod{
		Base:   modifier.NewBase("a-teaspoon-of-transcendence-atk", -1),
		Amount: func() []float64 { return atk },
	})

	onChargedAttackHit := func(args ...any) {
		if len(args) < 2 {
			return
		}
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return
		}
		atkEvent, ok := args[1].(*info.AttackEvent)
		if !ok ||
			atkEvent.Info.ActorIndex != char.Index() ||
			atkEvent.Info.AttackTag != attacks.AttackTagExtra ||
			char.StatusIsActive(stackICDKey) {
			return
		}

		if !char.StatusIsActive(stellarKey) {
			w.stacks = 0
		}
		w.stacks = min(w.stacks+1, 3)
		char.AddStatus(stackICDKey, 12, true)
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag(stellarKey, 5*60),
			Amount: func(ai info.AttackInfo) float64 {
				if !isStellarReaction(ai.AttackTag) {
					return 0
				}
				return stellarPerStack * float64(w.stacks)
			},
		})
	}
	c.Events.Subscribe(event.OnEnemyHit, onChargedAttackHit, "a-teaspoon-of-transcendence-"+char.Base.Key.String())

	return w, nil
}

func passiveBonuses(refine int) (float64, float64) {
	return 0.21 + 0.07*float64(refine), 0.12 + 0.04*float64(refine)
}

func isStellarReaction(tag attacks.AttackTag) bool {
	switch tag {
	case attacks.AttackTagDirectStellarConduct,
		attacks.AttackTagDirectStellarSwirl,
		attacks.AttackTagReactionStellarSwirl:
		return true
	default:
		return false
	}
}
