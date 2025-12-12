package ingest

const (
	OpDispatch            = 0
	OpHeartbeat           = 1
	OpIdentify            = 2
	OpPresenceUpdate      = 3
	OpVoiceStateUpdate    = 4
	OpResume              = 6
	OpReconnect           = 7
	OpRequestGuildMembers = 8
	OpInvalidSession      = 9
	OpHello               = 10
	OpHeartbeatACK        = 11
)

type OpcodeFilter struct{}

func NewOpcodeFilter() *OpcodeFilter {
	return &OpcodeFilter{}
}

func (of *OpcodeFilter) ShouldProcess(opcode int) bool {
	return opcode == OpDispatch
}

func (of *OpcodeFilter) IsControlMessage(opcode int) bool {
	return opcode == OpHeartbeat ||
		opcode == OpHeartbeatACK ||
		opcode == OpHello ||
		opcode == OpReconnect ||
		opcode == OpInvalidSession
}

func (of *OpcodeFilter) RequiresResponse(opcode int) bool {
	return opcode == OpHello || opcode == OpReconnect
}
