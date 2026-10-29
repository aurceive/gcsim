package hereticsmoltenblade

import (
	"math"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	gleamKey    = "heretics-molten-blade-gleam"
	skillICDKey = "heretics-molten-blade-skill-icd"
)

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	minATK := 0.135 + 0.045*r
	maxATK := 0.27 + 0.09*r
	atkBonus := minATK
	distance := 0.0
	frames := 0
	var previousPosition info.Point
	atk := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("heretics-molten-blade-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			if char.StatusDuration(gleamKey) == 0 {
				return nil
			}
			atk[attributes.ATKP] = atkBonus
			return atk
		},
	})
	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusDuration(skillICDKey) > 0 {
			return
		}
		char.AddStatus(gleamKey, 14*60, true)
		char.AddStatus(skillICDKey, 14*60, true)
		atkBonus, distance, frames = minATK, 0, 0
		previousPosition = c.Combat.Player().Pos()
	}, "heretics-molten-blade-skill-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnMovement, func(args ...any) {
		if char.StatusDuration(gleamKey) > 0 && c.Player.Active() == char.Index() {
			moved := args[0].(float64)
			if moved < 0 || math.IsNaN(moved) || math.IsInf(moved, 0) {
				c.Log.NewEvent("invalid movement distance for heretics molten blade", glog.LogWarnings, char.Index()).
					Write("distance", moved)
				return
			}
			distance += moved
		}
	}, "heretics-molten-blade-movement-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnTick, func(args ...any) {
		if char.StatusDuration(gleamKey) == 0 || c.Player.Active() != char.Index() {
			return
		}
		position := c.Combat.Player().Pos()
		distance += position.Distance(previousPosition)
		previousPosition = position
		frames++
		if frames < 60 {
			return
		}
		atkBonus = minATK + (maxATK-minATK)*min(distance/7, 1)
		distance, frames = 0, 0
	}, "heretics-molten-blade-distance-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev, next := args[0].(int), args[1].(int)
		if prev != char.Index() || next == char.Index() {
			return
		}
		char.DeleteStatus(gleamKey)
		distance, frames, atkBonus = 0, 0, minATK
	}, "heretics-molten-blade-swap-"+char.Base.Key.String())
	return w, nil
}
