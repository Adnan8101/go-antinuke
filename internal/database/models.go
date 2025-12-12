package database

// GuildConfig represents guild-specific configuration
type GuildConfig struct {
	GuildID       string
	PanicMode     bool
	LogChannelID  string
	EnabledEvents string // Comma-separated event IDs
	CreatedAt     int64
	UpdatedAt     int64
}

// EventLog represents a logged event
type EventLog struct {
	ID               int64
	GuildID          string
	EventType        int
	ActorID          string
	TargetID         string
	DetectionSpeedUS int64 // Detection speed in microseconds
	ActionTaken      string
	Timestamp        int64
}

// EventType represents an event type definition
type EventType struct {
	ID          int
	Name        string
	Description string
	Emoji       string
}

// BannedUser represents a user banned by the bot
type BannedUser struct {
	ID       int64
	GuildID  string
	UserID   string
	Reason   string
	BannedAt int64
	BannedBy string // Actor ID that caused the ban
	IsBot    bool   // Whether the banned entity is a bot
	AddedBy  string // User ID who added the bot (if IsBot=true)
}
