package decision

import (
	"go-antinuke-2.0/internal/state"
)

type QuarantineManager struct{}

func NewQuarantineManager() *QuarantineManager {
	return &QuarantineManager{}
}

func (qm *QuarantineManager) QuarantineActor(actorIndex uint32) {
	as := state.GetActorState()
	as.UpdateThreatLevel(actorIndex, 100)
}

func (qm *QuarantineManager) ReleaseActor(actorIndex uint32) {
	as := state.GetActorState()
	as.UpdateThreatLevel(actorIndex, 0)
}

func (qm *QuarantineManager) IsQuarantined(actorIndex uint32) bool {
	as := state.GetActorState()
	return as.GetThreatLevel(actorIndex) >= 100
}

func (qm *QuarantineManager) GetThreatLevel(actorIndex uint32) uint32 {
	as := state.GetActorState()
	return as.GetThreatLevel(actorIndex)
}
