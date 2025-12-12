package state

import (
	"sync/atomic"
)

const (
	MaxHazardEntries = 4096
	HazardMask       = MaxHazardEntries - 1
)

type HazardEntry struct {
	GuildID        uint64
	ActorCount     uint32
	DestructiveOps uint32
	TimeWindow     int64
	Score          uint32
	_              [12]byte
}

type HazardScores struct {
	entries [MaxHazardEntries]HazardEntry
}

var globalHazardScores *HazardScores

func InitHazardScores() {
	globalHazardScores = &HazardScores{}
}

func GetHazardScores() *HazardScores {
	return globalHazardScores
}

func (h *HazardScores) GetEntry(guildIndex uint32) *HazardEntry {
	return &h.entries[guildIndex&HazardMask]
}

func (h *HazardScores) IncrementActorCount(guildIndex uint32) uint32 {
	return atomic.AddUint32(&h.entries[guildIndex&HazardMask].ActorCount, 1)
}

func (h *HazardScores) IncrementDestructive(guildIndex uint32) uint32 {
	return atomic.AddUint32(&h.entries[guildIndex&HazardMask].DestructiveOps, 1)
}

func (h *HazardScores) UpdateScore(guildIndex, score uint32) {
	atomic.StoreUint32(&h.entries[guildIndex&HazardMask].Score, score)
}

func (h *HazardScores) GetScore(guildIndex uint32) uint32 {
	return atomic.LoadUint32(&h.entries[guildIndex&HazardMask].Score)
}

func (h *HazardScores) ResetEntry(guildIndex uint32) {
	entry := &h.entries[guildIndex&HazardMask]
	atomic.StoreUint32(&entry.ActorCount, 0)
	atomic.StoreUint32(&entry.DestructiveOps, 0)
	atomic.StoreUint32(&entry.Score, 0)
}

func CalculateHazardScore(actorCount, destructiveOps, velocity uint32) uint32 {
	score := uint32(0)

	if actorCount > 1 {
		score += actorCount * 10
	}

	score += destructiveOps * 5

	if velocity > 10 {
		score += velocity * 2
	}

	return score
}
