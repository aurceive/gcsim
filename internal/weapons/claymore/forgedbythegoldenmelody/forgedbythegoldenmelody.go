package forgedbythegoldenmelody

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

// Every 10s, the equipping character cycles through 18%-36% ATK, 120-240 EM,
// and 28%-56% Stellar Glimmer reaction DMG, each lasting 10s. Triggering a
// Stellar Glimmer reaction grants a separate 12s copy of the current buff, once
// every 12s; this copy stacks with the current buff. Both effects can trigger
// while the equipping character is off-field.
const (
	weaponKey     = "golden-melody"
	atkBuffSuffix = "-atk"
	emBuffSuffix  = "-em"
	stellarSuffix = "-stellar"
	contraICDKey  = "golden-melody-contrapuntal-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.ForgedByTheGoldenMelody, NewWeapon)
}

type Weapon struct {
	Index       int
	curBuff     BuffType
	char        *character.CharWrapper
	atkBonus    []float64
	emBonus     []float64
	stellarBuff float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		char:     char,
		curBuff:  BuffType(p.Params["buff_type"] % 3),
		atkBonus: make([]float64, attributes.EndStatType),
		emBonus:  make([]float64, attributes.EndStatType),
	}
	w.atkBonus[attributes.ATKP] = 0.135 + float64(p.Refine)*0.045
	w.emBonus[attributes.EM] = float64(90 + 30*p.Refine)
	w.stellarBuff = 0.21 + float64(p.Refine)*0.07

	w.applyBuff(false, 10*60)
	char.QueueCharTask(w.switchBuff, 10*60)

	onStellar := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || char.StatusIsActive(contraICDKey) {
			return
		}

		char.AddStatus(contraICDKey, 12*60, true)
		w.applyBuff(true, 12*60)
	}

	c.Events.Subscribe(event.OnStellarConduct, onStellar, "forged-by-the-golden-melody-on-stellar-conduct-"+char.Base.Key.String())
	c.Events.Subscribe(event.OnStellarSwirl, onStellar, "forged-by-the-golden-melody-on-stellar-swirl-"+char.Base.Key.String())

	return w, nil
}

func (w *Weapon) switchBuff() {
	w.curBuff = w.curBuff.Next()
	w.applyBuff(false, 10*60)
	w.char.QueueCharTask(w.switchBuff, 10*60)
}

func (w *Weapon) applyBuff(contrapuntal bool, duration int) {
	middle := ""
	if contrapuntal {
		middle = "-contrapuntal"
	}

	switch w.curBuff {
	case EmBuff:
		w.char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(weaponKey+middle+emBuffSuffix, duration),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return w.emBonus
			},
		})
	case StellarBuff:
		w.char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag(weaponKey+middle+stellarSuffix, duration),
			Amount: func(ai info.AttackInfo) float64 {
				if isStellarDamageTag(ai.AttackTag) {
					return w.stellarBuff
				}
				return 0
			},
		})
	case AtkBuff:
		w.char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(weaponKey+middle+atkBuffSuffix, duration),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return w.atkBonus
			},
		})
	default:
		panic(fmt.Sprintf("golden melody: unknown buff type: %v", w.curBuff))
	}
}

func isStellarDamageTag(tag attacks.AttackTag) bool {
	return attacks.AttackTagIsStellar(tag) || tag == attacks.AttackTagReactionStellarSwirl
}

type BuffType int

const (
	AtkBuff BuffType = iota
	EmBuff
	StellarBuff
)

func (t BuffType) Next() BuffType {
	return (t + 1) % 3
}
