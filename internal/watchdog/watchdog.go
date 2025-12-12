package watchdog

import (
	"sync/atomic"
	"time"

	"go-antinuke-2.0/internal/logging"
)

type Watchdog struct {
	components     map[string]*ComponentHealth
	checkInterval  time.Duration
	running        uint32
	alertThreshold int
}

type ComponentHealth struct {
	Name          string
	LastHeartbeat int64
	IsHealthy     uint32
	Threshold     time.Duration
}

func NewWatchdog(checkInterval time.Duration) *Watchdog {
	return &Watchdog{
		components:     make(map[string]*ComponentHealth),
		checkInterval:  checkInterval,
		alertThreshold: 3,
	}
}

func (w *Watchdog) RegisterComponent(name string, threshold time.Duration) {
	w.components[name] = &ComponentHealth{
		Name:      name,
		IsHealthy: 1,
		Threshold: threshold,
	}
}

func (w *Watchdog) Heartbeat(name string) {
	if comp, exists := w.components[name]; exists {
		atomic.StoreInt64(&comp.LastHeartbeat, time.Now().UnixNano())
		atomic.StoreUint32(&comp.IsHealthy, 1)
	}
}

func (w *Watchdog) Start() {
	atomic.StoreUint32(&w.running, 1)
	go w.monitorLoop()
}

func (w *Watchdog) monitorLoop() {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for atomic.LoadUint32(&w.running) == 1 {
		<-ticker.C
		w.checkAllComponents()
	}
}

func (w *Watchdog) checkAllComponents() {
	now := time.Now().UnixNano()

	for name, comp := range w.components {
		lastBeat := atomic.LoadInt64(&comp.LastHeartbeat)
		if lastBeat == 0 {
			continue
		}

		elapsed := time.Duration(now - lastBeat)
		if elapsed > comp.Threshold {
			atomic.StoreUint32(&comp.IsHealthy, 0)
			logging.Error("Watchdog: %s unhealthy (no heartbeat for %v)", name, elapsed)
		}
	}
}

func (w *Watchdog) IsHealthy(name string) bool {
	if comp, exists := w.components[name]; exists {
		return atomic.LoadUint32(&comp.IsHealthy) == 1
	}
	return false
}

func (w *Watchdog) Stop() {
	atomic.StoreUint32(&w.running, 0)
}

func (w *Watchdog) GetStatus() map[string]bool {
	status := make(map[string]bool)
	for name, comp := range w.components {
		status[name] = atomic.LoadUint32(&comp.IsHealthy) == 1
	}
	return status
}
