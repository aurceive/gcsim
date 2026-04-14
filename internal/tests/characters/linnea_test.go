package characters

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/amber"
	_ "github.com/genshinsim/gcsim/internal/characters/barbara"
	_ "github.com/genshinsim/gcsim/internal/characters/linnea"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
)

func TestLinneaC2BuffSkipsNonGeoHydroTeammates(t *testing.T) {
	c, trg := makeCore(1)

	amberIdx, err := c.AddChar(defProfile(keys.Amber))
	if err != nil {
		t.Fatalf("error adding amber: %v", err)
	}

	linneaProf := defProfile(keys.Linnea)
	linneaProf.Base.Cons = 2
	linneaIdx, err := c.AddChar(linneaProf)
	if err != nil {
		t.Fatalf("error adding linnea: %v", err)
	}

	barbaraIdx, err := c.AddChar(defProfile(keys.Barbara))
	if err != nil {
		t.Fatalf("error adding barbara: %v", err)
	}

	c.Player.SetActive(linneaIdx)
	if err := c.Init(); err != nil {
		t.Fatalf("error initializing core: %v", err)
	}

	beforeAmber := c.Player.ByIndex(amberIdx).Stat(attributes.CD)
	beforeLinnea := c.Player.ByIndex(linneaIdx).Stat(attributes.CD)
	beforeBarbara := c.Player.ByIndex(barbaraIdx).Stat(attributes.CD)

	c.Events.Emit(event.OnMoondriftHarmony, trg[0], &info.AttackEvent{})

	afterAmber := c.Player.ByIndex(amberIdx).Stat(attributes.CD)
	afterLinnea := c.Player.ByIndex(linneaIdx).Stat(attributes.CD)
	afterBarbara := c.Player.ByIndex(barbaraIdx).Stat(attributes.CD)

	if math.Abs(afterAmber-beforeAmber) > 1e-9 {
		t.Fatalf("expected amber crit damage to remain unchanged, before=%v after=%v", beforeAmber, afterAmber)
	}
	if math.Abs((afterLinnea-beforeLinnea)-0.4) > 1e-9 {
		t.Fatalf("expected linnea to gain 0.4 crit damage from c2, before=%v after=%v", beforeLinnea, afterLinnea)
	}
	if math.Abs((afterBarbara-beforeBarbara)-0.4) > 1e-9 {
		t.Fatalf("expected barbara to gain 0.4 crit damage from c2, before=%v after=%v", beforeBarbara, afterBarbara)
	}
}