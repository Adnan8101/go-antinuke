package correlator

import (
	"go-antinuke-2.0/pkg/util"
)

type Alert struct {
	GuildID   uint64
	ActorID   uint64
	TargetID  uint64
	EventType uint8
	Severity  uint8
	PanicMode uint8
	_         uint8
	Flags     uint32
	Timestamp int64
	_         [12]byte
}

type AlertQueue struct {
	alerts []Alert
	mask   uint32
	head   uint32
	tail   uint32
	_      [52]byte
}

func NewAlertQueue(size uint32) *AlertQueue {
	if size&(size-1) != 0 {
		size = nextPowerOf2Alert(size)
	}

	return &AlertQueue{
		alerts: make([]Alert, size),
		mask:   size - 1,
		head:   0,
		tail:   0,
	}
}

func (aq *AlertQueue) Get() *Alert {
	return &Alert{}
}

func (aq *AlertQueue) Enqueue(alert *Alert) bool {
	head := util.AtomicLoadU32(&aq.head)
	tail := util.AtomicLoadU32(&aq.tail)

	nextHead := (head + 1) & aq.mask
	if nextHead == tail {
		return false
	}

	aq.alerts[head] = *alert
	util.AtomicStoreU32(&aq.head, nextHead)
	return true
}

func (aq *AlertQueue) Dequeue() (*Alert, bool) {
	head := util.AtomicLoadU32(&aq.head)
	tail := util.AtomicLoadU32(&aq.tail)

	if tail == head {
		return nil, false
	}

	alert := &aq.alerts[tail]
	util.AtomicStoreU32(&aq.tail, (tail+1)&aq.mask)
	return alert, true
}

func (aq *AlertQueue) IsEmpty() bool {
	return util.AtomicLoadU32(&aq.head) == util.AtomicLoadU32(&aq.tail)
}

func (aq *AlertQueue) Size() uint32 {
	head := util.AtomicLoadU32(&aq.head)
	tail := util.AtomicLoadU32(&aq.tail)

	if head >= tail {
		return head - tail
	}
	return (aq.mask + 1) - (tail - head)
}

func nextPowerOf2Alert(n uint32) uint32 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
