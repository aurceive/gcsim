package starsealededict

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/sucrose"
	"github.com/genshinsim/gcsim/pkg/catalog"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
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
		{keys.StarsealedEdict, "starsealededict", 14525, keys.Sucrose},
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

func TestStarsealedEdictStacksAndRadiance(t *testing.T) {
	c, char := portWeaponCore(t, keys.StarsealedEdict, keys.Sucrose)
	baseEM := char.Stat(attributes.EM)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "base EM", char.Stat(attributes.EM)-baseEM, 48)
	target := enemy.New(c, info.EnemyProfile{})
	hit := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0}}
	c.Events.Emit(event.OnShatter, target, hit)
	wantFloat(t, "shatter does not add a stack", char.Stat(attributes.EM)-baseEM, 48)
	for range 4 {
		c.Events.Emit(event.OnVaporize, target, hit)
	}
	wantFloat(t, "three reaction stacks", char.Stat(attributes.EM)-baseEM, 120)
	atk := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, AttackTag: attacks.AttackTagDirectStellarConduct}}
	char.ApplyAttackMods(atk, target)
	wantFloat(t, "no crit without radiance", atk.Snapshot.Stats[attributes.CD], 0)
	char.AddStatus(reactable.SswKey, 480, false)
	char.ApplyAttackMods(atk, target)
	wantFloat(t, "stellar conduct crit damage", atk.Snapshot.Stats[attributes.CD], 0.36)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "skill resets stacks", char.Stat(attributes.EM)-baseEM, 48)
	c.F = 1200
	wantFloat(t, "EM after expiry", char.Stat(attributes.EM)-baseEM, 0)
}
