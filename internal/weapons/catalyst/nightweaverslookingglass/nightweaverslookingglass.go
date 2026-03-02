package nightweaverslookingglass

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
	core.RegisterWeaponFunc(keys.NightweaversLookingGlass, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	prayerStatus = "nightweaverslookingglass-prayer"
	verseStatus  = "nightweaverslookingglass-verse"

	prayerModKey = "nightweaverslookingglass-prayer-em"
	verseModKey  = "nightweaverslookingglass-verse-em"

	teamFlagKey  = "nightweaverslookingglass-team-installed"
	teamModKey   = "nightweaverslookingglass-team-react"
	refineTagKey = "nightweaverslookingglass-refine"
)

// Nightweaver's Looking Glass (i_n14520):
//   - Grants two separate self EM buffs, each scaling with refinement (R1..R5: +60/+75/+90/+105/+120 EM).
//     These two buffs can overlap and stack additively (i.e., +2x EM when both are active).
//   - Prayer: when the equipping character's Elemental Skill deals Hydro/Dendro DMG, gain Prayer for 4.5s.
//     Implemented via OnEnemyDamage + Skill attack tags; this can trigger off-field if the Skill continues dealing damage.
//   - New Moon Verse: when a party member (player-controlled character) triggers Lunar-Bloom, gain Verse for 10s.
//   - While both Prayer and Verse are active on any holder, all party members gain a non-stacking reaction DMG bonus
//     (this is an additive multiplier term in gcsim's reaction damage formula):
//     Bloom: +120/150/180/210/240%
//     Hyperbloom/Burgeon: +80/100/120/140/160%
//     Direct Lunar-Bloom DMG: +40/50/60/70/80%
//     If multiple copies are active, the best (highest) applicable bonus is used.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// Store refine so the shared team mod can compute the non-stacking best bonus.
	char.SetTag(refineTagKey, r)

	emBonus := 45 + 15*float64(r) // 60/75/90/105/120

	mPrayer := make([]float64, attributes.EndStatType)
	mPrayer[attributes.EM] = emBonus
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(prayerModKey, -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			if !char.StatusIsActive(prayerStatus) {
				return nil
			}
			return mPrayer
		},
	})

	mVerse := make([]float64, attributes.EndStatType)
	mVerse[attributes.EM] = emBonus
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(verseModKey, -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			if !char.StatusIsActive(verseStatus) {
				return nil
			}
			return mVerse
		},
	})

	// When the equipping character's Elemental Skill deals Hydro or Dendro DMG, gain Prayer.
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
			// ok
		default:
			return
		}
		if atk.Info.Element != attributes.Hydro && atk.Info.Element != attributes.Dendro {
			return
		}

		char.AddStatus(prayerStatus, 270, true) // 4.5s
	}, fmt.Sprintf("nightweaverslookingglass-skill-dmg-%v", char.Base.Key.String()))

	// When any party member triggers Lunar-Bloom, the equipping character gains New Moon Verse.
	c.Events.Subscribe(event.OnLunarBloom, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex < 0 || atk.Info.ActorIndex >= len(c.Player.Chars()) {
			return
		}
		char.AddStatus(verseStatus, 10*60, true)
	}, fmt.Sprintf("nightweaverslookingglass-lunarbloom-%v", char.Base.Key.String()))

	// Shared non-stacking team reaction bonus when both Prayer and Verse are active.
	installNightweaversLookingGlassTeamBonus(c)

	return w, nil
}

func installNightweaversLookingGlassTeamBonus(c *core.Core) {
	if c.Flags.Custom[teamFlagKey] != 0 {
		return
	}
	c.Flags.Custom[teamFlagKey] = 1

	for _, dst := range c.Player.Chars() {
		dst.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(teamModKey, -1),
			Amount: func(ai info.AttackInfo) float64 {
				return nightweaversLookingGlassBestReactBonus(c, ai)
			},
		})
	}
}

func nightweaversLookingGlassBestReactBonus(c *core.Core, ai info.AttackInfo) float64 {
	best := 0.0

	for _, src := range c.Player.Chars() {
		srcR := src.Tag(refineTagKey)
		if srcR == 0 {
			continue
		}
		if !src.StatusIsActive(prayerStatus) || !src.StatusIsActive(verseStatus) {
			continue
		}

		rf := float64(srcR)

		switch ai.AttackTag {
		case attacks.AttackTagBloom:
			bonus := 0.9 + 0.3*rf // 1.2/1.5/1.8/2.1/2.4
			if bonus > best {
				best = bonus
			}
		case attacks.AttackTagHyperbloom, attacks.AttackTagBurgeon:
			bonus := 0.6 + 0.2*rf // 0.8/1.0/1.2/1.4/1.6
			if bonus > best {
				best = bonus
			}
		case attacks.AttackTagDirectLunarBloom:
			bonus := 0.3 + 0.1*rf // 0.4/0.5/0.6/0.7/0.8
			if bonus > best {
				best = bonus
			}
		}
	}

	return best
}
