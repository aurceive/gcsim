package hymnofthemaelstrom

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	buffKey  = "hymn-of-the-maelstrom-buff"
	boostKey = "hymn-of-the-maelstrom-boost"
	teamKey  = "hymn-of-the-maelstrom-team-atk"
)

type Weapon struct {
	Index  int
	core   *core.Core
	char   *character.CharWrapper
	refine int
	stacks int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	for _, partyChar := range w.core.Player.Chars() {
		atk := make([]float64, attributes.EndStatType)
		partyChar.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(teamKey, -1),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				if w.core.Player.Active() != partyChar.Index() {
					return nil
				}
				best := 0.0
				for _, member := range w.core.Player.Chars() {
					other, ok := member.Equip.Weapon.(*Weapon)
					if ok {
						best = max(best, other.atkBonus())
					}
				}
				atk[attributes.ATKP] = best
				return atk
			},
		})
	}
	return nil
}

func (w *Weapon) atkBonus() float64 {
	if w.char.StatusDuration(buffKey) == 0 {
		return 0
	}
	r := float64(w.refine)
	over := max(w.char.MaxHP()-40000, 0)
	bonus := min(0.06+0.02*r, over/1000*(0.003+0.001*r)) * float64(w.stacks)
	if w.char.StatusDuration(boostKey) > 0 {
		bonus *= 1.75
	}
	return bonus
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{core: c, char: char, refine: p.Refine}
	gain := 0.03 + 0.01*float64(p.Refine)
	heal := make([]float64, attributes.EndStatType)
	heal[attributes.Heal] = gain
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("hymn-of-the-maelstrom-heal", -1),
		AffectedStat: attributes.Heal,
		Amount:       func() []float64 { return heal },
	})
	hp := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("hymn-of-the-maelstrom-hp", -1),
		AffectedStat: attributes.HPP,
		Amount: func() []float64 {
			if char.StatusDuration(buffKey) == 0 {
				return nil
			}
			hp[attributes.HPP] = gain * float64(w.stacks)
			if char.StatusDuration(boostKey) > 0 {
				hp[attributes.HPP] *= 1.75
			}
			return hp
		},
	})
	groupFrame := -1
	groups := make(map[*info.HealGroup]struct{})
	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		hi := args[0].(*info.HealInfo)
		if hi.Caller != char.Index() {
			return
		}
		if hi.Group != nil {
			if groupFrame != c.F {
				clear(groups)
				groupFrame = c.F
			}
			if _, seen := groups[hi.Group]; seen {
				return
			}
			groups[hi.Group] = struct{}{}
		} else {
			// Player.Heal emits one event per recipient, in party order.
			if hi.Target == -1 && args[1].(int) != 0 {
				return
			}
		}
		if char.StatusDuration(buffKey) == 0 {
			w.stacks = 0
		}
		w.stacks = min(w.stacks+1, 3)
		char.AddStatus(buffKey, 10*60, true)
	}, "hymn-of-the-maelstrom-heal-"+char.Base.Key.String())
	boost := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex < 0 || atk.Info.ActorIndex >= len(c.Player.Chars()) {
			return
		}
		char.AddStatus(boostKey, 5*60, true)
	}
	c.Events.Subscribe(event.OnFrozen, boost, "hymn-of-the-maelstrom-frozen-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnStellarSwirl, boost, "hymn-of-the-maelstrom-swirl-"+char.Base.Key.String())
	return w, nil
}
