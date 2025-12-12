package forensics

import (
	"sort"
	"time"
)

type TimelineEvent struct {
	Timestamp int64
	EventType string
	GuildID   uint64
	ActorID   uint64
	TargetID  uint64
	Action    string
	Metadata  map[string]interface{}
}

type Timeline struct {
	events []TimelineEvent
}

func NewTimeline() *Timeline {
	return &Timeline{
		events: make([]TimelineEvent, 0),
	}
}

func (t *Timeline) AddEvent(event TimelineEvent) {
	t.events = append(t.events, event)
}

func (t *Timeline) Sort() {
	sort.Slice(t.events, func(i, j int) bool {
		return t.events[i].Timestamp < t.events[j].Timestamp
	})
}

func (t *Timeline) GetEvents(startTime, endTime int64) []TimelineEvent {
	result := make([]TimelineEvent, 0)

	for _, event := range t.events {
		if event.Timestamp >= startTime && event.Timestamp <= endTime {
			result = append(result, event)
		}
	}

	return result
}

func (t *Timeline) Reconstruct(guildID uint64) []TimelineEvent {
	result := make([]TimelineEvent, 0)

	for _, event := range t.events {
		if event.GuildID == guildID {
			result = append(result, event)
		}
	}

	return result
}

func (t *Timeline) GetDuration() time.Duration {
	if len(t.events) == 0 {
		return 0
	}

	t.Sort()
	first := t.events[0].Timestamp
	last := t.events[len(t.events)-1].Timestamp

	return time.Duration(last - first)
}
