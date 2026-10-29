package kokomi

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/barbara"
	_ "github.com/genshinsim/gcsim/internal/characters/sucrose"
	_ "github.com/genshinsim/gcsim/internal/weapons/catalyst/hymnofthemaelstrom"
	_ "github.com/genshinsim/gcsim/internal/weapons/catalyst/thrilling"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func TestBurstHealingIsOneHymnStack(t *testing.T) {
	c, err := core.New(core.Opt{})
	if err != nil {
		t.Fatal(err)
	}
	for i, k := range []keys.Char{keys.SangonomiyaKokomi, keys.Barbara, keys.Sucrose} {
		weapon := keys.ThrillingTalesOfDragonSlayers
		if i == 0 {
			weapon = keys.HymnOfTheMaelstrom
		}
		p := testhelper.DefaultProfile(k, weapon)
		p.Weapon.Refine = 1
		if i == 0 {
			p.Base.Cons = 2
		}
		if _, err := c.AddChar(p); err != nil {
			t.Fatal(err)
		}
	}
	c.Player.SetActive(0)
	if err := c.Player.InitializeTeam(); err != nil {
		t.Fatal(err)
	}
	c.Player.ByIndex(1).SetHPByRatio(0.25)
	wielder := c.Player.ByIndex(0)
	baseHP := wielder.Stat(attributes.HPP)
	c.Status.Add("kokomiburst", 60)
	actor := wielder.Character.(*char)
	target := enemy.New(c, info.EnemyProfile{})
	for i := 1; i <= 2; i++ {
		actor.makeBurstHealCB()(info.AttackCB{Target: target})
		if got, want := wielder.Stat(attributes.HPP)-baseHP, 0.04*float64(i); math.Abs(got-want) > 1e-8 {
			t.Fatalf("healing episode %d: HP bonus = %v, want %v", i, got, want)
		}
	}
}
