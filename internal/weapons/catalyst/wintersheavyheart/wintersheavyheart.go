package wintersheavyheart

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.WintersHeavyHeart, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	emPerCryo := 18 + float64(r)*6
	atkPerElectro := 0.036 + float64(r)*0.012
	mATK := make([]float64, attributes.EndStatType)

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("winters-heavy-heart", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			if c.Player.GetMoonsignLevel() >= 2 {
				mATK[attributes.ATKP] = 0
				return mATK
			}
			electro := 0
			for _, partyChar := range c.Player.Chars() {
				if partyChar.Base.Element == attributes.Electro {
					electro++
				}
			}
			mATK[attributes.ATKP] = atkPerElectro * float64(electro)
			if electro > 4 {
				mATK[attributes.ATKP] = atkPerElectro * 4
			}
			return mATK
		},
	})
	mEM := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("winters-heavy-heart-em", -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			if c.Player.GetMoonsignLevel() >= 2 {
				count := 0
				for _, partyChar := range c.Player.Chars() {
					if partyChar.Base.Element == attributes.Cryo || partyChar.Base.Element == attributes.Electro {
						count++
					}
				}
				mEM[attributes.EM] = (15 + float64(r)*5) * float64(min(count, 4))
				return mEM
			}
			cryo := 0
			for _, partyChar := range c.Player.Chars() {
				if partyChar.Base.Element == attributes.Cryo {
					cryo++
				}
			}
			mEM[attributes.EM] = emPerCryo * float64(cryo)
			if cryo > 4 {
				mEM[attributes.EM] = emPerCryo * 4
			}
			return mEM
		},
	})
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("winters-heavy-heart-react", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if c.Player.GetMoonsignLevel() < 2 {
				return 0
			}
			count := 0
			for _, partyChar := range c.Player.Chars() {
				if partyChar.Base.Element == attributes.Cryo || partyChar.Base.Element == attributes.Electro {
					count++
				}
			}
			switch ai.AttackTag {
			case attacks.AttackTagDirectStellarConduct, attacks.AttackTagDirectStellarSwirl, attacks.AttackTagReactionStellarSwirl:
				return (0.045 + float64(r)*0.015) * float64(min(count, 4))
			default:
				return 0
			}
		},
	})

	return w, nil
}
