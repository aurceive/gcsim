package nocturnescurtaincall

import (
	"fmt"

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

const (
	// Buff is refreshed on every trigger.
	buffKey = "nocturnes-curtain-call-buff"
	// Energy restore is limited by an 18s ICD.
	energyIcdKey = "nocturnes-curtain-call-energy-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.NocturnesCurtainCall, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Ballad of the Crossroads
//
// HP% increased.
// When the character triggers a Lunar reaction or deals Lunar Reaction DMG:
// - gain a 12s buff (refreshable with no cooldown): additional HP% and Lunar Reaction DMG CRIT DMG bonus
// - restore Energy (can only happen once every 18s)
// Can trigger off-field.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	hpBonus := []float64{0.10, 0.12, 0.14, 0.16, 0.18}[r-1]
	energyRestore := float64(13 + r) // 14..18
	buffHPBonus := []float64{0.14, 0.16, 0.18, 0.20, 0.22}[r-1]
	lunarCritDmgBonus := []float64{0.60, 0.80, 1.00, 1.20, 1.40}[r-1]

	permHP := make([]float64, attributes.EndStatType)
	permHP[attributes.HPP] = hpBonus
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("nocturnes-curtain-call-hp%", -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			return permHP, true
		},
	})

	buffHP := make([]float64, attributes.EndStatType)
	buffHP[attributes.HPP] = buffHPBonus

	buffCD := make([]float64, attributes.EndStatType)
	buffCD[attributes.CD] = lunarCritDmgBonus

	proc := func() {
		// Buff can be refreshed with no cooldown.
		char.AddStatus(buffKey, 12*60, true)

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey+"-hp%", 12*60),
			AffectedStat: attributes.HPP,
			Amount: func() ([]float64, bool) {
				return buffHP, true
			},
		})

		// Applies to direct lunar reaction damage (AttackMods skip non-direct reaction damage).
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(buffKey+"-cd", 12*60),
			Amount: func(atk *info.AttackEvent, t info.Target) ([]float64, bool) {
				if !attacks.AttackTagIsLunar(atk.Info.AttackTag) {
					return nil, false
				}
				return buffCD, true
			},
		})

		// Energy restore can only happen once every 18s.
		if char.StatusIsActive(energyIcdKey) {
			return
		}
		char.AddStatus(energyIcdKey, 18*60, true)
		char.AddEnergy("nocturnes-curtain-call", energyRestore)
	}

	// Triggered a Lunar reaction (OnLunarBloom does not necessarily imply lunar-tagged damage).
	triggerReaction := func(args ...any) bool {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return false
		}
		proc()
		return false
	}

	c.Events.Subscribe(event.OnLunarCharged, triggerReaction, fmt.Sprintf("nocturnes-curtain-call-lc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnLunarCrystallize, triggerReaction, fmt.Sprintf("nocturnes-curtain-call-lcr-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnLunarBloom, triggerReaction, fmt.Sprintf("nocturnes-curtain-call-lb-%v", char.Base.Key.String()))

	// Deals Lunar Reaction DMG.
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return false
		}
		if !attacks.AttackTagIsLunar(atk.Info.AttackTag) {
			return false
		}
		proc()
		return false
	}, fmt.Sprintf("nocturnes-curtain-call-lunar-dmg-%v", char.Base.Key.String()))

	// Apply CRIT DMG bonus to non-direct lunar reaction damage contributions (LC/LCr).
	c.Events.Subscribe(event.OnLunarReactionAttack, func(args ...any) bool {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return false
		}
		if !char.StatusIsActive(buffKey) {
			return false
		}
		atk.Snapshot.Stats[attributes.CD] += lunarCritDmgBonus
		return false
	}, fmt.Sprintf("nocturnes-curtain-call-lunar-react-atk-%v", char.Base.Key.String()))

	return w, nil
}
