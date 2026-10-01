package hymnofthemaelstrom

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.HymnOfTheMaelstrom, NewWeapon)
}

type Weapon struct {
	Core   *core.Core
	Char   *character.CharWrapper
	Refine int
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine
	w.Core, w.Char, w.Refine = c, char, r
	hpGain := 0.02 + float64(r)*0.02
	atkPerK := 0.002 + float64(r)*0.002
	atkCap := 0.06 + float64(r)*0.02
	healBonus := 0.02 + float64(r)*0.02
	stacks := 0
	const (
		buffKey  = "hymn-of-the-maelstrom-buff"
		statKey  = "hymn-of-the-maelstrom-hp-stat"
		boostKey = "hymn-of-the-maelstrom-boost"
	)
	mHeal := make([]float64, attributes.EndStatType)
	mHeal[attributes.Heal] = healBonus
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("hymn-of-the-maelstrom-heal", -1),
		AffectedStat: attributes.Heal,
		Amount:       func() []float64 { return mHeal },
	})
	mHP := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base: modifier.NewBase(statKey, -1), AffectedStat: attributes.HPP,
		Amount: func() []float64 {
			mHP[attributes.HPP] = hpGain * float64(stacks)
			if char.StatusIsActive(boostKey) {
				mHP[attributes.HPP] *= 1.75
			}
			return mHP
		},
	})
	onHeal := func(args ...any) {
		source := args[0].(*info.HealInfo)
		if source.Caller != char.Index() {
			return
		}
		if !char.StatusIsActive(buffKey) {
			stacks = 0
		}
		stacks = min(stacks+1, 3)
		char.AddStatus(buffKey, 10*60, true)
	}
	c.Events.Subscribe(event.OnHeal, onHeal, fmt.Sprintf("hymn-of-the-maelstrom-heal-%v", char.Base.Key.String()))
	boost := func(args ...any) { char.AddStatus(boostKey, 5*60, true) }
	c.Events.Subscribe(event.OnFrozen, boost, fmt.Sprintf("hymn-of-the-maelstrom-frozen-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStellarSwirl, boost, fmt.Sprintf("hymn-of-the-maelstrom-swirl-%v", char.Base.Key.String()))
	for _, partyChar := range c.Player.Chars() {
		atk := make([]float64, attributes.EndStatType)
		partyChar.AddStatMod(character.StatMod{
			Base: modifier.NewBase("hymn-of-the-maelstrom-atk", -1), AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				if c.Player.Active() != partyChar.Index() || !char.StatusIsActive(buffKey) {
					return nil
				}
				over := char.MaxHP() - 40000
				if over < 0 {
					over = 0
				}
				atk[attributes.ATKP] = min(atkCap, over*0.001*atkPerK) * float64(stacks)
				if char.StatusIsActive(boostKey) {
					atk[attributes.ATKP] *= 1.75
				}
				return atk
			},
		})
	}

	return w, nil
}
