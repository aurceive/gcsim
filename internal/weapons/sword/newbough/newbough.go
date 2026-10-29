package newbough

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

// Hitting an opponent within 12s after using an Elemental Skill grants a 6s
// Verdant stack, once per second, up to 3 stacks. Normally, each stack grants
// 4%-8% ATK and 20-40 EM. While Radiance: Stellar Glimmer (Stellar Conduct or
// Stellar Swirl) is active, each instead grants 6%-12% ATK and 8%-16% Stellar
// Glimmer reaction DMG. Hits can trigger while off-field.
const (
	skillWindowKey = "new-bough-skill-window"
	stackICDKey    = "new-bough-stack-icd"
)

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	stackKeys := []string{
		"new-bough-verdant-1",
		"new-bough-verdant-2",
		"new-bough-verdant-3",
	}
	stackCount := func() int {
		count := 0
		for _, key := range stackKeys {
			if char.StatusDuration(key) > 0 {
				count++
			}
		}
		return count
	}

	atk := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("new-bough-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			atkPerStack, _, _ := stackBonuses(p.Refine, stellarGlimmerActive(char))
			atk[attributes.ATKP] = atkPerStack * float64(stackCount())
			if atk[attributes.ATKP] == 0 {
				return nil
			}
			return atk
		},
	})
	em := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("new-bough-em", -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			_, emPerStack, _ := stackBonuses(p.Refine, stellarGlimmerActive(char))
			if emPerStack == 0 {
				return nil
			}
			em[attributes.EM] = emPerStack * float64(stackCount())
			if em[attributes.EM] == 0 {
				return nil
			}
			return em
		},
	})
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("new-bough-stellar", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if stackCount() == 0 {
				return 0
			}
			_, _, reactionPerStack := stackBonuses(p.Refine, stellarGlimmerActive(char))
			switch ai.AttackTag {
			case attacks.AttackTagDirectStellarConduct,
				attacks.AttackTagDirectStellarSwirl,
				attacks.AttackTagReactionStellarSwirl:
				return reactionPerStack * float64(stackCount())
			default:
				return 0
			}
		},
	})

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() == char.Index() {
			char.AddStatus(skillWindowKey, 12*60, true)
		}
	}, fmt.Sprintf("new-bough-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atkEvent := args[1].(*info.AttackEvent)
		if atkEvent.Info.ActorIndex != char.Index() ||
			char.StatusDuration(skillWindowKey) == 0 ||
			char.StatusDuration(stackICDKey) > 0 {
			return
		}

		slot := -1
		for i, key := range stackKeys {
			if char.StatusDuration(key) == 0 {
				slot = i
				break
			}
		}
		if slot == -1 {
			slot = 0
			for i := 1; i < len(stackKeys); i++ {
				if char.StatusExpiry(stackKeys[i]) < char.StatusExpiry(stackKeys[slot]) {
					slot = i
				}
			}
		}
		char.AddStatus(stackKeys[slot], 6*60, true)
		char.AddStatus(stackICDKey, 60, true)
	}, fmt.Sprintf("new-bough-hit-%v", char.Base.Key.String()))

	return w, nil
}

func stellarGlimmerActive(char *character.CharWrapper) bool {
	return char.StatusIsActive(reactable.PolestarFieldKey) ||
		char.StatusIsActive(reactable.SswKey)
}

func stackBonuses(refine int, stellarGlimmer bool) (float64, float64, float64) {
	atk := 0.03 + 0.01*float64(refine)
	em := float64(15 + 5*refine)
	reaction := 0.0
	if stellarGlimmer {
		atk = 0.045 + 0.015*float64(refine)
		em = 0
		reaction = 0.06 + 0.02*float64(refine)
	}
	return atk, em, reaction
}
