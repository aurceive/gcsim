package info

type HealInfo struct {
	Caller  int        // index of healing character
	Target  int        // index of char receiving the healing. use -1 to heal all characters
	Group   *HealGroup // optional group for simultaneous per-target heals in one episode
	Type    HealType
	Message string
	Src     float64 // depends on the type
	Bonus   float64
}

// HealGroup identifies simultaneous per-target heals belonging to one episode.
type HealGroup struct{ _ byte }

type HealType int

const (
	HealTypeAbsolute HealType = iota // regular number
	HealTypePercent                  // percent of the target's max hp
)

type DrainInfo struct {
	ActorIndex int
	Abil       string
	Amount     float64
	External   bool
}
