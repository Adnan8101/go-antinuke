package ingest

import (
	"go-antinuke-2.0/pkg/util"
)

type RingBuffer struct {
	buffer []Event
	mask   uint32
	head   uint32
	tail   uint32
	_      [52]byte
}

func NewRingBuffer(size uint32) *RingBuffer {
	if size&(size-1) != 0 {
		size = nextPowerOf2(size)
	}

	return &RingBuffer{
		buffer: make([]Event, size),
		mask:   size - 1,
		head:   0,
		tail:   0,
	}
}

func (rb *RingBuffer) Enqueue(event *Event) bool {
	head := util.AtomicLoadU32(&rb.head)
	tail := util.AtomicLoadU32(&rb.tail)

	nextHead := (head + 1) & rb.mask
	if nextHead == tail {
		return false
	}

	rb.buffer[head] = *event
	util.AtomicStoreU32(&rb.head, nextHead)
	return true
}

func (rb *RingBuffer) Dequeue() (*Event, bool) {
	head := util.AtomicLoadU32(&rb.head)
	tail := util.AtomicLoadU32(&rb.tail)

	if tail == head {
		return nil, false
	}

	event := &rb.buffer[tail]
	util.AtomicStoreU32(&rb.tail, (tail+1)&rb.mask)
	return event, true
}

func (rb *RingBuffer) Peek() (*Event, bool) {
	head := util.AtomicLoadU32(&rb.head)
	tail := util.AtomicLoadU32(&rb.tail)

	if tail == head {
		return nil, false
	}

	return &rb.buffer[tail], true
}

func (rb *RingBuffer) IsEmpty() bool {
	return util.AtomicLoadU32(&rb.head) == util.AtomicLoadU32(&rb.tail)
}

func (rb *RingBuffer) IsFull() bool {
	head := util.AtomicLoadU32(&rb.head)
	tail := util.AtomicLoadU32(&rb.tail)
	nextHead := (head + 1) & rb.mask
	return nextHead == tail
}

func (rb *RingBuffer) Size() uint32 {
	head := util.AtomicLoadU32(&rb.head)
	tail := util.AtomicLoadU32(&rb.tail)

	if head >= tail {
		return head - tail
	}
	return (rb.mask + 1) - (tail - head)
}

func (rb *RingBuffer) Capacity() uint32 {
	return rb.mask + 1
}

func (rb *RingBuffer) Reset() {
	util.AtomicStoreU32(&rb.head, 0)
	util.AtomicStoreU32(&rb.tail, 0)
}

func nextPowerOf2(n uint32) uint32 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

var GlobalRingBuffer *RingBuffer

func InitRingBuffer(size uint32) {
	GlobalRingBuffer = NewRingBuffer(size)
}

func GetRingBuffer() *RingBuffer {
	if GlobalRingBuffer == nil {
		InitRingBuffer(65536)
	}
	return GlobalRingBuffer
}
