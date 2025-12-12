package ingest

import (
	"sync/atomic"

	"go-antinuke-2.0/internal/logging"
)

type OverflowStrategy uint8

const (
	OverflowDrop OverflowStrategy = iota
	OverflowBlock
	OverflowOverwrite
)

type RingBufferOverflow struct {
	strategy       OverflowStrategy
	droppedEvents  uint64
	overwriteCount uint64
}

func NewRingBufferOverflow(strategy OverflowStrategy) *RingBufferOverflow {
	return &RingBufferOverflow{
		strategy: strategy,
	}
}

func (rbo *RingBufferOverflow) HandleOverflow(rb *RingBuffer, event *Event) bool {
	switch rbo.strategy {
	case OverflowDrop:
		return rbo.handleDrop(event)
	case OverflowOverwrite:
		return rbo.handleOverwrite(rb, event)
	case OverflowBlock:
		return rbo.handleBlock(rb, event)
	default:
		return false
	}
}

func (rbo *RingBufferOverflow) handleDrop(event *Event) bool {
	atomic.AddUint64(&rbo.droppedEvents, 1)

	if event.Priority >= 3 {
		logging.Warn("Dropped critical event due to overflow: type=%d, guild=%d", event.EventType, event.GuildID)
	}

	return false
}

func (rbo *RingBufferOverflow) handleOverwrite(rb *RingBuffer, event *Event) bool {
	atomic.AddUint64(&rbo.overwriteCount, 1)

	rb.Dequeue()

	return rb.Enqueue(event)
}

func (rbo *RingBufferOverflow) handleBlock(rb *RingBuffer, event *Event) bool {
	for rb.IsFull() {
	}
	return rb.Enqueue(event)
}

func (rbo *RingBufferOverflow) GetDroppedCount() uint64 {
	return atomic.LoadUint64(&rbo.droppedEvents)
}

func (rbo *RingBufferOverflow) GetOverwriteCount() uint64 {
	return atomic.LoadUint64(&rbo.overwriteCount)
}
