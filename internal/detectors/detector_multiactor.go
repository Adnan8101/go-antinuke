package detectors

import (
	"go-antinuke-2.0/internal/state"
)

type MultiActorDetector struct{}

func NewMultiActorDetector() *MultiActorDetector {
	return &MultiActorDetector{}
}

func (d *MultiActorDetector) Detect(guildIndex uint32, threshold uint32) (bool, uint32) {
	hs := state.GetHazardScores()
	entry := hs.GetEntry(guildIndex)

	actorCount := entry.ActorCount
	destructiveOps := entry.DestructiveOps

	score := state.CalculateHazardScore(actorCount, destructiveOps, 0)
	hs.UpdateScore(guildIndex, score)

	triggered := BranchlessGreaterEqual(score, threshold)

	return triggered != 0, score
}

func (d *MultiActorDetector) IncrementActor(guildIndex uint32) uint32 {
	hs := state.GetHazardScores()
	return hs.IncrementActorCount(guildIndex)
}

func (d *MultiActorDetector) IncrementDestructive(guildIndex uint32) uint32 {
	hs := state.GetHazardScores()
	return hs.IncrementDestructive(guildIndex)
}

func (d *MultiActorDetector) GetScore(guildIndex uint32) uint32 {
	hs := state.GetHazardScores()
	return hs.GetScore(guildIndex)
}

func (d *MultiActorDetector) Reset(guildIndex uint32) {
	hs := state.GetHazardScores()
	hs.ResetEntry(guildIndex)
}
