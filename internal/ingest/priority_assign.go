package ingest

type PriorityLevel uint8

const (
	PriorityNone PriorityLevel = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

var EventPriorityMap = map[uint8]PriorityLevel{
	EventTypeUnknown:       PriorityNone,
	EventTypeBan:           PriorityCritical,
	EventTypeKick:          PriorityMedium,
	EventTypeChannelDelete: PriorityCritical,
	EventTypeRoleDelete:    PriorityCritical,
	EventTypeWebhook:       PriorityMedium,
	EventTypePermChange:    PriorityMedium,
}

func AssignPriority(event *Event) {
	if priority, exists := EventPriorityMap[event.EventType]; exists {
		event.Priority = uint8(priority)
	} else {
		event.Priority = uint8(PriorityNone)
	}
}

func ShouldEnterHotPath(event *Event) bool {
	return event.Priority >= uint8(PriorityMedium)
}

func ShouldBypass(event *Event) bool {
	return event.Priority < uint8(PriorityMedium)
}

func PrioritizeEvent(event *Event, guildSize uint32) {
	basePriority := EventPriorityMap[event.EventType]

	if guildSize > 10000 && basePriority == PriorityMedium {
		event.Priority = uint8(PriorityHigh)
	} else {
		event.Priority = uint8(basePriority)
	}
}

func GetPriorityName(priority uint8) string {
	switch PriorityLevel(priority) {
	case PriorityNone:
		return "none"
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}
