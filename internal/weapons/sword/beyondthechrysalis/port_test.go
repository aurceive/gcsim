package beyondthechrysalis

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/albedo"
	"github.com/genshinsim/gcsim/pkg/catalog"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/shortcut"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func portWeaponCore(t *testing.T, weapon keys.Weapon, charKey keys.Char) (*core.Core, *character.CharWrapper) {
	t.Helper()
	c, err := core.New(core.Opt{})
	if err != nil {
		t.Fatal(err)
	}
	profile := testhelper.DefaultProfile(charKey, weapon)
	profile.Weapon.Level = 90
	profile.Weapon.MaxLevel = 90
	profile.Weapon.Refine = 1
	if _, err := c.AddChar(profile); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(0)
	return c, c.Player.ByIndex(0)
}

func wantFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-8 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func TestPortWeaponRegistrationAndStats(t *testing.T) {
	for _, tt := range []struct {
		key  keys.Weapon
		name string
		id   uint32
		char keys.Char
	}{
		{keys.BeyondTheChrysalis, "beyondthechrysalis", 11522, keys.Albedo},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := keys.WeaponString(tt.name); err != nil || got != tt.key {
				t.Fatalf("WeaponString(%q) = %v, %v", tt.name, got, err)
			}
			if shortcut.WeaponNameToKey[tt.name] != tt.key {
				t.Fatalf("missing shortcut for %s", tt.name)
			}
			data := catalog.WeaponMap[tt.key]
			if data == nil || data.Id != tt.id || data.Key != tt.name {
				t.Fatalf("unexpected weapon data: %+v", data)
			}
			stats, err := character.WeaponBaseStat(info.WeaponProfile{Level: 90, MaxLevel: 90}, data)
			if err != nil || stats[attributes.BaseATK] <= 0 {
				t.Fatalf("invalid level 90 stats: %v, %v", stats, err)
			}
			_, char := portWeaponCore(t, tt.key, tt.char)
			if err := char.UpdateBaseStats(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBeyondTheChrysalisSequenceAndCooldown(t *testing.T) {
	c, char := portWeaponCore(t, keys.BeyondTheChrysalis, keys.Albedo)
	char.EnergyMax = 100
	char.Energy = 0
	base := char.Stat(attributes.CD)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "crit bonus", char.Stat(attributes.CD)-base, 0.56)
	c.Events.Emit(event.OnBurst)
	ai := info.AttackInfo{AttackTag: attacks.AttackTagReactionStellarSwirl}
	wantFloat(t, "stellar swirl bonus", char.ReactBonus(ai), 0.36)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "energy", char.Energy, 5)
	c.Events.Emit(event.OnCharacterSwap, 0, 1)
	wantFloat(t, "swap removes crit bonus", char.Stat(attributes.CD)-base, 0)
	wantFloat(t, "swap removes reaction bonus", char.ReactBonus(ai), 0)
	c.Events.Emit(event.OnSkill)
	c.Events.Emit(event.OnSkill)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "swap does not reset energy ICD", char.Energy, 5)
	c.F = 239
	c.Events.Emit(event.OnSkill)
	c.Events.Emit(event.OnSkill)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "energy before ICD expiry", char.Energy, 5)
	c.F = 240
	c.Events.Emit(event.OnSkill)
	c.Events.Emit(event.OnSkill)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "energy after ICD", char.Energy, 10)
}
