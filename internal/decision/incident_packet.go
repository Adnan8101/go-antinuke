package decision

type IncidentPacket struct {
	GuildID    uint64
	ActorID    uint64
	TargetID   uint64
	EventType  uint8
	Severity   uint8
	Confidence uint8
	SafetyMode uint8
	PanicMode  uint8
	_          [3]byte
	Flags      uint32
	Timestamp  int64
	_          [8]byte
}

type IncidentType uint8

const (
	IncidentBanSpike IncidentType = iota
	IncidentChannelNuke
	IncidentRoleNuke
	IncidentPermEscalation
	IncidentWebhookSpam
	IncidentMultiActor
	IncidentVelocitySpike
)

func NewIncidentPacket(guildID, actorID uint64, incidentType IncidentType) *IncidentPacket {
	return &IncidentPacket{
		GuildID:    guildID,
		ActorID:    actorID,
		EventType:  uint8(incidentType),
		Confidence: 90,
	}
}

func (ip *IncidentPacket) IsCritical() bool {
	return ip.Severity >= uint8(SeverityCritical)
}

func (ip *IncidentPacket) IsHighConfidence() bool {
	return ip.Confidence >= 80
}

func (ip *IncidentPacket) RequiresImmediate() bool {
	return ip.IsCritical() && ip.IsHighConfidence()
}
