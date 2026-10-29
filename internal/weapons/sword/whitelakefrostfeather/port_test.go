package whitelakefrostfeather

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
		{keys.WhitelakeFrostfeather, "whitelakefrostfeather", 11520, keys.Albedo},
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

func TestWhitelakeRequiresThreeStacks(t *testing.T) {
	c, char := portWeaponCore(t, keys.WhitelakeFrostfeather, keys.Albedo)
	char.EnergyMax = 100
	char.Energy = 0
	hit := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, AttackTag: attacks.AttackTagElementalArt}}
	reaction := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, AttackTag: attacks.AttackTagDirectStellarConduct}}
	c.Events.Emit(event.OnStellarConduct, nil, reaction)
	wantFloat(t, "energy without stacks", char.Energy, 0)
	c.Events.Emit(event.OnEnemyHit, nil, hit)
	c.F = 5
	c.Events.Emit(event.OnEnemyHit, nil, hit)
	wantFloat(t, "stack ICD", char.Stat(attributes.ATKP), 0.08)
	for _, frame := range []int{6, 12} {
		c.F = frame
		c.Events.Emit(event.OnEnemyHit, nil, hit)
	}
	wantFloat(t, "three ATK stacks", char.Stat(attributes.ATKP), 0.24)
	c.Events.Emit(event.OnStellarConduct, nil, reaction)
	wantFloat(t, "energy with stacks", char.Energy, 4)
	c.Events.Emit(event.OnStellarConduct, nil, reaction)
	wantFloat(t, "energy ICD", char.Energy, 4)
	other := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1, AttackTag: attacks.AttackTagReactionStellarSwirl}}
	c.Events.Emit(event.OnSpecialReactionAttack, nil, other)
	wantFloat(t, "other contributor crit", other.Snapshot.Stats[attributes.CD], 0)
	own := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, AttackTag: attacks.AttackTagReactionStellarSwirl}}
	c.Events.Emit(event.OnSpecialReactionAttack, nil, own)
	wantFloat(t, "own contributor crit", own.Snapshot.Stats[attributes.CD], 0.5)
	c.F = 480
	c.Events.Emit(event.OnStellarConduct, nil, reaction)
	wantFloat(t, "energy after a stack expires", char.Energy, 4)
}
