package songofthevigil

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

// Triggering an Elemental Reaction restores 4-8 Energy to the equipping character
// once every 9s. Triggering a Stellar Glimmer reaction grants 20%-40% ATK for 12s.
// Both effects can trigger while the equipping character is off-field.
const energyICDKey = "song-of-the-vigil-energy-icd"

func init() {
	core.RegisterWeaponFunc(keys.SongOfTheVigil, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	energy := float64(3 + p.Refine)
	atkBonus := make([]float64, attributes.EndStatType)
	atkBonus[attributes.ATKP] = 0.15 + float64(p.Refine)*0.05

	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || char.StatusIsActive(energyICDKey) {
			return
		}

		char.AddStatus(energyICDKey, 9*60, true)
		char.AddEnergy("song-of-the-vigil-energy", energy)
	}

	onStellar := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("song-of-the-vigil-atk", 12*60),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return atkBonus
			},
		})
	}

	for evt := event.ReactionEventStartDelim + 1; evt < event.ReactionEventEndDelim; evt++ {
		c.Events.Subscribe(evt, onReaction, "song-of-the-vigil-on-reaction-"+char.Base.Key.String())
	}

	c.Events.Subscribe(event.OnStellarConduct, onStellar, fmt.Sprintf("song-of-the-vigil-on-stellar-conduct-%s", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStellarSwirl, onStellar, fmt.Sprintf("song-of-the-vigil-on-stellar-swirl-%s", char.Base.Key.String()))

	return &Weapon{}, nil
}
