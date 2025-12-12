package models

type Event struct {
	EventType  uint8
	Priority   uint8
	Flags      uint16
	GuildID    uint64
	ActorID    uint64
	TargetID   uint64
	Metadata   uint64
	Timestamp  int64
	SequenceID uint64
	_          [8]byte
}

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
)

func NewEvent() *Event {
	return &Event{}
}

func (e *Event) Reset() {
	e.EventType = 0
	e.Priority = 0
	e.Flags = 0
	e.GuildID = 0
	e.ActorID = 0
	e.TargetID = 0
	e.Metadata = 0
	e.Timestamp = 0
	e.SequenceID = 0
}

func (e *Event) IsDestructive() bool {
	return e.EventType == EventTypeBan ||
		e.EventType == EventTypeChannelDelete ||
		e.EventType == EventTypeRoleDelete
}

func (e *Event) RequiresFastPath() bool {
	return e.Priority >= 2
}
