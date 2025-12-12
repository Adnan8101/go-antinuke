package decision

import (
	"go-antinuke-2.0/internal/detectors"
)

type HeuristicResult struct {
	ShouldBan        bool
	ShouldQuarantine bool
	ShouldLockdown   bool
	ShouldFreeze     bool
	Confidence       uint8
}

func ApplyHeuristics(flags uint32, severity uint8) *HeuristicResult {
	result := &HeuristicResult{
		Confidence: 80,
	}

	destructiveCount := CountDestructiveFlags(flags)

	if destructiveCount >= 3 {
		result.ShouldBan = true
		result.ShouldLockdown = true
		result.Confidence = 95
		return result
	}

	if severity >= uint8(SeverityCritical) {
		result.ShouldBan = true
		result.ShouldLockdown = true
		result.Confidence = 90
		return result
	}

	if severity >= uint8(SeverityHigh) {
		result.ShouldBan = true
		result.ShouldFreeze = true
		result.Confidence = 85
		return result
	}

	if severity >= uint8(SeverityMedium) {
		result.ShouldQuarantine = true
		result.ShouldFreeze = true
		result.Confidence = 75
		return result
	}

	return result
}

func CountDestructiveFlags(flags uint32) uint32 {
	count := uint32(0)

	if (flags & detectors.FlagBanTriggered) != 0 {
		count++
	}
	if (flags & detectors.FlagChannelTriggered) != 0 {
		count++
	}
	if (flags & detectors.FlagRoleTriggered) != 0 {
		count++
	}
	if (flags & detectors.FlagPermissionTriggered) != 0 {
		count++
	}

	return count
}

func ShouldEscalate(flags uint32, velocity uint32) bool {
	hasMultiActor := (flags & detectors.FlagMultiActorTriggered) != 0
	hasVelocity := (flags & detectors.FlagVelocityTriggered) != 0
	hasDestructive := CountDestructiveFlags(flags) >= 2

	return (hasMultiActor && hasDestructive) || (hasVelocity && velocity > 20)
}
