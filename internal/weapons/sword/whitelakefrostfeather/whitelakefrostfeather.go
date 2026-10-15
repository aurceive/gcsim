package whitelakefrostfeather

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

// Elemental Skill hits grant 8%-16% ATK per stack for 8s, up to 3 stacks,
// with independent durations. At 3 stacks, the equipping character's Stellar
// Glimmer reaction DMG gains 50%-110% CRIT DMG; triggering those reactions
// restores 4-6 Energy once every 3.5s. The effects work off-field.
const (
	stackICDKey  = "whitelake-frostfeather-stack-icd"
	energyICDKey = "whitelake-frostfeather-energy-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.WhitelakeFrostfeather, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	atkGain := 0.06 + float64(r)*0.02
	reactionCritDamage := 0.35 + float64(r)*0.15
	energyRestore := 3.5 + float64(r)*0.5

	mATK := make([]float64, attributes.EndStatType)
	mCritDamage := make([]float64, attributes.EndStatType)
	mCritDamage[attributes.CD] = reactionCritDamage
	stackKeys := []string{
		"whitelake-frostfeather-stack-1",
		"whitelake-frostfeather-stack-2",
		"whitelake-frostfeather-stack-3",
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

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("whitelake-frostfeather-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			mATK[attributes.ATKP] = atkGain * float64(stackCount())
			if mATK[attributes.ATKP] == 0 {
				return nil
			}
			return mATK
		},
	})
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("whitelake-frostfeather-stellar-crit", -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if stackCount() < 3 || !isStellarGlimmerDamage(atk.Info.AttackTag) {
				return nil
			}
			return mCritDamage
		},
	})

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || char.StatusIsActive(stackICDKey) {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
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
			char.AddStatus(stackKeys[slot], 8*60, true)
			char.AddStatus(stackICDKey, 6, true)
		}
	}, fmt.Sprintf("whitelake-frostfeather-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnStellarConduct, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || stackCount() < 3 {
			return
		}
		if char.StatusIsActive(energyICDKey) {
			return
		}
		char.AddStatus(energyICDKey, 3.5*60, true)
		char.AddEnergy("whitelake-frostfeather", energyRestore)
	}, fmt.Sprintf("whitelake-frostfeather-energy-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStellarSwirl, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || stackCount() < 3 {
			return
		}
		if atk.Info.ActorIndex == char.Index() && !char.StatusIsActive(energyICDKey) {
			char.AddStatus(energyICDKey, 3.5*60, true)
			char.AddEnergy("whitelake-frostfeather", energyRestore)
		}
	}, fmt.Sprintf("whitelake-frostfeather-swirl-%v", char.Base.Key.String()))

	return w, nil
}

func isStellarGlimmerDamage(tag attacks.AttackTag) bool {
	switch tag {
	case attacks.AttackTagDirectStellarConduct,
		attacks.AttackTagDirectStellarSwirl,
		attacks.AttackTagReactionStellarSwirl:
		return true
	default:
		return false
	}
}
