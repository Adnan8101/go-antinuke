package config

type SafetyMode uint8

const (
	SafetyNormal SafetyMode = iota
	SafetyElevated
	SafetyHigh
	SafetyLockdown
	SafetyEmergency
)

type ModeConfig struct {
	Mode                SafetyMode
	AutoBanEnabled      bool
	AutoLockdownEnabled bool
	QuarantineEnabled   bool
	FreezeActorsEnabled bool
	ThresholdMultiplier float32
	ResponseDelayMs     uint32
}

var SafetyModeConfigs = map[SafetyMode]ModeConfig{
	SafetyNormal: {
		Mode:                SafetyNormal,
		AutoBanEnabled:      true,
		AutoLockdownEnabled: false,
		QuarantineEnabled:   true,
		FreezeActorsEnabled: true,
		ThresholdMultiplier: 1.0,
		ResponseDelayMs:     0,
	},
	SafetyElevated: {
		Mode:                SafetyElevated,
		AutoBanEnabled:      true,
		AutoLockdownEnabled: false,
		QuarantineEnabled:   true,
		FreezeActorsEnabled: true,
		ThresholdMultiplier: 0.8,
		ResponseDelayMs:     0,
	},
	SafetyHigh: {
		Mode:                SafetyHigh,
		AutoBanEnabled:      true,
		AutoLockdownEnabled: true,
		QuarantineEnabled:   true,
		FreezeActorsEnabled: true,
		ThresholdMultiplier: 0.6,
		ResponseDelayMs:     0,
	},
	SafetyLockdown: {
		Mode:                SafetyLockdown,
		AutoBanEnabled:      true,
		AutoLockdownEnabled: true,
		QuarantineEnabled:   true,
		FreezeActorsEnabled: true,
		ThresholdMultiplier: 0.4,
		ResponseDelayMs:     0,
	},
	SafetyEmergency: {
		Mode:                SafetyEmergency,
		AutoBanEnabled:      true,
		AutoLockdownEnabled: true,
		QuarantineEnabled:   true,
		FreezeActorsEnabled: true,
		ThresholdMultiplier: 0.2,
		ResponseDelayMs:     0,
	},
}

func GetSafetyModeConfig(mode SafetyMode) ModeConfig {
	return SafetyModeConfigs[mode]
}

func ApplyThresholdMultiplier(base uint32, mode SafetyMode) uint32 {
	config := SafetyModeConfigs[mode]
	adjusted := float32(base) * config.ThresholdMultiplier
	if adjusted < 1.0 {
		adjusted = 1.0
	}
	return uint32(adjusted)
}

func ShouldAutoBan(mode SafetyMode) bool {
	return SafetyModeConfigs[mode].AutoBanEnabled
}

func ShouldAutoLockdown(mode SafetyMode) bool {
	return SafetyModeConfigs[mode].AutoLockdownEnabled
}

func ShouldFreezeActors(mode SafetyMode) bool {
	return SafetyModeConfigs[mode].FreezeActorsEnabled
}

func ShouldQuarantine(mode SafetyMode) bool {
	return SafetyModeConfigs[mode].QuarantineEnabled
}

func (s SafetyMode) String() string {
	switch s {
	case SafetyNormal:
		return "normal"
	case SafetyElevated:
		return "elevated"
	case SafetyHigh:
		return "high"
	case SafetyLockdown:
		return "lockdown"
	case SafetyEmergency:
		return "emergency"
	default:
		return "unknown"
	}
}

func ParseSafetyMode(s string) SafetyMode {
	switch s {
	case "normal":
		return SafetyNormal
	case "elevated":
		return SafetyElevated
	case "high":
		return SafetyHigh
	case "lockdown":
		return SafetyLockdown
	case "emergency":
		return SafetyEmergency
	default:
		return SafetyNormal
	}
}
