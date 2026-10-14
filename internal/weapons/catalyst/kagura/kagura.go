package kagura

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
	core.RegisterWeaponFunc(keys.KagurasVerity, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// Each stack increases Elemental Skill and Stellar-Conduct DMG by 12%-24%
	// for 24s. At 3 stacks, the wielder also gains 12%-24% All Elemental DMG.
	w := &Weapon{}
	stacks := 0
	key := fmt.Sprintf("kaguradance-%v", char.Base.Key.String())
	dmg := kaguraDamageBonus(p.Refine)
	val := make([]float64, attributes.EndStatType)
	const stackKey = "kaguras-verity-stacks"
	stackDuration := 24 * 60

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		if !char.StatusIsActive(stackKey) {
			// reset stacks back to 0
			stacks = 0
		}
		char.AddStatus(stackKey, stackDuration, true)
		if stacks < 3 {
			stacks++
		}
		// bonus ele damage if 3 stacks
		if stacks == 3 {
			val[attributes.PyroP] = dmg
			val[attributes.HydroP] = dmg
			val[attributes.CryoP] = dmg
			val[attributes.ElectroP] = dmg
			val[attributes.AnemoP] = dmg
			val[attributes.GeoP] = dmg
			val[attributes.DendroP] = dmg
		} else {
			// clean stacks ele dmg% otherwise
			val[attributes.PyroP] = 0
			val[attributes.HydroP] = 0
			val[attributes.CryoP] = 0
			val[attributes.ElectroP] = 0
			val[attributes.AnemoP] = 0
			val[attributes.GeoP] = 0
			val[attributes.DendroP] = 0
		}
		// add mod for duration, override last
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("kaguras-verity", stackDuration),
			Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
				if atk.Info.ActorIndex != char.Index() {
					return nil
				}
				if atk.Info.AttackTag == attacks.AttackTagElementalArt || atk.Info.AttackTag == attacks.AttackTagElementalArtHold {
					val[attributes.DmgP] = dmg * float64(stacks)
				} else {
					val[attributes.DmgP] = 0
				}
				return val
			},
		})
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag("kaguras-verity-stellar-conduct", stackDuration),
			Amount: func(ai info.AttackInfo) float64 {
				return stellarConductBonus(ai.AttackTag, stacks, dmg)
			},
		})
	}, key)

	return w, nil
}

func kaguraDamageBonus(refine int) float64 {
	return 0.09 + 0.03*float64(refine)
}

func stellarConductBonus(tag attacks.AttackTag, stacks int, bonusPerStack float64) float64 {
	if tag != attacks.AttackTagDirectStellarConduct {
		return 0
	}
	return bonusPerStack * float64(stacks)
}
