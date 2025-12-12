package decision

import (
	"go-antinuke-2.0/internal/models"
)

type ActionPacket struct {
	Action      *models.Action
	Priority    uint8
	Timestamp   int64
	SourceAlert uint64
}

func NewActionPacket(action *models.Action) *ActionPacket {
	return &ActionPacket{
		Action:    action,
		Priority:  action.Priority,
		Timestamp: action.Timestamp,
	}
}

func (ap *ActionPacket) IsCritical() bool {
	return ap.Priority >= 3
}

func (ap *ActionPacket) ShouldBypassCooldown() bool {
	return ap.Action.Type == models.ActionTypeLockdown
}
