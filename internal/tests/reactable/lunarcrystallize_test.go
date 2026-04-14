package reactable_test

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

func runDirectLunarCrystallizeDamage(t *testing.T, installHook func(*core.Core)) float64 {
	t.Helper()

	c, trg := makeCore(1)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	if installHook != nil {
		installHook(c)
	}

	snap := info.Snapshot{CharLvl: 90}
	snap.Stats[attributes.EM] = 100

	c.QueueAttackEvent(&info.AttackEvent{
		Info: info.AttackInfo{
			ActorIndex:       0,
			Abil:             "test-direct-lcr",
			AttackTag:        attacks.AttackTagDirectLunarCrystallize,
			ICDTag:           attacks.ICDTagNone,
			ICDGroup:         attacks.ICDGroupDefault,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Geo,
			UseEM:            true,
			Mult:             1,
			IgnoreDefPercent: 1,
		},
		Snapshot: snap,
		Pattern:  combat.NewCircleHitOnTarget(trg[0], nil, 100),
	}, 0)

	advanceCoreFrame(c)
	return c.Combat.TotalDamage
}

func runReactionLunarCrystallizeDamage(t *testing.T, installHook func(*core.Core)) float64 {
	t.Helper()

	c, trg := makeCore(1)
	c.Flags.Custom[reactable.LunarCrystallizeEnableKey] = 1
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}
	if installHook != nil {
		installHook(c)
	}

	c.QueueAttackEvent(&info.AttackEvent{
		Info: info.AttackInfo{
			ActorIndex: 0,
			Abil:       "test-hydro-aura",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 100,
		},
		Pattern: combat.NewCircleHitOnTarget(trg[0], nil, 100),
	}, 0)
	advanceCoreFrame(c)

	for range 3 {
		c.QueueAttackEvent(&info.AttackEvent{
			Info: info.AttackInfo{
				ActorIndex: 1,
				Abil:       "test-geo-trigger",
				AttackTag:  attacks.AttackTagElementalArt,
				ICDTag:     attacks.ICDTagNone,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeDefault,
				Element:    attributes.Geo,
				Durability: 25,
			},
			Pattern: combat.NewCircleHitOnTarget(trg[0], nil, 100),
		}, 0)
		advanceCoreFrame(c)
	}

	for range 10 {
		advanceCoreFrame(c)
	}

	return c.Combat.TotalDamage
}

func TestDirectLunarCrystallizeBaseDmgBonusUsesOnEnemyHit(t *testing.T) {
	baseline := runDirectLunarCrystallizeDamage(t, nil)
	bonusDamage := runDirectLunarCrystallizeDamage(t, func(c *core.Core) {
		c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			atk := args[1].(*info.AttackEvent)
			if atk.Info.AttackTag != attacks.AttackTagDirectLunarCrystallize {
				return
			}
			atk.Info.BaseDmgBonus += 1
		}, "test-direct-lcr-bonus")
	})

	if bonusDamage <= baseline {
		t.Fatalf("expected OnEnemyHit base damage bonus to increase direct LCr damage, baseline=%v bonus=%v", baseline, bonusDamage)
	}
}

func TestReactionLunarCrystallizeBaseDmgBonusUsesReactionHook(t *testing.T) {
	baseline := runReactionLunarCrystallizeDamage(t, nil)
	enemyHitDamage := runReactionLunarCrystallizeDamage(t, func(c *core.Core) {
		c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			atk := args[1].(*info.AttackEvent)
			if atk.Info.AttackTag != attacks.AttackTagReactionLunarCrystallize {
				return
			}
			atk.Info.BaseDmgBonus += 1
		}, "test-reaction-lcr-enemy-hit")
	})
		reactionHookDamage := runReactionLunarCrystallizeDamage(t, func(c *core.Core) {
			c.Events.Subscribe(event.OnLunarCrystallizeReactionAttack, func(args ...any) {
			atk := args[1].(*info.AttackEvent)
			if atk.Info.AttackTag != attacks.AttackTagReactionLunarCrystallize {
				return
			}
			atk.Info.BaseDmgBonus += 1
		}, "test-reaction-lcr-hook")
	})

	if baseline <= 0 {
		t.Fatalf("expected positive baseline reaction LCr damage, got %v", baseline)
	}
	if math.Abs(enemyHitDamage-baseline) > 1e-9 {
		t.Fatalf("expected OnEnemyHit base damage bonus to be ineffective for reaction LCr, baseline=%v enemyHit=%v", baseline, enemyHitDamage)
	}
	if reactionHookDamage <= baseline {
		t.Fatalf("expected OnLunarCrystallizeReactionAttack base damage bonus to increase reaction LCr damage, baseline=%v reaction=%v", baseline, reactionHookDamage)
	}
}