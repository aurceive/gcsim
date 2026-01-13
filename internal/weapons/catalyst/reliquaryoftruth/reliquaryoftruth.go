package reliquaryoftruth

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
	core.RegisterWeaponFunc(keys.ReliquaryOfTruth, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	crKey         = "reliquaryoftruth-cr"
	secretKey     = "reliquaryoftruth-secret"
	moonKey       = "reliquaryoftruth-moon"
	secretStatus  = "reliquaryoftruth-secret-status"
	moonStatus    = "reliquaryoftruth-moon-status"
	overlapMult   = 1.5
	secretDur     = 12 * 60
	moonDur       = 4 * 60
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	crBonus := 0.06 + 0.02*float64(r)   // 8/10/12/14/16%
	emBonus := 60 + 20*float64(r)       // 80/100/120/140/160
	cdBonus := 0.18 + 0.06*float64(r)   // 24/30/36/42/48%

	mCR := make([]float64, attributes.EndStatType)
	mCR[attributes.CR] = crBonus
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(crKey, -1),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			return mCR, true
		},
	})

	mEM := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(secretKey, -1),
		AffectedStat: attributes.EM,
		Amount: func() ([]float64, bool) {
			if !char.StatusIsActive(secretStatus) {
				return mEM, false
			}
			val := emBonus
			if char.StatusIsActive(moonStatus) {
				val *= overlapMult
			}
			mEM[attributes.EM] = val
			return mEM, true
		},
	})

	mCD := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(moonKey, -1),
		AffectedStat: attributes.CD,
		Amount: func() ([]float64, bool) {
			if !char.StatusIsActive(moonStatus) {
				return mCD, false
			}
			val := cdBonus
			if char.StatusIsActive(secretStatus) {
				val *= overlapMult
			}
			mCD[attributes.CD] = val
			return mCD, true
		},
	})

	c.Events.Subscribe(event.OnSkill, func(args ...any) bool {
		if c.Player.Active() != char.Index() {
			return false
		}
		char.AddStatus(secretStatus, secretDur, true)
		return false
	}, fmt.Sprintf("reliquaryoftruth-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) bool {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagDirectLunarBloom {
			return false
		}
		char.AddStatus(moonStatus, moonDur, true)
		return false
	}, fmt.Sprintf("reliquaryoftruth-lb-dmg-%v", char.Base.Key.String()))

	return w, nil
}
