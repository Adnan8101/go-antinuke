package decision

import (
	"go-antinuke-2.0/internal/state"
)

type AutoBanManager struct{}

func NewAutoBanManager() *AutoBanManager {
	return &AutoBanManager{}
}

func (abm *AutoBanManager) ShouldAutoBan(actorIndex uint32, severity uint8, confidence uint8) bool {
	if confidence < 80 {
		return false
	}

	if severity >= uint8(SeverityHigh) {
		return true
	}

	as := state.GetActorState()
	threatLevel := as.GetThreatLevel(actorIndex)

	return threatLevel >= 90
}

func (abm *AutoBanManager) MarkBanned(actorIndex uint32) {
	as := state.GetActorState()
	as.SetBanned(actorIndex, true)
}

func (abm *AutoBanManager) IsBanned(actorIndex uint32) bool {
	as := state.GetActorState()
	return as.IsBanned(actorIndex)
}

func (abm *AutoBanManager) GetBanReason(flags uint32, severity uint8) string {
	if severity >= uint8(SeverityCritical) {
		return "Critical threat detected - automated ban"
	}
	if severity >= uint8(SeverityHigh) {
		return "High severity threat - automated protection"
	}
	return "Anti-nuke protection triggered"
}
