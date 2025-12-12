package ingest

import (
	"sync/atomic"
	"time"
)

type HeartbeatMonitor struct {
	lastHeartbeatSent int64
	lastHeartbeatACK  int64
	missedBeats       uint32
	isHealthy         uint32
}

func NewHeartbeatMonitor() *HeartbeatMonitor {
	return &HeartbeatMonitor{
		isHealthy: 1,
	}
}

func (hm *HeartbeatMonitor) RecordSent() {
	atomic.StoreInt64(&hm.lastHeartbeatSent, time.Now().UnixNano())
}

func (hm *HeartbeatMonitor) RecordACK() {
	atomic.StoreInt64(&hm.lastHeartbeatACK, time.Now().UnixNano())
	atomic.StoreUint32(&hm.missedBeats, 0)
	atomic.StoreUint32(&hm.isHealthy, 1)
}

func (hm *HeartbeatMonitor) RecordMissed() {
	missed := atomic.AddUint32(&hm.missedBeats, 1)
	if missed >= 3 {
		atomic.StoreUint32(&hm.isHealthy, 0)
	}
}

func (hm *HeartbeatMonitor) IsHealthy() bool {
	return atomic.LoadUint32(&hm.isHealthy) == 1
}

func (hm *HeartbeatMonitor) GetMissedCount() uint32 {
	return atomic.LoadUint32(&hm.missedBeats)
}

func (hm *HeartbeatMonitor) GetLatency() int64 {
	sent := atomic.LoadInt64(&hm.lastHeartbeatSent)
	ack := atomic.LoadInt64(&hm.lastHeartbeatACK)

	if sent == 0 || ack == 0 {
		return 0
	}

	return ack - sent
}
