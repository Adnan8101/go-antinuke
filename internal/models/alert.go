package models

type Alert struct {
	GuildID    uint64
	ActorID    uint64
	TargetID   uint64
	EventType  uint8
	Severity   uint8
	Confidence uint8
	_          uint8
	Flags      uint32
	Timestamp  int64
	Metadata   uint64
	_          [8]byte
}

const (
	SeverityNone = iota
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func NewAlert() *Alert {
	return &Alert{}
}

func (a *Alert) IsCritical() bool {
	return a.Severity >= SeverityCritical
}

func (a *Alert) IsHighConfidence() bool {
	return a.Confidence >= 80
}

func (a *Alert) RequiresImmediate() bool {
	return a.IsCritical() && a.IsHighConfidence()
}

func (a *Alert) Reset() {
	a.GuildID = 0
	a.ActorID = 0
	a.TargetID = 0
	a.EventType = 0
	a.Severity = 0
	a.Confidence = 0
	a.Flags = 0
	a.Timestamp = 0
	a.Metadata = 0
}
