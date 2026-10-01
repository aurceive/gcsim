package clashofkings

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
	core.RegisterWeaponFunc(keys.ClashOfKings, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	atkGain := 0.15 + float64(r)*0.05
	emGain := 50 * float64(r)

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atkGain
	em := make([]float64, attributes.EndStatType)
	em[attributes.EM] = emGain

	const (
		buffKey = "clash-of-kings-laws"
		statKey = "clash-of-kings-laws-stat"
		icdKey  = "clash-of-kings-icd"
	)
	extended := false
	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusIsActive(icdKey) {
			return
		}
		extended = false
		char.AddStatus(buffKey, 6*60, true)
		char.AddStatus(icdKey, 12*60, true)
	}, fmt.Sprintf("clash-of-kings-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		if atk.Info.AttackTag == attacks.AttackTagExtra && !extended && char.StatusIsActive(buffKey) {
			char.ExtendStatus(buffKey, 6*60)
			extended = true
		}
	}, fmt.Sprintf("clash-of-kings-%v", char.Base.Key.String()))

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(statKey, -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			if char.StatusIsActive(buffKey) {
				return m
			}
			return nil
		},
	})
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("clash-of-kings-em", -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			if char.StatusIsActive(buffKey) {
				return em
			}
			return nil
		},
	})
	return w, nil
}
