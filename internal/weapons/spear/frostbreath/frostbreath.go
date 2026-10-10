package frostbreath

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	buffKey    = "frostbreath-atk"
	energyKey  = "frostbreath-energy"
	buffICDKey = "frostbreath-buff-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.Frostbreath, NewWeapon)
}

type Weapon struct {
	Index  int
	char   *character.CharWrapper
	core   *core.Core
	refine int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	atkBonus := make([]float64, attributes.EndStatType)
	atkBonus[attributes.ATKP] = 0.15 + float64(w.refine)*0.05
	energy := 4.5 + float64(w.refine)*1.5

	onHydroOrCryoReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != w.char.Index() || w.char.StatusIsActive(buffICDKey) {
			return
		}

		w.char.AddStatus(buffICDKey, 16*60, true)
		w.char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, 15*60),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return atkBonus
			},
		})

		for _, member := range w.core.Player.Chars() {
			if member.Index() == w.char.Index() {
				continue
			}
			member.AddEnergy(energyKey, energy)
		}
	}

	key := "frostbreath-on-hydro-or-cryo-reaction-" + w.char.Base.Key.String()
	for _, evt := range []event.Event{
		event.OnMelt,
		event.OnFrozen,
		event.OnShatter,
		event.OnSuperconduct,
		event.OnSwirlCryo,
		event.OnSwirlHydro,
		event.OnCrystallizeCryo,
		event.OnCrystallizeHydro,
		event.OnVaporize,
		event.OnElectroCharged,
		event.OnBloom,
		event.OnLunarCharged,
		event.OnLunarBloom,
		event.OnLunarCrystallize,
		event.OnStellarConduct,
		event.OnStellarSwirl,
	} {
		w.core.Events.Subscribe(evt, onHydroOrCryoReaction, key)
	}

	return nil
}

// Triggering a Cryo- or Hydro-related reaction grants the equipping character
// 20%-40% ATK for 15s and restores 6-12 Energy to the other party members.
// The effect can trigger once every 16s, including while the equipping character
// is off-field.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	return &Weapon{
		char:   char,
		core:   c,
		refine: p.Refine,
	}, nil
}
