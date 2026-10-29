package breezebornerefrain

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/amber"
	_ "github.com/genshinsim/gcsim/internal/characters/fischl"
	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/weapons/sword/dullblade"
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
		{keys.BreezeborneRefrain, "breezebornerefrain", 15437, keys.Amber},
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

func TestBreezeborneIndependentWielders(t *testing.T) {
	c, err := core.New(core.Opt{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		char   keys.Char
		weapon keys.Weapon
		refine int
	}{
		{keys.Amber, keys.BreezeborneRefrain, 5},
		{keys.Fischl, keys.BreezeborneRefrain, 1},
		{keys.Kaeya, keys.DullBlade, 1},
	} {
		profile := testhelper.DefaultProfile(tt.char, tt.weapon)
		profile.Weapon.Refine = tt.refine
		if _, err := c.AddChar(profile); err != nil {
			t.Fatal(err)
		}
	}
	c.Player.SetActive(0)
	if err := c.Player.InitializeTeam(); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		index int
		want  float64
	}{{0, 0.4}, {1, 0.2}} {
		ch := c.Player.ByIndex(tt.index)
		wantFloat(t, "personal ER", ch.Stat(attributes.ER)-ch.BaseStats[attributes.ER], tt.want)
	}
	hit := func(frame, actor int, tag attacks.AttackTag) {
		c.F = frame
		c.Events.Emit(event.OnEnemyHit, nil, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: actor, AttackTag: tag}})
	}
	check := func(frame int, want float64) {
		t.Helper()
		c.F = frame
		for i := 0; i < 3; i++ {
			for _, tag := range []attacks.AttackTag{attacks.AttackTagDirectStellarConduct, attacks.AttackTagDirectStellarSwirl, attacks.AttackTagReactionStellarSwirl} {
				wantFloat(t, "party stellar bonus", c.Player.ByIndex(i).ReactBonus(info.AttackInfo{AttackTag: tag}), want)
			}
			wantFloat(t, "non-stellar bonus", c.Player.ByIndex(i).ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagBloom}), 0)
		}
	}
	hit(0, 0, attacks.AttackTagElementalArt)
	hit(1, 0, attacks.AttackTagElementalArt)
	hit(2, 0, attacks.AttackTagElementalArtHold)
	check(3, 0)
	hit(4, 0, attacks.AttackTagElementalBurst)
	check(4, 0.48)
	for _, frame := range []int{100, 102, 104} {
		hit(frame, 1, attacks.AttackTagElementalArt)
	}
	check(104, 0.48)
	for _, frame := range []int{200, 202, 204} {
		hit(frame, 0, attacks.AttackTagElementalArt)
	}
	check(723, 0.48)
	check(724, 0.24)
	check(823, 0.24)
	check(824, 0)
	hit(824, 0, attacks.AttackTagElementalArt)
	hit(826, 0, attacks.AttackTagElementalArt)
	check(827, 0)
	hit(828, 0, attacks.AttackTagElementalArt)
	check(828, 0.48)
	check(1548, 0)
	for _, frame := range []int{1600, 1602, 1604} {
		hit(frame, 1, attacks.AttackTagElementalArt)
	}
	check(1604, 0.24)
	for _, frame := range []int{1700, 1702, 1704} {
		hit(frame, 0, attacks.AttackTagElementalBurst)
	}
	check(1704, 0.48)
	check(2324, 0.48)
	check(2424, 0)
}
