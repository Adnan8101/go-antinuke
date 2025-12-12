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

// EventLimit represents rate limit configuration for an event
type EventLimit struct {
	ID         int64
	GuildID    string
	EventType  int
	MaxActions int    // Maximum number of actions allowed
	TimeWindow int    // Time window in seconds
	Punishment string // "kick", "ban", "timeout"
	CreatedAt  int64
	UpdatedAt  int64
}

// Whitelist represents whitelisted users/roles for specific events
type Whitelist struct {
	ID         int64
	GuildID    string
	TargetID   string // User ID or Role ID
	TargetType string // "user" or "role"
	EventType  int    // Event type ID, 0 means all events
	CreatedAt  int64
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
