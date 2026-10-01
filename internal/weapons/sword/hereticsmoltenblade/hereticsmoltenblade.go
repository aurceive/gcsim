package hereticsmoltenblade

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

const (
	gleamKey     = "heretics-molten-blade-gleam"
	skillICDKey  = "heretics-molten-blade-skill-icd"
	maxTravelled = 7.0
)

func init() {
	core.RegisterWeaponFunc(keys.HereticsMoltenBlade, NewWeapon)
}

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
			if !char.StatusIsActive(gleamKey) {
				return nil
			}
			atk[attributes.ATKP] = atkBonus
			return atk
		},
	})

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusIsActive(skillICDKey) {
			return
		}
		char.AddStatus(gleamKey, 14*60, true)
		char.AddStatus(skillICDKey, 14*60, true)
		atkBonus = minATK
		distance = 0
		frames = 0
		previousPosition = c.Combat.Player().Pos()
	}, fmt.Sprintf("heretics-molten-blade-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnTick, func(args ...any) {
		if !char.StatusIsActive(gleamKey) || c.Player.Active() != char.Index() {
			return
		}

		position := c.Combat.Player().Pos()
		distance += position.Distance(previousPosition)
		previousPosition = position
		frames++
		if frames < 60 {
			return
		}

		travelScale := distance / maxTravelled
		if travelScale > 1 {
			travelScale = 1
		}
		atkBonus = minATK + (maxATK-minATK)*travelScale
		distance = 0
		frames = 0
	}, fmt.Sprintf("heretics-molten-blade-distance-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev := args[0].(int)
		next := args[1].(int)
		if prev != char.Index() || next == char.Index() {
			return
		}
		char.DeleteStatus(gleamKey)
		distance = 0
		frames = 0
		atkBonus = minATK
	}, fmt.Sprintf("heretics-molten-blade-swap-%v", char.Base.Key.String()))

	return w, nil
}
