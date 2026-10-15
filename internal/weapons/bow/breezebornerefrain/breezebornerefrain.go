package breezebornerefrain

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// Grants 20%-40% Energy Recharge. Elemental Skill or Burst hits build up to
// 3 stacks, once every 0.03s; at 3 stacks, they are consumed to grant nearby
// party members 24%-48% Stellar Glimmer reaction DMG for 12s. No stacks can be
// gained while this buff is active, and both effects work off-field.
const (
	viperKey        = "thus-lied-the-viper"
	hymnStackICDKey = "hymn-of-the-pure-stack-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.BreezeborneRefrain, NewWeapon)
}

type Weapon struct {
	Index        int
	stackCounter int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	energyRecharge := 0.15 + 0.05*r
	stellarBonus := 0.18 + 0.06*r

	er := make([]float64, attributes.EndStatType)
	er[attributes.ER] = energyRecharge
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("breezeborne-refrain-er", -1),
		AffectedStat: attributes.ER,
		Amount: func() []float64 {
			return er
		},
	})

	for _, partyChar := range c.Player.Chars() {
		partyChar.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase("breezeborne-refrain-stellar", -1),
			Amount: func(ai info.AttackInfo) float64 {
				if !char.StatusIsActive(viperKey) || !isStellarReaction(ai.AttackTag) {
					return 0
				}
				return stellarBonus
			},
		})
	}

	onSkillOrBurstHit := func(args ...any) {
		atk, ok := args[1].(*info.AttackEvent)
		if !ok {
			return
		}

		if atk.Info.ActorIndex != char.Index() ||
			char.StatusIsActive(viperKey) ||
			char.StatusIsActive(hymnStackICDKey) {
			return
		}

		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt:
		case attacks.AttackTagElementalArtHold:
		case attacks.AttackTagElementalBurst:
		default:
			return
		}

		char.AddStatus(hymnStackICDKey, 2, true)

		w.stackCounter++
		if w.stackCounter < 3 {
			return
		}
		w.stackCounter = 0
		char.AddStatus(viperKey, 12*60, true)
	}

	c.Events.Subscribe(event.OnEnemyHit, onSkillOrBurstHit, "breezeborne-refrain-on-hit-"+char.Base.Key.String())

	return w, nil
}

func isStellarReaction(tag attacks.AttackTag) bool {
	switch tag {
	case attacks.AttackTagDirectStellarConduct,
		attacks.AttackTagDirectStellarSwirl,
		attacks.AttackTagReactionStellarSwirl:
		return true
	default:
		return false
	}
}
