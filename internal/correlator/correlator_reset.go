package correlator

import (
	"time"

	"go-antinuke-2.0/internal/logging"
)

type ResetManager struct {
	resetInterval time.Duration
	lastReset     int64
}

func NewResetManager(interval time.Duration) *ResetManager {
	return &ResetManager{
		resetInterval: interval,
		lastReset:     time.Now().UnixNano(),
	}
}

func (rm *ResetManager) ShouldReset() bool {
	now := time.Now().UnixNano()
	elapsed := now - rm.lastReset
	return elapsed >= int64(rm.resetInterval)
}

func (rm *ResetManager) ResetCounters() {
	counters := GetCounters()
	counters.ResetAll()

	velocity := GetVelocity()
	for i := uint32(0); i < 8192; i++ {
		velocity.Reset(i)
	}

	rm.lastReset = time.Now().UnixNano()
	logging.Info("Counters reset complete")
}

func (rm *ResetManager) PartialReset(guildIndex uint32) {
	counters := GetCounters()
	counters.Reset(guildIndex)

	velocity := GetVelocity()
	velocity.Reset(guildIndex)
}

func (rm *ResetManager) Start() {
	go rm.resetLoop()
}

func (rm *ResetManager) resetLoop() {
	ticker := time.NewTicker(rm.resetInterval)
	defer ticker.Stop()

	for range ticker.C {
		if rm.ShouldReset() {
			rm.ResetCounters()
		}
	}
}
