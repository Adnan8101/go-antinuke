package decision

import (
	"go-antinuke-2.0/internal/state"
)

type LockdownManager struct{}

func NewLockdownManager() *LockdownManager {
	return &LockdownManager{}
}

func (lm *LockdownManager) ActivateLockdown(guildIndex uint32) {
	gs := state.GetGuildState()
	gs.SetLockdown(guildIndex, true)
}

func (lm *LockdownManager) DeactivateLockdown(guildIndex uint32) {
	gs := state.GetGuildState()
	gs.SetLockdown(guildIndex, false)
}

func (lm *LockdownManager) IsLockdown(guildIndex uint32) bool {
	gs := state.GetGuildState()
	return gs.IsLockdown(guildIndex)
}

func (lm *LockdownManager) GetLockdownActions() []string {
	return []string{
		"Freeze all role modifications",
		"Block channel creation/deletion",
		"Suspend webhook operations",
		"Restrict permission changes",
		"Enable enhanced logging",
	}
}
