package main

import (
	"os"
	"testing"

	"github.com/genshinsim/gcsim/pkg/catalog"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	excel "github.com/shizukayuki/excel-hk4e"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/proto"
)

func TestLocalWeaponFallback(t *testing.T) {
	root, err := os.OpenRoot("..")
	if err != nil {
		t.Fatal(err)
	}
	oldRoot := projectRoot
	oldWeapons := excel.WeaponExcelConfigData
	projectRoot = root
	t.Cleanup(func() {
		projectRoot = oldRoot
		excel.WeaponExcelConfigData = oldWeapons
		root.Close()
	})

	for _, tt := range []struct {
		path, key, name string
		russian         string
		id              uint32
		weapon          keys.Weapon
	}{
		{"internal/weapons/bow/breezebornerefrain/config.yml", "breezebornerefrain", "Breezeborne Refrain", "Напевы нежного ветра", 15437, keys.BreezeborneRefrain},
		{"internal/weapons/catalyst/starsealededict/config.yml", "starsealededict", "Starsealed Edict", "", 14525, keys.StarsealedEdict},
		{"internal/weapons/catalyst/wintersheavyheart/config.yml", "wintersheavyheart", "Winter's Heavy Heart", "", 14437, keys.WintersHeavyHeart},
		{"internal/weapons/sword/newbough/config.yml", "newbough", "New Bough", "", 11437, keys.NewBough},
		{"internal/weapons/sword/beyondthechrysalis/config.yml", "beyondthechrysalis", "Beyond the Chrysalis", "За пределами кокона", 11522, keys.BeyondTheChrysalis},
		{"internal/weapons/sword/silverlight/config.yml", "silverlight", "Silver Light", "Серебряный свет", 11438, keys.SilverLight},
		{"internal/weapons/sword/whitelakefrostfeather/config.yml", "whitelakefrostfeather", "Whitelake Frostfeather", "Морозное перо белого озера", 11520, keys.WhitelakeFrostfeather},
	} {
		t.Run(tt.key, func(t *testing.T) {
			data, err := projectRoot.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			cfg := &Config{}
			if err := yaml.Load(data, cfg, yaml.WithV4Defaults(), yaml.WithKnownFields()); err != nil {
				t.Fatal(err)
			}
			cfg.Path = tt.path
			compiled := &Compiled{
				Localization: map[string]Localization{
					"English": {WeaponNames: make(map[string]string)},
					"Russian": {WeaponNames: make(map[string]string)},
				},
			}
			excel.WeaponExcelConfigData = nil
			if err := compiled.build(cfg); err != nil {
				t.Fatal(err)
			}
			if cfg.Weapon.ref != nil {
				t.Fatal("expected local fallback")
			}
			if !proto.Equal(cfg.Weapon.Model, catalog.WeaponMap[tt.weapon]) {
				t.Fatalf("local and runtime weapon data differ: %v", cfg.Weapon.Model)
			}
			if got := compiled.Localization["English"].WeaponNames[cfg.Name]; got != tt.name {
				t.Fatalf("English name = %q", got)
			}
			if got := compiled.Localization["Russian"].WeaponNames[cfg.Name]; got != tt.russian {
				t.Fatalf("Russian name = %q, want %q", got, tt.russian)
			}
			excel.WeaponExcelConfigData = []*excel.Weapon{{Item: excel.Item{Id: tt.id}}}
			if spec, err := buildWeaponSpec(cfg); err != nil || spec.ref != nil {
				t.Fatalf("expected fallback for missing upstream name: %v, %v", spec, err)
			}
		})
	}

	cfg := &Config{
		Path:        "internal/weapons/catalyst/starsealededict/config.yml",
		Name:        "starsealededict",
		Override:    Override{Id: 14525},
		LocalWeapon: &LocalWeapon{Name: "Starsealed Edict", Data: "data.textproto"},
	}
	cfg.Override.Id = 1
	if _, err := buildLocalWeaponSpec(cfg); err == nil {
		t.Fatal("expected an error for a mismatched weapon ID")
	}
	cfg.Override.Id = 14525
	cfg.LocalWeapon.Data = "missing.textproto"
	if _, err := buildLocalWeaponSpec(cfg); err == nil {
		t.Fatal("expected an error for a missing data file")
	}
}
