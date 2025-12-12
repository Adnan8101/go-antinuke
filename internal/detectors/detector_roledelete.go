package detectors

import (
	"go-antinuke-2.0/internal/state"
)

type RoleDeleteDetector struct{}

func NewRoleDeleteDetector() *RoleDeleteDetector {
	return &RoleDeleteDetector{}
}

func (d *RoleDeleteDetector) Detect(guildIndex, actorIndex uint32, timestamp int64, threshold uint32) (bool, uint32) {
	gs := state.GetGuildState()
	as := state.GetActorState()

	// Panic mode (threshold = 0): trigger on EVERY event
	if threshold == 0 {
		// ALWAYS trigger in panic mode - ban on every malicious action
		guildCount := gs.IncrementRoleDeletes(guildIndex)
		as.IncrementRoleDeletes(actorIndex)
		return true, guildCount
	}

	// Normal mode: check thresholds
	guildCount := gs.IncrementRoleDeletes(guildIndex)
	actorCount := as.IncrementRoleDeletes(actorIndex)

	guildTrigger := BranchlessGreaterEqual(guildCount, threshold)
	actorTrigger := BranchlessGreaterEqual(actorCount, threshold)

	triggered := guildTrigger | actorTrigger

	// CRITICAL: Set triggered flag immediately to prevent race conditions
	if triggered != 0 {
		as.SetTriggered(actorIndex, true)
	}

	if triggered != 0 {
		counters := gs.GetCounters(guildIndex)
		counters.LastRoleTime = timestamp
	}

	return triggered != 0, guildCount
}

func (d *RoleDeleteDetector) CheckVelocity(guildIndex uint32, currentTime int64, windowMs int64) bool {
	gs := state.GetGuildState()
	counters := gs.GetCounters(guildIndex)

	lastTime := counters.LastRoleTime
	deltaMs := (currentTime - lastTime) / 1000000

	return deltaMs < windowMs
}
