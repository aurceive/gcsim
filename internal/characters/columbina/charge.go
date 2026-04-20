package columbina

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var (
	chargeFrames  []int // basic Charge Attack (C)
	cleanseFrames []int // Moondew Cleanse (Cb)
)

// Frame data source: https://www.youtube.com/watch?v=fw4ll8DuwlU
const (
	chargeHitmark    = 66
	cleanseDewAbsorb = 60 // frame at which dew is consumed
)

// Cb hitmarks (not evenly spaced). Source: same video.
var cleanseHitmarks = [3]int{78, 84, 102}

func init() {
	// Basic Charge Attack (C)
	// Source: https://www.youtube.com/watch?v=fw4ll8DuwlU
	chargeFrames = frames.InitAbilSlice(102) // C>W
	chargeFrames[action.ActionAttack] = 96   // C>N1
	chargeFrames[action.ActionCharge] = 88   // C>C
	chargeFrames[action.ActionDash] = 25     // assumed - no data
	chargeFrames[action.ActionJump] = 25     // assumed - no data
	chargeFrames[action.ActionSwap] = 25     // assumed - no data
	chargeFrames[action.ActionWalk] = 102    // C>W

	// Moondew Cleanse (Cb)
	// Source: https://www.youtube.com/watch?v=fw4ll8DuwlU
	cleanseFrames = frames.InitAbilSlice(127) // Cb>W
	cleanseFrames[action.ActionAttack] = 95   // Cb>N1
	cleanseFrames[action.ActionCharge] = 82   // Cb>Cb (note: Cb>C = 103)
	cleanseFrames[action.ActionDash] = 25     // assumed - no data
	cleanseFrames[action.ActionJump] = 25     // assumed - no data
	cleanseFrames[action.ActionSwap] = 25     // assumed - no data
	cleanseFrames[action.ActionWalk] = 127    // Cb>W
}

// ActionStamina implements character.StaminaProvider.
// Moondew Cleanse does not consume stamina.
func (c *char) ActionStamina(a action.Action, p map[string]int) action.StaminaSpec {
	switch a {
	case action.ActionCharge:
		if c.Core.Player.AvailableDew() > 0 {
			return action.StaminaSpec{}
		}
		return action.StaminaSpec{
			Requirement: 50,
			Consume:     50,
			Timing:      action.StaminaConsumeOnExec,
		}
	case action.ActionDash:
		return action.StaminaSpec{
			Requirement: 18,
			Consume:     18,
			Timing:      action.StaminaConsumeByAbility,
		}
	default:
		return action.StaminaSpec{}
	}
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.Core.Player.AvailableDew() > 0 {
		return c.moondewCleanse(p)
	}
	return c.basicCharge(p)
}

// basicCharge is the standard Hydro CA with stamina cost.
func (c *char) basicCharge(p map[string]int) (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 2.0),
		chargeHitmark,
		chargeHitmark,
	)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] },
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeFrames[action.ActionDash],
		State:           action.ChargeAttackState,
	}, nil
}

// moondewCleanse consumes 1 Verdant Dew (or Moonridge Dew) and deals
// 3 instances of Dendro DMG based on Max HP. This DMG is Lunar-Bloom DMG.
// Dew availability is checked immediately, but consumption happens at frame 60.
func (c *char) moondewCleanse(p map[string]int) (action.Info, error) {
	// Consume dew at the absorption frame (60), not immediately
	c.Core.Tasks.Add(func() {
		c.Core.Player.ConsumeVerdantDew(1)
		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("columbina moondew cleanse consuming 1 dew", glog.LogCharacterEvent, c.Index())
		}
	}, cleanseDewAbsorb)

	for i := 0; i < 3; i++ {
		ai := info.AttackInfo{
			ActorIndex:       c.Index(),
			Abil:             "Moondew Cleanse (Lunar-Bloom)",
			AttackTag:        attacks.AttackTagDirectLunarBloom,
			ICDTag:           attacks.ICDTagNone,
			ICDGroup:         attacks.ICDGroupDefault,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Dendro,
			Durability:       0,
			UseHP:            true,
			Mult:             chargeLC[c.TalentLvlAttack()],
			IgnoreDefPercent: 1,
		}
		if i == 0 {
			ai.Durability = 25
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				info.Point{Y: -2},
				5, 5,
			),
			cleanseHitmarks[i],
			cleanseHitmarks[i],
		)
	}

	return action.Info{
		Frames:          func(next action.Action) int { return cleanseFrames[next] },
		AnimationLength: cleanseFrames[action.InvalidAction],
		CanQueueAfter:   cleanseFrames[action.ActionDash],
		State:           action.ChargeAttackState,
	}, nil
}
