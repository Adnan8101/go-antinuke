package decision

import (
	"go-antinuke-2.0/internal/detectors"
)

type SeverityLevel uint8

const (
	SeverityNone SeverityLevel = iota
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func EvaluateSeverity(flags uint32, flagCount uint8) uint8 {
	score := uint32(flagCount) * 10

	if (flags & detectors.FlagBanTriggered) != 0 {
		score += 30
	}
	if (flags & detectors.FlagChannelTriggered) != 0 {
		score += 60
	}
	if (flags & detectors.FlagRoleTriggered) != 0 {
		score += 60
	}
	if (flags & detectors.FlagPermissionTriggered) != 0 {
		score += 50
	}
	if (flags & detectors.FlagMultiActorTriggered) != 0 {
		score += 25
	}
	if (flags & detectors.FlagVelocityTriggered) != 0 {
		score += 20
	}

	return ScoreToSeverity(score)
}

func ScoreToSeverity(score uint32) uint8 {
	switch {
	case score >= 100:
		return uint8(SeverityCritical)
	case score >= 60:
		return uint8(SeverityHigh)
	case score >= 30:
		return uint8(SeverityMedium)
	case score >= 10:
		return uint8(SeverityLow)
	default:
		return uint8(SeverityNone)
	}
}

func (s SeverityLevel) String() string {
	switch s {
	case SeverityNone:
		return "none"
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}
