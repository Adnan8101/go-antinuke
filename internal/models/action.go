package models

type Action struct {
	Type       uint8
	Priority   uint8
	_          [2]byte
	GuildID    uint64
	TargetID   uint64
	Reason     string
	Metadata   map[string]interface{}
	Timestamp  int64
	RetryCount uint8
	_          [7]byte
}

const (
	ActionTypeBan = iota
	ActionTypeKick
	ActionTypeQuarantine
	ActionTypeLockdown
	ActionTypeRoleRemove
	ActionTypeRoleRestore
	ActionTypeChannelFreeze
)

func NewAction(actionType uint8, guildID, targetID uint64, reason string) *Action {
	return &Action{
		Type:     actionType,
		GuildID:  guildID,
		TargetID: targetID,
		Reason:   reason,
		Metadata: make(map[string]interface{}),
	}
}

func NewBanAction(guildID, targetID uint64, reason string) *Action {
	return NewAction(ActionTypeBan, guildID, targetID, reason)
}

func NewKickAction(guildID, targetID uint64, reason string) *Action {
	return NewAction(ActionTypeKick, guildID, targetID, reason)
}

func NewLockdownAction(guildID uint64) *Action {
	return NewAction(ActionTypeLockdown, guildID, 0, "Emergency lockdown")
}

func (a *Action) ShouldRetry() bool {
	return a.RetryCount < 3
}

func (a *Action) IncrementRetry() {
	a.RetryCount++
}
