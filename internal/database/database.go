package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

var globalDB *Database

// Initialize creates and initializes the SQLite database
func Initialize(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("failed to enable WAL: %w", err)
	}

	_, err = db.Exec("PRAGMA synchronous=NORMAL")
	if err != nil {
		return fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	globalDB = &Database{db: db}

	if err := globalDB.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := globalDB.seedEventTypes(); err != nil {
		return fmt.Errorf("failed to seed event types: %w", err)
	}

	return nil
}

// GetDB returns the global database instance
func GetDB() *Database {
	if globalDB != nil && globalDB.db != nil {
		if err := globalDB.db.Ping(); err != nil {
			return nil
		}
	}
	return globalDB
}

// IsConnected checks if database connection is alive
func IsConnected() bool {
	if globalDB == nil || globalDB.db == nil {
		return false
	}
	return globalDB.db.Ping() == nil
}

// Close closes the database connection
func Close() error {
	if globalDB != nil && globalDB.db != nil {
		return globalDB.db.Close()
	}
	return nil
}

// createTables creates all necessary database tables
func (d *Database) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS event_types (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		emoji TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS guild_config (
		guild_id TEXT PRIMARY KEY,
		panic_mode INTEGER DEFAULT 0,
		log_channel_id TEXT DEFAULT '',
		enabled_events TEXT DEFAULT '',
		created_at INTEGER DEFAULT 0,
		updated_at INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS event_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		event_type INTEGER NOT NULL,
		actor_id TEXT NOT NULL,
		target_id TEXT DEFAULT '',
		detection_speed_us INTEGER NOT NULL,
		action_taken TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		FOREIGN KEY (event_type) REFERENCES event_types(id)
	);

	CREATE INDEX IF NOT EXISTS idx_event_logs_guild ON event_logs(guild_id);
	CREATE INDEX IF NOT EXISTS idx_event_logs_timestamp ON event_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_guild_config_guild ON guild_config(guild_id);

	CREATE TABLE IF NOT EXISTS banned_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		reason TEXT NOT NULL,
		banned_at INTEGER NOT NULL,
		banned_by TEXT NOT NULL,
		is_bot INTEGER DEFAULT 0,
		added_by TEXT DEFAULT '',
		UNIQUE(guild_id, user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_banned_users_guild ON banned_users(guild_id);
	CREATE INDEX IF NOT EXISTS idx_banned_users_user ON banned_users(user_id);

	CREATE TABLE IF NOT EXISTS event_limits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		event_type INTEGER NOT NULL,
		max_actions INTEGER DEFAULT 3,
		time_window INTEGER DEFAULT 10,
		punishment TEXT DEFAULT 'ban',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		UNIQUE(guild_id, event_type),
		FOREIGN KEY (event_type) REFERENCES event_types(id)
	);

	CREATE INDEX IF NOT EXISTS idx_event_limits_guild ON event_limits(guild_id);

	CREATE TABLE IF NOT EXISTS whitelist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		target_id TEXT NOT NULL,
		target_type TEXT NOT NULL,
		event_type INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		UNIQUE(guild_id, target_id, event_type)
	);

	CREATE INDEX IF NOT EXISTS idx_whitelist_guild ON whitelist(guild_id);
	CREATE INDEX IF NOT EXISTS idx_whitelist_target ON whitelist(guild_id, target_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// seedEventTypes populates the event_types table with all 26 event types
func (d *Database) seedEventTypes() error {
	eventTypes := []struct {
		id          int
		name        string
		description string
		emoji       string
	}{
		{1, "anti_ban", "Anti Ban", "🔨"},
		{2, "anti_unban", "Anti Unban", "🔓"},
		{3, "anti_kick", "Anti Kick", "👢"},
		{4, "anti_bot", "Anti Bot", "🤖"},
		{5, "anti_channel_create", "Anti Channel Create", "➕"},
		{6, "anti_channel_delete", "Anti Channel Delete", "🗑️"},
		{7, "anti_channel_update", "Anti Channel Update", "✏️"},
		{8, "anti_emoji_sticker_create", "Anti Emoji/Sticker Create", "😀"},
		{9, "anti_emoji_sticker_delete", "Anti Emoji/Sticker Delete", "🚫"},
		{10, "anti_emoji_sticker_update", "Anti Emoji/Sticker Update", "🔄"},
		{11, "anti_everyone_here_ping", "Anti Everyone/Here Ping", "📢"},
		{12, "anti_link_role", "Anti Link Role", "🔗"},
		{13, "anti_role_create", "Anti Role Create", "🎭"},
		{14, "anti_role_delete", "Anti Role Delete", "❌"},
		{15, "anti_role_update", "Anti Role Update", "🔧"},
		{16, "anti_role_ping", "Anti Role Ping", "🔔"},
		{17, "anti_member_update", "Anti Member Update", "👤"},
		{18, "anti_integration", "Anti Integration", "🔌"},
		{19, "anti_server_update", "Anti Server Update", "⚙️"},
		{20, "anti_automod_rule_create", "Anti Automod Rule Create", "🛡️"},
		{21, "anti_automod_rule_update", "Anti Automod Rule Update", "🔐"},
		{22, "anti_automod_rule_delete", "Anti Automod Rule Delete", "🗝️"},
		{23, "anti_guild_event_create", "Anti Guild Event Create", "📅"},
		{24, "anti_guild_event_update", "Anti Guild Event Update", "📝"},
		{25, "anti_guild_event_delete", "Anti Guild Event Delete", "🗓️"},
		{26, "anti_webhook", "Anti Webhook", "🪝"},
	}

	for _, et := range eventTypes {
		_, err := d.db.Exec(
			`INSERT OR IGNORE INTO event_types (id, name, description, emoji) VALUES (?, ?, ?, ?)`,
			et.id, et.name, et.description, et.emoji,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetGuildConfig retrieves guild configuration
func (d *Database) GetGuildConfig(guildID string) (*GuildConfig, error) {
	var config GuildConfig
	err := d.db.QueryRow(
		`SELECT guild_id, panic_mode, log_channel_id, enabled_events, created_at, updated_at 
		 FROM guild_config WHERE guild_id = ?`,
		guildID,
	).Scan(&config.GuildID, &config.PanicMode, &config.LogChannelID, &config.EnabledEvents, &config.CreatedAt, &config.UpdatedAt)

	if err == sql.ErrNoRows {
		// Return default config if not found
		return &GuildConfig{
			GuildID:       guildID,
			PanicMode:     false,
			LogChannelID:  "",
			EnabledEvents: "",
			CreatedAt:     time.Now().Unix(),
			UpdatedAt:     time.Now().Unix(),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// UpsertGuildConfig creates or updates guild configuration
func (d *Database) UpsertGuildConfig(config *GuildConfig) error {
	config.UpdatedAt = time.Now().Unix()
	if config.CreatedAt == 0 {
		config.CreatedAt = time.Now().Unix()
	}

	_, err := d.db.Exec(
		`INSERT OR REPLACE INTO guild_config (guild_id, panic_mode, log_channel_id, enabled_events, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		config.GuildID, config.PanicMode, config.LogChannelID, config.EnabledEvents, config.CreatedAt, config.UpdatedAt,
	)

	return err
}

// LogEvent logs an event to the database
func (d *Database) LogEvent(log *EventLog) error {
	log.Timestamp = time.Now().Unix()

	_, err := d.db.Exec(
		`INSERT INTO event_logs (guild_id, event_type, actor_id, target_id, detection_speed_us, action_taken, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		log.GuildID, log.EventType, log.ActorID, log.TargetID, log.DetectionSpeedUS, log.ActionTaken, log.Timestamp,
	)

	return err
}

// GetRecentLogs retrieves recent event logs for a guild
func (d *Database) GetRecentLogs(guildID string, limit int) ([]*EventLog, error) {
	rows, err := d.db.Query(
		`SELECT id, guild_id, event_type, actor_id, target_id, detection_speed_us, action_taken, timestamp
		 FROM event_logs WHERE guild_id = ? ORDER BY timestamp DESC LIMIT ?`,
		guildID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*EventLog
	for rows.Next() {
		var log EventLog
		if err := rows.Scan(&log.ID, &log.GuildID, &log.EventType, &log.ActorID, &log.TargetID, &log.DetectionSpeedUS, &log.ActionTaken, &log.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

// GetEventTypes retrieves all event types
func (d *Database) GetEventTypes() ([]*EventType, error) {
	rows, err := d.db.Query(`SELECT id, name, description, emoji FROM event_types ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []*EventType
	for rows.Next() {
		var t EventType
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Emoji); err != nil {
			return nil, err
		}
		types = append(types, &t)
	}

	return types, rows.Err()
}

// AddBannedUser adds a user to the banned list
func (d *Database) AddBannedUser(guildID, userID, reason, bannedBy string, isBot bool, addedBy string) error {
	_, err := d.db.Exec(
		`INSERT OR REPLACE INTO banned_users (guild_id, user_id, reason, banned_at, banned_by, is_bot, added_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		guildID, userID, reason, time.Now().Unix(), bannedBy, isBot, addedBy,
	)
	return err
}

// IsBannedUser checks if a user is in the banned list
func (d *Database) IsBannedUser(guildID, userID string) bool {
	var count int
	err := d.db.QueryRow(
		`SELECT COUNT(*) FROM banned_users WHERE guild_id = ? AND user_id = ?`,
		guildID, userID,
	).Scan(&count)
	return err == nil && count > 0
}

// GetBannedUser retrieves a specific banned user record
func (d *Database) GetBannedUser(guildID, userID string) (*BannedUser, error) {
	var user BannedUser
	var isBot int
	err := d.db.QueryRow(
		`SELECT id, guild_id, user_id, reason, banned_at, banned_by, is_bot, added_by
		 FROM banned_users WHERE guild_id = ? AND user_id = ?`,
		guildID, userID,
	).Scan(&user.ID, &user.GuildID, &user.UserID, &user.Reason, &user.BannedAt, &user.BannedBy, &isBot, &user.AddedBy)

	if err != nil {
		return nil, err
	}

	user.IsBot = isBot != 0
	return &user, nil
}

// RemoveBannedUser removes a user from the banned list
func (d *Database) RemoveBannedUser(guildID, userID string) error {
	_, err := d.db.Exec(
		`DELETE FROM banned_users WHERE guild_id = ? AND user_id = ?`,
		guildID, userID,
	)
	return err
}

// ===== Event Limits =====

// UpsertEventLimit creates or updates an event limit
func (d *Database) UpsertEventLimit(limit *EventLimit) error {
	now := time.Now().Unix()
	limit.UpdatedAt = now
	if limit.CreatedAt == 0 {
		limit.CreatedAt = now
	}

	_, err := d.db.Exec(
		`INSERT OR REPLACE INTO event_limits (guild_id, event_type, max_actions, time_window, punishment, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		limit.GuildID, limit.EventType, limit.MaxActions, limit.TimeWindow, limit.Punishment, limit.CreatedAt, limit.UpdatedAt,
	)
	return err
}

// GetEventLimit retrieves an event limit for a specific event type
func (d *Database) GetEventLimit(guildID string, eventType int) (*EventLimit, error) {
	var limit EventLimit
	err := d.db.QueryRow(
		`SELECT id, guild_id, event_type, max_actions, time_window, punishment, created_at, updated_at
		 FROM event_limits WHERE guild_id = ? AND event_type = ?`,
		guildID, eventType,
	).Scan(&limit.ID, &limit.GuildID, &limit.EventType, &limit.MaxActions, &limit.TimeWindow, &limit.Punishment, &limit.CreatedAt, &limit.UpdatedAt)

	if err == sql.ErrNoRows {
		// Return default limits
		return &EventLimit{
			GuildID:    guildID,
			EventType:  eventType,
			MaxActions: 3,
			TimeWindow: 10,
			Punishment: "ban",
		}, nil
	}

	return &limit, err
}

// GetAllEventLimits retrieves all event limits for a guild
func (d *Database) GetAllEventLimits(guildID string) ([]*EventLimit, error) {
	rows, err := d.db.Query(
		`SELECT id, guild_id, event_type, max_actions, time_window, punishment, created_at, updated_at
		 FROM event_limits WHERE guild_id = ?`,
		guildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var limits []*EventLimit
	for rows.Next() {
		var limit EventLimit
		if err := rows.Scan(&limit.ID, &limit.GuildID, &limit.EventType, &limit.MaxActions, &limit.TimeWindow, &limit.Punishment, &limit.CreatedAt, &limit.UpdatedAt); err != nil {
			return nil, err
		}
		limits = append(limits, &limit)
	}

	return limits, rows.Err()
}

// ===== Whitelist =====

// AddWhitelist adds a user/role to the whitelist
func (d *Database) AddWhitelist(guildID, targetID, targetType string, eventType int) error {
	_, err := d.db.Exec(
		`INSERT OR IGNORE INTO whitelist (guild_id, target_id, target_type, event_type, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		guildID, targetID, targetType, eventType, time.Now().Unix(),
	)
	return err
}

// RemoveWhitelist removes a user/role from the whitelist
func (d *Database) RemoveWhitelist(guildID, targetID string, eventType int) error {
	_, err := d.db.Exec(
		`DELETE FROM whitelist WHERE guild_id = ? AND target_id = ? AND event_type = ?`,
		guildID, targetID, eventType,
	)
	return err
}

// IsWhitelisted checks if a user/role is whitelisted for an event
func (d *Database) IsWhitelisted(guildID, targetID string, eventType int) bool {
	var count int
	// Check for specific event or all events (event_type = 0)
	err := d.db.QueryRow(
		`SELECT COUNT(*) FROM whitelist WHERE guild_id = ? AND target_id = ? AND (event_type = ? OR event_type = 0)`,
		guildID, targetID, eventType,
	).Scan(&count)
	return err == nil && count > 0
}

// GetWhitelistByTarget retrieves all whitelist entries for a specific target
func (d *Database) GetWhitelistByTarget(guildID, targetID string) ([]*Whitelist, error) {
	rows, err := d.db.Query(
		`SELECT id, guild_id, target_id, target_type, event_type, created_at
		 FROM whitelist WHERE guild_id = ? AND target_id = ?`,
		guildID, targetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*Whitelist
	for rows.Next() {
		var entry Whitelist
		if err := rows.Scan(&entry.ID, &entry.GuildID, &entry.TargetID, &entry.TargetType, &entry.EventType, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}

// GetAllWhitelist retrieves all whitelist entries for a guild
func (d *Database) GetAllWhitelist(guildID string) ([]*Whitelist, error) {
	rows, err := d.db.Query(
		`SELECT id, guild_id, target_id, target_type, event_type, created_at
		 FROM whitelist WHERE guild_id = ?`,
		guildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*Whitelist
	for rows.Next() {
		var entry Whitelist
		if err := rows.Scan(&entry.ID, &entry.GuildID, &entry.TargetID, &entry.TargetType, &entry.EventType, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}

// GetBannedUsers retrieves all banned users for a guild
func (d *Database) GetBannedUsers(guildID string) ([]*BannedUser, error) {
	rows, err := d.db.Query(
		`SELECT id, guild_id, user_id, reason, banned_at, banned_by, is_bot, added_by
		 FROM banned_users WHERE guild_id = ? ORDER BY banned_at DESC`,
		guildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*BannedUser
	for rows.Next() {
		var user BannedUser
		var isBot int
		if err := rows.Scan(&user.ID, &user.GuildID, &user.UserID, &user.Reason, &user.BannedAt, &user.BannedBy, &isBot, &user.AddedBy); err != nil {
			return nil, err
		}
		user.IsBot = isBot != 0
		users = append(users, &user)
	}

	return users, rows.Err()
}
