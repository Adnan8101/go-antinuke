package config

type GuildSizeCategory uint8

const (
	SizeTiny GuildSizeCategory = iota
	SizeSmall
	SizeMedium
	SizeLarge
	SizeHuge
)

type ThresholdMatrix struct {
	BanThreshold      uint32
	KickThreshold     uint32
	ChannelThreshold  uint32
	RoleThreshold     uint32
	WebhookThreshold  uint32
	PermThreshold     uint32
	VelocityThreshold uint32
	WindowMs          uint32
}

var DefaultThresholdMatrix = map[GuildSizeCategory]ThresholdMatrix{
	SizeTiny: {
		BanThreshold:      3,
		KickThreshold:     5,
		ChannelThreshold:  1,
		RoleThreshold:     1,
		WebhookThreshold:  5,
		PermThreshold:     3,
		VelocityThreshold: 10,
		WindowMs:          100,
	},
	SizeSmall: {
		BanThreshold:      5,
		KickThreshold:     8,
		ChannelThreshold:  1,
		RoleThreshold:     1,
		WebhookThreshold:  8,
		PermThreshold:     5,
		VelocityThreshold: 15,
		WindowMs:          100,
	},
	SizeMedium: {
		BanThreshold:      7,
		KickThreshold:     12,
		ChannelThreshold:  5,
		RoleThreshold:     5,
		WebhookThreshold:  10,
		PermThreshold:     7,
		VelocityThreshold: 20,
		WindowMs:          150,
	},
	SizeLarge: {
		BanThreshold:      10,
		KickThreshold:     15,
		ChannelThreshold:  7,
		RoleThreshold:     7,
		WebhookThreshold:  15,
		PermThreshold:     10,
		VelocityThreshold: 30,
		WindowMs:          200,
	},
	SizeHuge: {
		BanThreshold:      15,
		KickThreshold:     20,
		ChannelThreshold:  10,
		RoleThreshold:     10,
		WebhookThreshold:  20,
		PermThreshold:     15,
		VelocityThreshold: 40,
		WindowMs:          250,
	},
}

func GetThresholdMatrix(category GuildSizeCategory) ThresholdMatrix {
	return DefaultThresholdMatrix[category]
}

func GetCategoryBySize(memberCount uint32) GuildSizeCategory {
	switch {
	case memberCount < 100:
		return SizeTiny
	case memberCount < 1000:
		return SizeSmall
	case memberCount < 5000:
		return SizeMedium
	case memberCount < 20000:
		return SizeLarge
	default:
		return SizeHuge
	}
}

type CustomThreshold struct {
	GuildID    uint64
	Thresholds ThresholdMatrix
}

type ThresholdConfig struct {
	DefaultMatrix map[GuildSizeCategory]ThresholdMatrix
	CustomGuilds  map[uint64]ThresholdMatrix
}

var GlobalThresholds *ThresholdConfig

func InitThresholds() {
	GlobalThresholds = &ThresholdConfig{
		DefaultMatrix: DefaultThresholdMatrix,
		CustomGuilds:  make(map[uint64]ThresholdMatrix),
	}
}

func GetThresholds() *ThresholdConfig {
	if GlobalThresholds == nil {
		InitThresholds()
	}
	return GlobalThresholds
}

func SetCustomThreshold(guildID uint64, thresholds ThresholdMatrix) {
	GlobalThresholds.CustomGuilds[guildID] = thresholds
}

func GetGuildThresholds(guildID uint64, memberCount uint32) ThresholdMatrix {
	if custom, exists := GlobalThresholds.CustomGuilds[guildID]; exists {
		return custom
	}
	category := GetCategoryBySize(memberCount)
	return DefaultThresholdMatrix[category]
}
