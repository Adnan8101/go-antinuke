package metrics

import (
	"fmt"
)

type MetricsExporter struct {
	latencyHist      *LatencyHistogram
	ingressRate      *IngressRateCounter
	correlatorHealth *CorrelatorHealth
}

func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{
		latencyHist:      NewLatencyHistogram(),
		ingressRate:      NewIngressRateCounter(),
		correlatorHealth: NewCorrelatorHealth(),
	}
}

func (me *MetricsExporter) RecordLatency(latencyNs uint64) {
	me.latencyHist.Record(latencyNs)
}

func (me *MetricsExporter) IncrementIngress() {
	me.ingressRate.Increment()
}

func (me *MetricsExporter) RecordCorrelatorIteration() {
	me.correlatorHealth.RecordIteration()
}

func (me *MetricsExporter) RecordAlert() {
	me.correlatorHealth.RecordAlert()
}

func (me *MetricsExporter) Export() string {
	stats := me.latencyHist.GetStats()
	rate := me.ingressRate.GetRate()
	healthy := me.correlatorHealth.IsHealthy()

	return fmt.Sprintf(
		"latency_min_ns %d\nlatency_max_ns %d\nlatency_avg_ns %d\ningress_rate_eps %.2f\ncorrelator_healthy %v\n",
		stats.Min, stats.Max, stats.Avg, rate, healthy,
	)
}

func (me *MetricsExporter) GetLatencyStats() LatencyStats {
	return me.latencyHist.GetStats()
}

func (me *MetricsExporter) GetIngressRate() float64 {
	return me.ingressRate.GetRate()
}

func (me *MetricsExporter) IsCorrelatorHealthy() bool {
	return me.correlatorHealth.IsHealthy()
}
