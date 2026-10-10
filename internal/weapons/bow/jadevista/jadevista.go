package jadevista

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.JadeVista, NewWeapon)
}

type Weapon struct {
	Index  int
	refine int
	char   *character.CharWrapper
	core   *core.Core
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	sameElement, differentElement := 0, 0
	for _, member := range w.core.Player.Chars() {
		if member.Index() == w.char.Index() {
			continue
		}
		if member.Base.Element == w.char.Base.Element {
			sameElement++
		} else {
			differentElement++
		}
	}

	sameStacks, differentStacks := stackCounts(sameElement, differentElement)
	bonus := make([]float64, attributes.EndStatType)
	emPerStack, atkPerStack := stackBonuses(w.refine)
	bonus[attributes.EM] = emPerStack * float64(sameStacks)
	bonus[attributes.ATKP] = atkPerStack * float64(differentStacks)

	w.char.AddStatMod(character.StatMod{
		Base: modifier.NewBase("jade-vista", -1),
		Amount: func() []float64 {
			return bonus
		},
	})

	return nil
}

func stackBonuses(refine int) (float64, float64) {
	return float64(48 + 16*refine), 0.09 + 0.03*float64(refine)
}

func stackCounts(sameElement, differentElement int) (int, int) {
	sameStacks := min(sameElement, 3)
	differentStacks := min(differentElement, 3-sameStacks)
	return sameStacks, differentStacks
}

// Jade Vista grants 64-128 EM per same-element party member and 12%-24% ATK
// per different-element party member, up to 3 total stacks. EM stacks take priority.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	return &Weapon{
		refine: p.Refine,
		char:   char,
		core:   c,
	}, nil
}
