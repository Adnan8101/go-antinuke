package detectors

import (
	"go-antinuke-2.0/internal/state"
)

type BanDetector struct{}

func NewBanDetector() *BanDetector {
	return &BanDetector{}
}

func (d *BanDetector) Detect(guildIndex, actorIndex uint32, timestamp int64, threshold uint32) (bool, uint32) {
	gs := state.GetGuildState()
	as := state.GetActorState()

	// Panic mode (threshold = 0): trigger on EVERY event
	if threshold == 0 {
		// ALWAYS trigger in panic mode - ban on every malicious action
		guildCount := gs.IncrementBans(guildIndex)
		as.IncrementBans(actorIndex)
		return true, guildCount
	}

	// Normal mode: check thresholds
	guildCount := gs.IncrementBans(guildIndex)
	actorCount := as.IncrementBans(actorIndex)

	guildTrigger := BranchlessGreaterEqual(guildCount, threshold)
	actorTrigger := BranchlessGreaterEqual(actorCount, threshold)

	triggered := guildTrigger | actorTrigger

	// CRITICAL: Set triggered flag immediately to prevent race conditions
	if triggered != 0 {
		as.SetTriggered(actorIndex, true)
	}

	if triggered != 0 {
		counters := gs.GetCounters(guildIndex)
		counters.LastBanTime = timestamp
	}

	return triggered != 0, guildCount
}

func (d *BanDetector) CheckVelocity(guildIndex uint32, currentTime int64, windowMs int64) bool {
	gs := state.GetGuildState()
	counters := gs.GetCounters(guildIndex)

	lastTime := counters.LastBanTime
	deltaMs := (currentTime - lastTime) / 1000000

	return deltaMs < windowMs
}

func BranchlessGreaterEqual(value, threshold uint32) uint32 {
	diff := int32(value - threshold)
	return uint32((diff>>31)+1) & 1
}
