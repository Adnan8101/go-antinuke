package detectors

import (
	"fmt"

	"go-antinuke-2.0/internal/state"
)

type ChannelDeleteDetector struct{}

func NewChannelDeleteDetector() *ChannelDeleteDetector {
	return &ChannelDeleteDetector{}
}

func (d *ChannelDeleteDetector) Detect(guildIndex, actorIndex uint32, timestamp int64, threshold uint32) (bool, uint32) {
	gs := state.GetGuildState()
	as := state.GetActorState()

	fmt.Printf("[DETECTOR] Channel - threshold=%d, actorIndex=%d\n", threshold, actorIndex)

	// Panic mode (threshold = 0): trigger on EVERY event
	if threshold == 0 {
		// ALWAYS trigger in panic mode - ban on every malicious action
		gs.IncrementChannelDeletes(guildIndex)
		as.IncrementChannelDeletes(actorIndex)
		guildCount := gs.GetChannelDeleteCount(guildIndex)
		fmt.Printf("[DETECTOR] PANIC MODE TRIGGER! Actor %d, returning TRUE\n", actorIndex)
		return true, guildCount
	}

	// Normal mode: increment first, then check thresholds
	// This way threshold=1 means "ban after 1 action"
	guildCount := gs.IncrementChannelDeletes(guildIndex)
	actorCount := as.IncrementChannelDeletes(actorIndex)

	guildTrigger := BranchlessGreaterEqual(guildCount, threshold)
	actorTrigger := BranchlessGreaterEqual(actorCount, threshold)

	triggered := guildTrigger | actorTrigger

	// CRITICAL: Set triggered flag immediately to prevent race conditions
	// This stops rapid subsequent events from being processed
	if triggered != 0 {
		as.SetTriggered(actorIndex, true)
	}

	if triggered != 0 {
		counters := gs.GetCounters(guildIndex)
		counters.LastChanTime = timestamp
	}

	return triggered != 0, guildCount
}

func (d *ChannelDeleteDetector) CheckVelocity(guildIndex uint32, currentTime int64, windowMs int64) bool {
	gs := state.GetGuildState()
	counters := gs.GetCounters(guildIndex)

	lastTime := counters.LastChanTime
	deltaMs := (currentTime - lastTime) / 1000000

	return deltaMs < windowMs
}
