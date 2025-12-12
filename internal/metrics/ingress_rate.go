package metrics

import (
	"sync/atomic"
	"time"
)

type IngressRateCounter struct {
	eventsProcessed uint64
	startTime       int64
}

func NewIngressRateCounter() *IngressRateCounter {
	return &IngressRateCounter{
		startTime: time.Now().UnixNano(),
	}
}

func (irc *IngressRateCounter) Increment() {
	atomic.AddUint64(&irc.eventsProcessed, 1)
}

func (irc *IngressRateCounter) GetRate() float64 {
	events := atomic.LoadUint64(&irc.eventsProcessed)
	elapsed := time.Now().UnixNano() - irc.startTime

	if elapsed == 0 {
		return 0
	}

	return float64(events) / (float64(elapsed) / 1e9)
}

func (irc *IngressRateCounter) GetCount() uint64 {
	return atomic.LoadUint64(&irc.eventsProcessed)
}

func (irc *IngressRateCounter) Reset() {
	atomic.StoreUint64(&irc.eventsProcessed, 0)
	irc.startTime = time.Now().UnixNano()
}
