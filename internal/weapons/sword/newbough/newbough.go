package newbough

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
	skillWindowKey = "new-bough-skill-window"
	stackICDKey    = "new-bough-stack-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.NewBough, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	atkPerStack := 0.045 + 0.015*r
	reactionPerStack := 0.06 + 0.02*r
	stackKeys := []string{
		"new-bough-verdant-1",
		"new-bough-verdant-2",
		"new-bough-verdant-3",
	}
	stackCount := func() int {
		count := 0
		for _, key := range stackKeys {
			if char.StatusIsActive(key) {
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
			atk[attributes.ATKP] = atkPerStack * float64(stackCount())
			if atk[attributes.ATKP] == 0 {
				return nil
			}
			return atk
		},
	})
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("new-bough-stellar", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if stackCount() == 0 {
				return 0
			}
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
			!char.StatusIsActive(skillWindowKey) ||
			char.StatusIsActive(stackICDKey) {
			return
		}

		slot := -1
		for i, key := range stackKeys {
			if !char.StatusIsActive(key) {
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
