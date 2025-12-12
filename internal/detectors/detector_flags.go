package detectors

const (
	FlagBanTriggered uint32 = 1 << iota
	FlagChannelTriggered
	FlagRoleTriggered
	FlagWebhookTriggered
	FlagPermissionTriggered
	FlagVelocityTriggered
	FlagMultiActorTriggered
	FlagLockdownActive
)

type FlagDetector struct{}

func NewFlagDetector() *FlagDetector {
	return &FlagDetector{}
}

func (d *FlagDetector) SetFlag(flags uint32, flag uint32) uint32 {
	return flags | flag
}

func (d *FlagDetector) ClearFlag(flags uint32, flag uint32) uint32 {
	return flags &^ flag
}

func (d *FlagDetector) HasFlag(flags uint32, flag uint32) bool {
	return (flags & flag) != 0
}

func (d *FlagDetector) ToggleFlag(flags uint32, flag uint32) uint32 {
	return flags ^ flag
}

func (d *FlagDetector) CountFlags(flags uint32) uint32 {
	count := uint32(0)
	for flags != 0 {
		count += flags & 1
		flags >>= 1
	}
	return count
}

func BranchlessHasFlag(flags, target uint32) uint32 {
	masked := flags & target
	return (masked | -masked) >> 31
}

func AnyFlagsSet(flags, mask uint32) bool {
	return (flags & mask) != 0
}

func AllFlagsSet(flags, mask uint32) bool {
	return (flags & mask) == mask
}

func GetSeverityFromFlags(flags uint32) uint8 {
	count := uint8(0)
	for flags != 0 {
		count += uint8(flags & 1)
		flags >>= 1
	}
	return count
}
