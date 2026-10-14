package starsealededict

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

// Grants 24%-48% CRIT DMG. Using an Elemental Skill grants 48-96 EM for 20s.
// While active, nearby party members' Elemental Reactions grant up to 3 stacks;
// each adds 24-48 EM and 12%-24% CRIT DMG to Stellar Conduct.
const (
	associationKey = "starsealed-edict-formula-association"
)

func init() {
	core.RegisterWeaponFunc(keys.StarsealedEdict, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	baseEM, emPerStack, critDMGPerStack := passiveBonuses(p.Refine)
	stacks := 0

	critDMG := make([]float64, attributes.EndStatType)
	critDMG[attributes.CD] = 0.18 + 0.06*float64(p.Refine)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("starsealed-edict-crit-dmg", -1),
		AffectedStat: attributes.CD,
		Amount:       func() []float64 { return critDMG },
	})

	em := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("starsealed-edict-em", -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			if !char.StatusIsActive(associationKey) {
				return nil
			}
			em[attributes.EM] = baseEM + emPerStack*float64(stacks)
			return em
		},
	})

	stellarCritDMG := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("starsealed-edict-stellar-conduct-crit", -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if !char.StatusIsActive(associationKey) ||
				atk.Info.AttackTag != attacks.AttackTagDirectStellarConduct ||
				stacks == 0 {
				return nil
			}
			stellarCritDMG[attributes.CD] = critDMGPerStack * float64(stacks)
			return stellarCritDMG
		},
	})

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		stacks = 0
		char.AddStatus(associationKey, 20*60, true)
	}, fmt.Sprintf("starsealed-edict-skill-%s", char.Base.Key.String()))

	onReaction := func(args ...any) {
		if len(args) < 2 {
			return
		}
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return
		}
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || !char.StatusIsActive(associationKey) || stacks >= 3 {
			return
		}
		for _, partyChar := range c.Player.Chars() {
			if partyChar.Index() == atk.Info.ActorIndex {
				stacks++
				return
			}
		}
	}
	for reaction := event.ReactionEventStartDelim + 1; reaction < event.ReactionEventEndDelim; reaction++ {
		c.Events.Subscribe(reaction, onReaction,
			fmt.Sprintf("starsealed-edict-reaction-%s-%d", char.Base.Key.String(), reaction))
	}

	return w, nil
}

func passiveBonuses(refine int) (float64, float64, float64) {
	return float64(36 + 12*refine), float64(18 + 6*refine), 0.09 + 0.03*float64(refine)
}
