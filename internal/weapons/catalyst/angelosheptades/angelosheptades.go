package angelosheptades

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	buffKey = "angelosheptades-pathfinderlight"
	icdKey  = "angelosheptades-icd"
	buffDur = 20 * 60 // 20s = 1200 frames
	icdDur  = 14 * 60 // 14s = 840 frames
)

func init() {
	core.RegisterWeaponFunc(keys.AngelosHeptades, NewWeapon)
}

type Weapon struct {
	Index   int
	tickSrc int
	bonus   []float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Crown of the Final Scion:
// ATK is increased by 12/15/18/21/24%.
// After the equipping character creates a Shield, they gain "Pathfinder's Light" for 20s:
// Increases your active party member's DMG by 10/13/16/19/22% for every 1,000 ATK the equipping
// character has, up to a maximum of 26/34/42/50/58%.
// Additionally, when the equipping character creates a Shield, they will also gain "Guide's Contentment":
// Restores 14/15/16/17/18 Elemental Energy to the equipping character.
// The aforementioned effect can trigger once every 14s.
// The equipping character may trigger this effect even when off-field.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	w.bonus = make([]float64, attributes.EndStatType)
	r := float64(p.Refine)

	// Static ATK% bonus: 9% + 3% * refine (R1→12%, R5→24%)
	mAtk := make([]float64, attributes.EndStatType)
	mAtk[attributes.ATKP] = 0.09 + 0.03*r
	char.AddStatMod(character.StatMod{
		Base: modifier.NewBase("angelosheptades-atkp", -1),
		Amount: func() []float64 {
			return mAtk
		},
	})

	// DMG per 1000 ATK: 7% + 3%*refine (R1→10%, R5→22%)
	perThousandATK := 0.07 + 0.03*r
	// Max DMG cap: 18% + 8%*refine (R1→26%, R5→58%)
	maxDMG := 0.18 + 0.08*r
	// Energy restore: 13 + refine (R1→14, R5→18)
	energyRestore := 13.0 + r

	c.Events.Subscribe(event.OnShielded, func(args ...any) {
		shd := args[0].(shield.Shield)
		if shd.ShieldOwner() != char.Index() {
			return
		}
		if char.StatusIsActive(icdKey) {
			return
		}
		char.AddStatus(icdKey, icdDur, true)

		// Guide's Contentment: restore energy to equipping character
		char.AddEnergy("angelosheptades-energy", energyRestore)

		src := c.F
		w.tickSrc = src
		char.QueueCharTask(func() {
			if src != w.tickSrc {
				return
			}
			for _, other := range c.Player.Chars() {
				other.DeleteAttackMod(buffKey)
			}
		}, buffDur)

		for _, x := range c.Player.Chars() {
			this := x
			this.AddAttackMod(character.AttackMod{
				Base: modifier.NewBase(buffKey, -1),
				Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
					if c.Player.Active() != this.Index() {
						return nil
					}
					// Pathfinder's Light: DMG bonus scales with holder's current ATK
					dmgBonus := char.TotalAtk() / 1000.0 * perThousandATK
					if dmgBonus > maxDMG {
						dmgBonus = maxDMG
					}
					w.bonus[attributes.DmgP] = dmgBonus
					return w.bonus
				},
			})
		}
	}, fmt.Sprintf("angelosheptades-onshielded-%v", char.Base.Key.String()))

	return w, nil
}
