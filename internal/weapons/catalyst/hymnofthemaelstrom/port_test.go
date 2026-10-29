package hymnofthemaelstrom

import (
	"fmt"
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/barbara"
	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/characters/kokomi"
	_ "github.com/genshinsim/gcsim/internal/weapons/catalyst/thrilling"
	_ "github.com/genshinsim/gcsim/internal/weapons/sword/dullblade"
	"github.com/genshinsim/gcsim/pkg/catalog"
	"github.com/genshinsim/gcsim/pkg/core"
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
		{keys.HymnOfTheMaelstrom, "hymnofthemaelstrom", 14524, keys.Barbara},
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
		{keys.HymnOfTheMaelstrom, 44.336, 0.144, 44.336*8.01 + 186.7, 0.144 * 4.594, attributes.HPP},
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

func hymnTeam(t *testing.T, secondWeapon keys.Weapon, secondRefine int) *core.Core {
	t.Helper()
	c, err := core.New(core.Opt{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		char   keys.Char
		weapon keys.Weapon
		refine int
	}{
		{keys.Barbara, keys.HymnOfTheMaelstrom, 1},
		{keys.SangonomiyaKokomi, secondWeapon, secondRefine},
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
	for _, i := range []int{0, 1} {
		ch := c.Player.ByIndex(i)
		ch.BaseStats[attributes.BaseHP] = 60000
		ch.BaseStats[attributes.HPP] = 0
		ch.BaseStats[attributes.HP] = 0
	}
	return c
}

func TestHymnRefinementScaling(t *testing.T) {
	for r := 1; r <= 5; r++ {
		t.Run(fmt.Sprint(r), func(t *testing.T) {
			c, err := core.New(core.Opt{})
			if err != nil {
				t.Fatal(err)
			}
			p := testhelper.DefaultProfile(keys.Barbara, keys.HymnOfTheMaelstrom)
			p.Weapon.Refine = r
			if _, err := c.AddChar(p); err != nil {
				t.Fatal(err)
			}
			c.Player.SetActive(0)
			if err := c.Player.InitializeTeam(); err != nil {
				t.Fatal(err)
			}
			ch := c.Player.ByIndex(0)
			ch.BaseStats[attributes.BaseHP] = 50000
			ch.BaseStats[attributes.HPP] = 0
			ch.BaseStats[attributes.HP] = 0
			baseHeal := ch.BaseStats[attributes.Heal]
			baseATK := ch.Stat(attributes.ATKP)
			gain := 0.03 + 0.01*float64(r)
			wantFloat(t, "healing bonus", ch.Stat(attributes.Heal)-baseHeal, gain)
			c.Player.Heal(info.HealInfo{Caller: 0, Target: -1, Src: 1})
			wantFloat(t, "HP bonus", ch.Stat(attributes.HPP), gain)
			wantFloat(t, "attack per 1000 HP", ch.Stat(attributes.ATKP)-baseATK, (50*(1+gain)-40)*(0.003+0.001*float64(r)))
		})
	}
}

func TestHymnPartyHealExpiryAndReaction(t *testing.T) {
	c := hymnTeam(t, keys.ThrillingTalesOfDragonSlayers, 1)
	wielder := c.Player.ByIndex(0)
	recipient := c.Player.ByIndex(2)
	baseATK := recipient.Stat(attributes.ATKP)
	c.Player.SetActive(2)
	heal := info.HealInfo{Caller: 0, Target: -1, Src: 1}
	c.Player.Heal(heal)
	wantFloat(t, "one party-wide heal is one stack", wielder.Stat(attributes.HPP), 0.04)
	wantFloat(t, "one capped stack", recipient.Stat(attributes.ATKP)-baseATK, 0.08)
	c.Player.Heal(heal)
	c.Player.Heal(heal)
	wantFloat(t, "separate same-frame heals stack", wielder.Stat(attributes.HPP), 0.12)
	wantFloat(t, "per-stack cap", recipient.Stat(attributes.ATKP)-baseATK, 0.24)
	invalid := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: -1}}
	c.Events.Emit(event.OnFrozen, nil, invalid)
	wantFloat(t, "non-party reaction ignored", wielder.Stat(attributes.HPP), 0.12)
	c.F = 10
	valid := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 2}}
	c.Events.Emit(event.OnStellarSwirl, nil, valid)
	wantFloat(t, "boosted HP", wielder.Stat(attributes.HPP), 0.21)
	wantFloat(t, "boost after per-stack cap", recipient.Stat(attributes.ATKP)-baseATK, 0.42)
	c.F = 310
	wantFloat(t, "reaction boost expired", wielder.Stat(attributes.HPP), 0.12)
	wantFloat(t, "reaction attack boost expired", recipient.Stat(attributes.ATKP)-baseATK, 0.24)
	c.Events.Emit(event.OnFrozen, nil, valid)
	wantFloat(t, "frozen also boosts HP", wielder.Stat(attributes.HPP), 0.21)
	c.F = 599
	wantFloat(t, "HP before expiry", wielder.Stat(attributes.HPP), 0.21)
	c.F = 600
	wantFloat(t, "HP expired", wielder.Stat(attributes.HPP), 0)
	wantFloat(t, "attack expired", recipient.Stat(attributes.ATKP)-baseATK, 0)
	c.Player.Heal(heal)
	wantFloat(t, "expired stacks reset with reaction boost", wielder.Stat(attributes.HPP), 0.07)
	c.F = 610
	wantFloat(t, "new stack after reaction expires", wielder.Stat(attributes.HPP), 0.04)
}

func TestHymnMultipleWieldersUseHighestActiveBonus(t *testing.T) {
	c := hymnTeam(t, keys.HymnOfTheMaelstrom, 5)
	c.Player.SetActive(2)
	recipient := c.Player.ByIndex(2)
	baseATK := recipient.Stat(attributes.ATKP)
	c.Player.Heal(info.HealInfo{Caller: 1, Target: -1, Src: 1})
	wantFloat(t, "R5 one stack", recipient.Stat(attributes.ATKP)-baseATK, 0.16)
	c.F = 100
	for range 3 {
		c.Player.Heal(info.HealInfo{Caller: 0, Target: -1, Src: 1})
	}
	wantFloat(t, "higher R1 total, not sum", recipient.Stat(attributes.ATKP)-baseATK, 0.24)
	wantFloat(t, "R1 holder HP", c.Player.ByIndex(0).Stat(attributes.HPP), 0.12)
	wantFloat(t, "R5 holder HP", c.Player.ByIndex(1).Stat(attributes.HPP), 0.08)
	c.Player.SetActive(0)
	wantFloat(t, "inactive recipient", recipient.Stat(attributes.ATKP)-baseATK, 0)
	c.Player.SetActive(2)
	c.F = 600
	wantFloat(t, "R1 remains after R5 expires", recipient.Stat(attributes.ATKP)-baseATK, 0.24)
	wantFloat(t, "R5 HP expires independently", c.Player.ByIndex(1).Stat(attributes.HPP), 0)
	wantFloat(t, "R1 HP still active", c.Player.ByIndex(0).Stat(attributes.HPP), 0.12)
	c.F = 700
	wantFloat(t, "both expired", recipient.Stat(attributes.ATKP)-baseATK, 0)
}

func TestHymnGroupedPerTargetHeals(t *testing.T) {
	c := hymnTeam(t, keys.HymnOfTheMaelstrom, 5)
	wielder := c.Player.ByIndex(1)
	group := &info.HealGroup{}
	c.Player.Heal(info.HealInfo{Caller: 1, Target: 0, Group: group, Src: 1})
	c.Player.Heal(info.HealInfo{Caller: 1, Target: 2, Group: group, Src: 1})
	wantFloat(t, "one group is one stack", wielder.Stat(attributes.HPP), 0.08)
	otherGroup := &info.HealGroup{}
	c.Player.Heal(info.HealInfo{Caller: 1, Target: 0, Group: otherGroup, Src: 1})
	c.Player.Heal(info.HealInfo{Caller: 1, Target: 2, Group: otherGroup, Src: 1})
	wantFloat(t, "separate same-frame group", wielder.Stat(attributes.HPP), 0.16)
	c.Player.Heal(info.HealInfo{Caller: 1, Target: 1, Src: 1})
	wantFloat(t, "separate single-target heal", wielder.Stat(attributes.HPP), 0.24)
}
