package hereticsmoltenblade

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/albedo"
	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/weapons/sword/dullblade"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/catalog"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
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
		{keys.HereticsMoltenBlade, "hereticsmoltenblade", 11435, keys.Albedo},
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

func TestAdditionalWeaponBaseStats(t *testing.T) {
	for _, tt := range []struct {
		key               keys.Weapon
		baseATK, subStat  float64
		level90ATK, sub90 float64
		stat              attributes.Stat
	}{
		{keys.HereticsMoltenBlade, 42.401, 0.06, 42.401*8.349 + 155.6, 0.06 * 4.594, attributes.CR},
	} {
		t.Run(tt.key.String(), func(t *testing.T) {
			one, err := character.WeaponBaseStat(info.WeaponProfile{Level: 1, MaxLevel: 20}, catalog.WeaponMap[tt.key])
			if err != nil {
				t.Fatal(err)
			}
			ninety, err := character.WeaponBaseStat(info.WeaponProfile{Level: 90, MaxLevel: 90}, catalog.WeaponMap[tt.key])
			if err != nil {
				t.Fatal(err)
			}
			wantFloat(t, "level 1 attack", one[attributes.BaseATK], tt.baseATK)
			wantFloat(t, "level 1 secondary stat", one[tt.stat], tt.subStat)
			wantFloat(t, "level 90 attack", ninety[attributes.BaseATK], tt.level90ATK)
			wantFloat(t, "level 90 secondary stat", ninety[tt.stat], tt.sub90)
		})
	}
}

func TestHereticsMovementSourcesAndCooldown(t *testing.T) {
	c, wielder := portWeaponCore(t, keys.HereticsMoltenBlade, keys.Albedo)
	p := testhelper.DefaultProfile(keys.Kaeya, keys.DullBlade)
	if _, err := c.AddChar(p); err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 0.5))
	base := wielder.Stat(attributes.ATKP)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "initial minimum", wielder.Stat(attributes.ATKP)-base, 0.18)
	c.Combat.SetPlayerPos(info.Point{X: 3.5})
	c.Events.Emit(event.OnMovement, 3.5)
	for frame := 1; frame <= 60; frame++ {
		c.F = frame
		c.Events.Emit(event.OnTick)
	}
	wantFloat(t, "movement sources add up", wielder.Stat(attributes.ATKP)-base, 0.36)
	for frame := 61; frame <= 120; frame++ {
		c.F = frame
		c.Events.Emit(event.OnTick)
	}
	wantFloat(t, "no movement returns to minimum", wielder.Stat(attributes.ATKP)-base, 0.18)
	c.Combat.SetPlayerPos(info.Point{X: 7})
	for frame := 121; frame <= 180; frame++ {
		c.F = frame
		c.Events.Emit(event.OnTick)
	}
	wantFloat(t, "coordinate-only movement", wielder.Stat(attributes.ATKP)-base, 0.27)
	if err := c.Player.Exec(action.ActionWalk, keys.Albedo, map[string]int{info.ParamMovement: 7}); err != nil {
		t.Fatal(err)
	}
	for frame := 181; frame <= 240; frame++ {
		c.F = frame
		c.Events.Emit(event.OnTick)
	}
	wantFloat(t, "action movement parameter", wielder.Stat(attributes.ATKP)-base, 0.36)
	c.Events.Emit(event.OnCharacterSwap, 0, 1)
	c.Player.SetActive(1)
	wantFloat(t, "swap removes effect", wielder.Stat(attributes.ATKP)-base, 0)
	c.Player.SetActive(0)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "cooldown survives swap", wielder.Stat(attributes.ATKP)-base, 0)
	c.F = 840
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "skill available at cooldown expiry", wielder.Stat(attributes.ATKP)-base, 0.18)
	c.F = 1680
	wantFloat(t, "effect expires", wielder.Stat(attributes.ATKP)-base, 0)
}
