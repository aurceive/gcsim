package reactable

// Verdant Dew is tracked on Player; Lunar Bloom triggers the tick window.
const VerdantDewKey = "verdant-dew"

func (r *Reactable) LunarBloom() {
	r.core.Player.OnLunarBloom()
}
