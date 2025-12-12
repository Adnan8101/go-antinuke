package metrics

import (
	"sync"
)

type MetricsRegistry struct {
	mu               sync.RWMutex
	latencyHist      *LatencyHistogram
	ingressRate      *IngressRateCounter
	correlatorHealth *CorrelatorHealth
	customMetrics    map[string]interface{}
}

func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		latencyHist:      NewLatencyHistogram(),
		ingressRate:      NewIngressRateCounter(),
		correlatorHealth: NewCorrelatorHealth(),
		customMetrics:    make(map[string]interface{}),
	}
}

func (mr *MetricsRegistry) RegisterCustom(name string, metric interface{}) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.customMetrics[name] = metric
}

func (mr *MetricsRegistry) GetCustom(name string) interface{} {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.customMetrics[name]
}

func (mr *MetricsRegistry) GetLatencyHistogram() *LatencyHistogram {
	return mr.latencyHist
}

func (mr *MetricsRegistry) GetIngressRate() *IngressRateCounter {
	return mr.ingressRate
}

func (mr *MetricsRegistry) GetCorrelatorHealth() *CorrelatorHealth {
	return mr.correlatorHealth
}

var GlobalRegistry *MetricsRegistry

func InitGlobalRegistry() {
	GlobalRegistry = NewMetricsRegistry()
}

func GetRegistry() *MetricsRegistry {
	if GlobalRegistry == nil {
		InitGlobalRegistry()
	}
	return GlobalRegistry
}
