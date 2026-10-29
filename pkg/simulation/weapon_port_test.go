package simulation

import (
	"math"
	"testing"

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
		{keys.BreezeborneRefrain, "breezebornerefrain", 15437, keys.Amber},
		{keys.HymnOfTheMaelstrom, "hymnofthemaelstrom", 14524, keys.Barbara},
		{keys.HereticsMoltenBlade, "hereticsmoltenblade", 11435, keys.Albedo},
		{keys.BeyondTheChrysalis, "beyondthechrysalis", 11522, keys.Albedo},
		{keys.NewBough, "newbough", 11437, keys.Albedo},
		{keys.SilverLight, "silverlight", 11438, keys.Albedo},
		{keys.WhitelakeFrostfeather, "whitelakefrostfeather", 11520, keys.Albedo},
		{keys.StarsealedEdict, "starsealededict", 14525, keys.Sucrose},
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

func TestSilverLightIndependentStacks(t *testing.T) {
	c, char := portWeaponCore(t, keys.SilverLight, keys.Albedo)
	base := char.Stat(attributes.EM)
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "one stack", char.Stat(attributes.EM)-base, 52)
	c.F = 11
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "stack ICD", char.Stat(attributes.EM)-base, 52)
	c.F = 12
	c.Events.Emit(event.OnSkill)
	wantFloat(t, "two stacks", char.Stat(attributes.EM)-base, 104)
	c.F = 720
	wantFloat(t, "first stack expires", char.Stat(attributes.EM)-base, 52)
	c.F = 732
	wantFloat(t, "both stacks expire", char.Stat(attributes.EM)-base, 0)
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

func TestNewBoughUsesNativeStellarSwirlStatus(t *testing.T) {
	c, char := portWeaponCore(t, keys.NewBough, keys.Albedo)
	baseEM := char.Stat(attributes.EM)
	c.Events.Emit(event.OnSkill)
	hit := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0}}
	c.Events.Emit(event.OnEnemyDamage, nil, hit)
	c.F = 59
	c.Events.Emit(event.OnEnemyDamage, nil, hit)
	wantFloat(t, "stack ICD", char.Stat(attributes.ATKP), 0.04)
	for _, frame := range []int{60, 120} {
		c.F = frame
		c.Events.Emit(event.OnEnemyDamage, nil, hit)
	}
	wantFloat(t, "normal ATK", char.Stat(attributes.ATKP), 0.12)
	wantFloat(t, "normal EM", char.Stat(attributes.EM)-baseEM, 60)
	char.AddStatus(reactable.SswKey, 480, false)
	wantFloat(t, "stellar ATK", char.Stat(attributes.ATKP), 0.18)
	wantFloat(t, "stellar EM", char.Stat(attributes.EM)-baseEM, 0)
	ai := info.AttackInfo{AttackTag: attacks.AttackTagReactionStellarSwirl}
	wantFloat(t, "stellar reaction", char.ReactBonus(ai), 0.24)
	c.F = 480
	wantFloat(t, "all stacks expired", char.Stat(attributes.ATKP), 0)
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
