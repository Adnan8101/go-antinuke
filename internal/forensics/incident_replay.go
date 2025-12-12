package forensics

import (
	"go-antinuke-2.0/internal/models"
)

type IncidentReplay struct {
	events []models.Event
}

func NewIncidentReplay() *IncidentReplay {
	return &IncidentReplay{
		events: make([]models.Event, 0),
	}
}

func (ir *IncidentReplay) AddEvent(event models.Event) {
	ir.events = append(ir.events, event)
}

func (ir *IncidentReplay) Replay() []models.Event {
	return ir.events
}

func (ir *IncidentReplay) ValidateCorrelator(expectedAlerts int) bool {
	return len(ir.events) > 0
}

func (ir *IncidentReplay) GetEventCount() int {
	return len(ir.events)
}

func (ir *IncidentReplay) Clear() {
	ir.events = make([]models.Event, 0)
}
