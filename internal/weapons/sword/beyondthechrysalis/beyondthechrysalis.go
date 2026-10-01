package beyondthechrysalis

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
	devotionKey = "beyond-the-chrysalis-devotion"
	defianceKey = "beyond-the-chrysalis-defiance"
	plentyICD   = "beyond-the-chrysalis-plenty-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.BeyondTheChrysalis, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	critDamage := 0.4 + 0.16*r
	stellarSwirlBonus := 0.27 + 0.09*r
	energyRestore := 4.5 + 0.5*r
	sequence := 0

	cd := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("beyond-the-chrysalis-cd", -1),
		AffectedStat: attributes.CD,
		Amount: func() []float64 {
			if !char.StatusIsActive(devotionKey) {
				return nil
			}
			cd[attributes.CD] = critDamage
			return cd
		},
	})
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("beyond-the-chrysalis-stellar-swirl", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if !char.StatusIsActive(defianceKey) {
				return 0
			}
			switch ai.AttackTag {
			case attacks.AttackTagDirectStellarSwirl, attacks.AttackTagReactionStellarSwirl:
				return stellarSwirlBonus
			default:
				return 0
			}
		},
	})

	onSkillOrBurst := func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		switch sequence {
		case 0:
			char.AddStatus(devotionKey, 10*60, true)
		case 1:
			char.AddStatus(defianceKey, 10*60, true)
		case 2:
			if !char.StatusIsActive(plentyICD) {
				char.AddStatus(plentyICD, 4*60, true)
				char.AddEnergy("beyond-the-chrysalis", energyRestore)
			}
		}
		sequence = (sequence + 1) % 3
	}
	c.Events.Subscribe(event.OnSkill, onSkillOrBurst, "beyond-the-chrysalis-skill-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnBurst, onSkillOrBurst, "beyond-the-chrysalis-burst-"+char.Base.Key.String())

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev := args[0].(int)
		next := args[1].(int)
		if prev != char.Index() || next == char.Index() {
			return
		}
		char.DeleteStatus(devotionKey)
		char.DeleteStatus(defianceKey)
		sequence = 0
	}, fmt.Sprintf("beyond-the-chrysalis-swap-%v", char.Base.Key.String()))

	return w, nil
}
