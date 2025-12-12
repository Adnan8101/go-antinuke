package metrics

import (
	"sync/atomic"
)

type LatencyHistogram struct {
	buckets [10]uint64
	min     uint64
	max     uint64
	count   uint64
	sum     uint64
}

func NewLatencyHistogram() *LatencyHistogram {
	return &LatencyHistogram{}
}

func (lh *LatencyHistogram) Record(latencyNs uint64) {
	atomic.AddUint64(&lh.count, 1)
	atomic.AddUint64(&lh.sum, latencyNs)

	for {
		oldMin := atomic.LoadUint64(&lh.min)
		if latencyNs >= oldMin && oldMin != 0 {
			break
		}
		if atomic.CompareAndSwapUint64(&lh.min, oldMin, latencyNs) {
			break
		}
	}

	for {
		oldMax := atomic.LoadUint64(&lh.max)
		if latencyNs <= oldMax {
			break
		}
		if atomic.CompareAndSwapUint64(&lh.max, oldMax, latencyNs) {
			break
		}
	}

	bucketIndex := lh.getBucketIndex(latencyNs)
	atomic.AddUint64(&lh.buckets[bucketIndex], 1)
}

func (lh *LatencyHistogram) getBucketIndex(latencyNs uint64) int {
	latencyUs := latencyNs / 1000

	switch {
	case latencyUs < 1:
		return 0
	case latencyUs < 10:
		return 1
	case latencyUs < 100:
		return 2
	case latencyUs < 1000:
		return 3
	case latencyUs < 10000:
		return 4
	case latencyUs < 100000:
		return 5
	case latencyUs < 1000000:
		return 6
	default:
		return 7
	}
}

func (lh *LatencyHistogram) GetStats() LatencyStats {
	count := atomic.LoadUint64(&lh.count)
	sum := atomic.LoadUint64(&lh.sum)

	avg := uint64(0)
	if count > 0 {
		avg = sum / count
	}

	return LatencyStats{
		Min:   atomic.LoadUint64(&lh.min),
		Max:   atomic.LoadUint64(&lh.max),
		Avg:   avg,
		Count: count,
	}
}

type LatencyStats struct {
	Min   uint64
	Max   uint64
	Avg   uint64
	Count uint64
}
