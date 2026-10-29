package wintersheavyheart

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/characters/lisa"
	_ "github.com/genshinsim/gcsim/internal/characters/sucrose"
	_ "github.com/genshinsim/gcsim/internal/weapons/sword/dullblade"
	"github.com/genshinsim/gcsim/pkg/catalog"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/reactable"
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
		{keys.WintersHeavyHeart, "wintersheavyheart", 14437, keys.Sucrose},
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

func TestWintersHeavyHeartUsesNativeRadiance(t *testing.T) {
	c, char := portWeaponCore(t, keys.WintersHeavyHeart, keys.Lisa)
	profile := testhelper.DefaultProfile(keys.Kaeya, keys.DullBlade)
	profile.Weapon.Refine = 1
	if _, err := c.AddChar(profile); err != nil {
		t.Fatal(err)
	}
	baseEM := char.BaseStats[attributes.EM]
	wantFloat(t, "normal EM", char.Stat(attributes.EM)-baseEM, 24)
	wantFloat(t, "normal ATK", char.Stat(attributes.ATKP), 0.048)
	char.AddStatus(reactable.SswKey, 480, false)
	wantFloat(t, "stellar EM", char.Stat(attributes.EM)-baseEM, 40)
	wantFloat(t, "stellar ATK", char.Stat(attributes.ATKP), 0)
	ai := info.AttackInfo{AttackTag: attacks.AttackTagDirectStellarConduct}
	wantFloat(t, "stellar bonus", char.ReactBonus(ai), 0.12)
	char.DeleteStatus(reactable.SswKey)
	char.AddStatus(reactable.PolestarFieldKey, 480, false)
	wantFloat(t, "polestar EM", char.Stat(attributes.EM)-baseEM, 40)
}
