package disasterandremorse

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
	pathKey        = "disasterandremorse-path"
	unforgivable   = "disasterandremorse-unforgivable"
	irreparable    = "disasterandremorse-irreparable"
	cdKey          = "disasterandremorse-cd"
	naIcdKey       = "disasterandremorse-na-icd"
	skillIcdKey    = "disasterandremorse-skill-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.DisasterAndRemorse, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// After the equipping character uses an Elemental Skill, they gain "Path of Conflict" for 17s
// (CD: 18s), as well as "Unforgivable" (Normal and Charged Attack DMG +40/50/60/70/80%)
// and "Irreparable" (Elemental Skill and Burst DMG +40/50/60/70/80%) for 3s each.
// While Path of Conflict is in effect:
//   - hitting with Normal/Charged Attack extends Irreparable by 1s (ICD: 0.1s)
//   - hitting with Elemental Skill/Burst extends Unforgivable by 1s (ICD: 0.1s)
//
// When Path of Conflict ends or the equipping character leaves the field, both are removed.
// Hexerei: Secret Rite: DMG boosts are increased by 75% (×1.75) when >= 2 Hexerei characters.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// R1=0.40, R2=0.50, R3=0.60, R4=0.70, R5=0.80
	dmg := 0.30 + float64(r)*0.10

	getBonus := func() float64 {
		if c.Player.GetHexereiCount() >= 2 {
			return 1.75
		}
		return 1.0
	}

	mNACA := make([]float64, attributes.EndStatType)
	mSkillBurst := make([]float64, attributes.EndStatType)

	// Unforgivable: NA/CA DMG bonus
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("disasterandremorse-unforgivable-mod", -1),
		Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
			if atk.Info.ActorIndex != char.Index() {
				return nil
			}
			if !char.StatusIsActive(pathKey) || !char.StatusIsActive(unforgivable) {
				return nil
			}
			switch atk.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra:
				mNACA[attributes.DmgP] = dmg * getBonus()
				return mNACA
			}
			return nil
		},
	})

	// Irreparable: Skill/Burst DMG bonus
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("disasterandremorse-irreparable-mod", -1),
		Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
			if atk.Info.ActorIndex != char.Index() {
				return nil
			}
			if !char.StatusIsActive(pathKey) || !char.StatusIsActive(irreparable) {
				return nil
			}
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagElementalBurst:
				mSkillBurst[attributes.DmgP] = dmg * getBonus()
				return mSkillBurst
			}
			return nil
		},
	})

	// On Skill: activate Path of Conflict, Unforgivable and Irreparable
	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		if char.StatusIsActive(cdKey) {
			return
		}
		char.AddStatus(pathKey, 17*60, true)
		char.AddStatus(unforgivable, 3*60, true)
		char.AddStatus(irreparable, 3*60, true)
		char.AddStatus(cdKey, 18*60, true)
	}, fmt.Sprintf("disasterandremorse-skill-%v", char.Base.Key.String()))

	// On damage: extend sub-buffs while Path is active (each with 0.1s ICD = 6 frames)
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		if !char.StatusIsActive(pathKey) {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal, attacks.AttackTagExtra:
			if !char.StatusIsActive(naIcdKey) {
				if char.ExtendStatus(irreparable, 60) {
					char.AddStatus(naIcdKey, 6, false)
				}
			}
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagElementalBurst:
			if !char.StatusIsActive(skillIcdKey) {
				if char.ExtendStatus(unforgivable, 60) {
					char.AddStatus(skillIcdKey, 6, false)
				}
			}
		}
	}, fmt.Sprintf("disasterandremorse-hit-%v", char.Base.Key.String()))

	// On swap-out: remove both sub-buffs
	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev := args[0].(int)
		if prev != char.Index() {
			return
		}
		char.DeleteStatus(unforgivable)
		char.DeleteStatus(irreparable)
	}, fmt.Sprintf("disasterandremorse-swap-%v", char.Base.Key.String()))

	return w, nil
}
