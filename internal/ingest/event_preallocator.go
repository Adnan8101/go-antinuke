package ingest

import (
	"sync"
	"go-antinuke-2.0/pkg/util"
)

const (
	EventTypeUnknown = iota
	EventTypeBan
	EventTypeUnban
	EventTypeKick
	EventTypeBot
	EventTypeChannelCreate
	EventTypeChannelDelete
	EventTypeChannelUpdate
	EventTypeEmojiStickerCreate
	EventTypeEmojiStickerDelete
	EventTypeEmojiStickerUpdate
	EventTypeEveryoneHerePing
	EventTypeLinkRole
	EventTypeRoleCreate
	EventTypeRoleDelete
	EventTypeRoleUpdate
	EventTypeRolePing
	EventTypeMemberUpdate
	EventTypeIntegration
	EventTypeServerUpdate
	EventTypeAutomodRuleCreate
	EventTypeAutomodRuleUpdate
	EventTypeAutomodRuleDelete
	EventTypeGuildEventCreate
	EventTypeGuildEventUpdate
	EventTypeGuildEventDelete
	EventTypeWebhook
	EventTypePermChange
)

type Event struct {
	EventType uint8
	Priority  uint8
	Flags     uint16
	GuildID   uint64
	ActorID   uint64
	TargetID  uint64
	Metadata  uint64
	Timestamp int64
	_         [16]byte
}

// Event pool using sync.Pool for better GC performance
var eventPool = sync.Pool{
	New: func() interface{} {
		return &Event{}
	},
}

// AcquireEvent gets an event from the pool
func AcquireEvent() *Event {
	return eventPool.Get().(*Event)
}

// ReleaseEvent returns an event to the pool
func ReleaseEvent(e *Event) {
	// Reset event fields
	e.EventType = 0
	e.Priority = 0
	e.Flags = 0
	e.GuildID = 0
	e.ActorID = 0
	e.TargetID = 0
	e.Metadata = 0
	e.Timestamp = 0
	eventPool.Put(e)
}

// CreateEvent creates a new event with the given parameters
func CreateEvent(eventType uint8, guildID, actorID, targetID, metadata uint64) *Event {
	return &Event{
		EventType: eventType,
		GuildID:   guildID,
		ActorID:   actorID,
		TargetID:  targetID,
		Metadata:  metadata,
		Timestamp: util.NowMono(),
	}
}

type EventPool struct {
	pool  []Event
	index uint32
	size  uint32
}

func NewEventPool(size uint32) *EventPool {
	events := make([]Event, size)
	return &EventPool{
		pool:  events,
		index: 0,
		size:  size,
	}
}

func (ep *EventPool) Get() *Event {
	idx := ep.index
	ep.index = (ep.index + 1) % ep.size
	event := &ep.pool[idx]
	event.EventType = 0
	event.Priority = 0
	event.Flags = 0
	event.GuildID = 0
	event.ActorID = 0
	event.TargetID = 0
	event.Metadata = 0
	event.Timestamp = 0
	return event
}

func (ep *EventPool) Reset() {
	ep.index = 0
	for i := range ep.pool {
		ep.pool[i] = Event{}
	}
}

var GlobalEventPool *EventPool

func InitEventPool(size uint32) {
	GlobalEventPool = NewEventPool(size)
}

func GetEvent() *Event {
	if GlobalEventPool == nil {
		InitEventPool(65536)
	}
	return GlobalEventPool.Get()
}
