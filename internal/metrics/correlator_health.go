package metrics

import (
	"sync/atomic"
	"time"
)

type CorrelatorHealth struct {
	loopIterations uint64
	alertsEmitted  uint64
	lastLoopTime   int64
	isHealthy      uint32
}

func NewCorrelatorHealth() *CorrelatorHealth {
	return &CorrelatorHealth{
		isHealthy: 1,
	}
}

func (ch *CorrelatorHealth) RecordIteration() {
	atomic.AddUint64(&ch.loopIterations, 1)
	atomic.StoreInt64(&ch.lastLoopTime, time.Now().UnixNano())
}

func (ch *CorrelatorHealth) RecordAlert() {
	atomic.AddUint64(&ch.alertsEmitted, 1)
}

func (ch *CorrelatorHealth) GetIterations() uint64 {
	return atomic.LoadUint64(&ch.loopIterations)
}

func (ch *CorrelatorHealth) GetAlerts() uint64 {
	return atomic.LoadUint64(&ch.alertsEmitted)
}

func (ch *CorrelatorHealth) IsHealthy() bool {
	lastLoop := atomic.LoadInt64(&ch.lastLoopTime)
	if lastLoop == 0 {
		return true
	}

	elapsed := time.Now().UnixNano() - lastLoop
	return elapsed < 100_000_000
}

func (ch *CorrelatorHealth) SetHealthy(healthy bool) {
	val := uint32(0)
	if healthy {
		val = 1
	}
	atomic.StoreUint32(&ch.isHealthy, val)
}
