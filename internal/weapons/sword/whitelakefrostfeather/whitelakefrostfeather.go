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

const (
	stackICDKey    = "whitelake-frostfeather-stack-icd"
	energyICDKey   = "whitelake-frostfeather-energy-icd"
	stellarCritKey = "whitelake-frostfeather-stellar-crit"
)

func init() {
	core.RegisterWeaponFunc(keys.WhitelakeFrostfeather, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Lake-Hued Lament: when the equipping character's Elemental Skill hits an enemy,
// ATK is increased by 8%/10%/12%/14%/16% for 8s. At 3 stacks, Stellar Glimmer
// reaction CRIT DMG is increased and the weapon restores energy.
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
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("whitelake-frostfeather-stellar-crit", -1),
		AffectedStat: attributes.CD,
		Amount: func() []float64 {
			if !char.StatusIsActive(stellarCritKey) {
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
		char.AddStatus(stellarCritKey, 1, true)
		if char.StatusIsActive(energyICDKey) {
			return
		}
		char.AddStatus(energyICDKey, 3.5*60, true)
		char.AddEnergy("whitelake-frostfeather", energyRestore)
	}, fmt.Sprintf("whitelake-frostfeather-energy-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStellarSwirl, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if stackCount() < 3 {
			return
		}
		char.AddStatus(stellarCritKey, 1, true)
		if atk.Info.ActorIndex == char.Index() && !char.StatusIsActive(energyICDKey) {
			char.AddStatus(energyICDKey, 3.5*60, true)
			char.AddEnergy("whitelake-frostfeather", energyRestore)
		}
	}, fmt.Sprintf("whitelake-frostfeather-swirl-%v", char.Base.Key.String()))

	return w, nil
}
