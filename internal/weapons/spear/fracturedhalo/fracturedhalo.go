package fracturedhalo

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.FracturedHalo, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	atkModKey     = "fracturedhalo-atk"
	edictStatus   = "fracturedhalo-edict"
	teamFlagKey   = "fracturedhalo-team-installed"
	teamModKey    = "fracturedhalo-team-react"
	refineTagKey  = "fracturedhalo-refine"
	atkBuffFrames = 20 * 60
)

// Fractured Halo:
//   - After the equipping character uses Elemental Skill or Burst, gain an ATK% buff for 20s.
//     (R1..R5: +24/+30/+36/+42/+48% ATK)
//   - If the equipping character creates a Shield while this ATK% buff is active, gain Electrifying Edict for 20s.
//   - While Electrifying Edict is active on any holder, all party members gain a non-stacking Lunar-Charged DMG bonus
//     (additive term in gcsim's Lunar-Charged formula): +40/+50/+60/+70/+80%.
//     If multiple copies are active, the best (highest) bonus is used.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// Store refine so the shared team mod can compute the non-stacking best bonus.
	char.SetTag(refineTagKey, r)

	atkBonus := 0.18 + 0.06*float64(r) // 0.24/0.30/0.36/0.42/0.48
	mAtk := make([]float64, attributes.EndStatType)
	mAtk[attributes.ATKP] = atkBonus

	applyAtkBuff := func() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(atkModKey, atkBuffFrames),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return mAtk, true
			},
		})
	}

	c.Events.Subscribe(event.OnSkill, func(args ...any) bool {
		if c.Player.Active() != char.Index() {
			return false
		}
		applyAtkBuff()
		return false
	}, fmt.Sprintf("fracturedhalo-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnBurst, func(args ...any) bool {
		if c.Player.Active() != char.Index() {
			return false
		}
		applyAtkBuff()
		return false
	}, fmt.Sprintf("fracturedhalo-burst-%v", char.Base.Key.String()))

	// In-game behavior (as verified): "creating a shield" for this weapon means triggering
	// (Lunar) Crystallize (i.e., closing the reaction), not picking up the shard.
	// So we hook the reaction events and require the holder to be the reaction trigger.
	triggerEdictOnReaction := func(args ...any) bool {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex != char.Index() {
			return false
		}
		if !char.StatModIsActive(atkModKey) {
			return false
		}
		char.AddStatus(edictStatus, atkBuffFrames, true)
		return false
	}

	c.Events.Subscribe(event.OnCrystallizeHydro, triggerEdictOnReaction, fmt.Sprintf("fracturedhalo-crystallize-hydro-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnCrystallizeCryo, triggerEdictOnReaction, fmt.Sprintf("fracturedhalo-crystallize-cryo-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnCrystallizeElectro, triggerEdictOnReaction, fmt.Sprintf("fracturedhalo-crystallize-electro-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnCrystallizePyro, triggerEdictOnReaction, fmt.Sprintf("fracturedhalo-crystallize-pyro-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnLunarCrystallize, triggerEdictOnReaction, fmt.Sprintf("fracturedhalo-lunarcrystallize-%v", char.Base.Key.String()))

	// Shields created directly by the character's abilities count (e.g., Thoma/Ineffa/Zhongli).
	// Crystallize shard pickup should NOT count for this weapon, so exclude shield.Crystallize here
	// and rely on the reaction event hook above instead.
	c.Events.Subscribe(event.OnShielded, func(args ...any) bool {
		shd := args[0].(shield.Shield)
		if shd.ShieldOwner() != char.Index() {
			return false
		}
		if shd.Type() == shield.Crystallize {
			return false
		}
		if !char.StatModIsActive(atkModKey) {
			return false
		}
		char.AddStatus(edictStatus, atkBuffFrames, true)
		return false
	}, fmt.Sprintf("fracturedhalo-shield-ability-%v", char.Base.Key.String()))

	installFracturedHaloTeamBonus(c)

	return w, nil
}

func installFracturedHaloTeamBonus(c *core.Core) {
	if c.Flags.Custom[teamFlagKey] != 0 {
		return
	}
	c.Flags.Custom[teamFlagKey] = 1

	for _, dst := range c.Player.Chars() {
		dst.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(teamModKey, -1),
			Amount: func(ai info.AttackInfo) (float64, bool) {
				return fracturedHaloBestReactBonus(c, ai), false
			},
		})
	}
}

func fracturedHaloBestReactBonus(c *core.Core, ai info.AttackInfo) float64 {
	if ai.AttackTag != attacks.AttackTagReactionLunarCharge && ai.AttackTag != attacks.AttackTagDirectLunarCharged {
		return 0
	}

	best := 0.0
	for _, src := range c.Player.Chars() {
		srcR := src.Tag(refineTagKey)
		if srcR == 0 {
			continue
		}
		if !src.StatusIsActive(edictStatus) {
			continue
		}

		bonus := 0.3 + 0.1*float64(srcR) // 0.4/0.5/0.6/0.7/0.8
		if bonus > best {
			best = bonus
		}
	}

	return best
}
