package ha

import (
	"sync/atomic"
	"time"
)

type HeartbeatManager struct {
	cluster  *Cluster
	interval time.Duration
	lastBeat int64
	running  uint32
}

func NewHeartbeatManager(cluster *Cluster, interval time.Duration) *HeartbeatManager {
	return &HeartbeatManager{
		cluster:  cluster,
		interval: interval,
	}
}

func (hm *HeartbeatManager) Start() {
	atomic.StoreUint32(&hm.running, 1)
	go hm.heartbeatLoop()
}

func (hm *HeartbeatManager) heartbeatLoop() {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for atomic.LoadUint32(&hm.running) == 1 {
		select {
		case <-ticker.C:
			hm.sendHeartbeat()
		}
	}
}

func (hm *HeartbeatManager) sendHeartbeat() {
	atomic.StoreInt64(&hm.lastBeat, time.Now().UnixNano())
}

func (hm *HeartbeatManager) GetLastHeartbeat() int64 {
	return atomic.LoadInt64(&hm.lastBeat)
}

func (hm *HeartbeatManager) Stop() {
	atomic.StoreUint32(&hm.running, 0)
}
