package database

import (
	"fmt"
	"strconv"
	"strings"

	"go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/pkg/util"
)

// SyncGuildStateFromDB loads guild configuration from database and syncs to in-memory store
func (d *Database) SyncGuildStateFromDB(guildID string) error {
	guildConfig, err := d.GetGuildConfig(guildID)
	if err != nil {
		return fmt.Errorf("failed to load guild config: %w", err)
	}

	// Convert string guild ID to uint64
	guildIDNum, err := util.StringToUint64(guildID)
	if err != nil {
		return fmt.Errorf("invalid guild ID: %w", err)
	}

	// Get or create profile store
	store := config.GetProfileStore()
	profile := store.GetOrCreate(guildIDNum)

	// Sync panic mode from database
	profile.PanicMode = guildConfig.PanicMode

	// Sync enabled state - if anti-nuke has events enabled, it's considered enabled
	profile.Enabled = guildConfig.EnabledEvents != ""

	// Update the profile in store
	store.Set(profile)

	return nil
}

// SyncAllGuildsFromDB loads all guild configurations from database and syncs to in-memory store
func (d *Database) SyncAllGuildsFromDB() error {
	rows, err := d.db.Query(`SELECT guild_id FROM guild_config`)
	if err != nil {
		return fmt.Errorf("failed to query guild configs: %w", err)
	}
	defer rows.Close()

	var guildIDs []string
	for rows.Next() {
		var guildID string
		if err := rows.Scan(&guildID); err != nil {
			return fmt.Errorf("failed to scan guild ID: %w", err)
		}
		guildIDs = append(guildIDs, guildID)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Sync each guild
	for _, guildID := range guildIDs {
		if err := d.SyncGuildStateFromDB(guildID); err != nil {
			// Log error but continue with other guilds
			fmt.Printf("Warning: Failed to sync guild %s: %v\n", guildID, err)
		}
	}

	return nil
}

// InitializeGuildWithDefaults creates a new guild config with all events enabled by default
func (d *Database) InitializeGuildWithDefaults(guildID string) error {
	// Check if guild already exists
	existing, err := d.GetGuildConfig(guildID)
	if err != nil {
		return err
	}

	// If guild already has events configured, don't override
	if existing.EnabledEvents != "" {
		return nil
	}

	// Get all event types
	eventTypes, err := d.GetEventTypes()
	if err != nil {
		return fmt.Errorf("failed to get event types: %w", err)
	}

	// Enable all events by default
	enabledEvents := make([]string, len(eventTypes))
	for i, et := range eventTypes {
		enabledEvents[i] = strconv.Itoa(et.ID)
	}

	// Update config with all events enabled
	existing.EnabledEvents = strings.Join(enabledEvents, ",")
	existing.PanicMode = false // Panic mode off by default

	if err := d.UpsertGuildConfig(existing); err != nil {
		return fmt.Errorf("failed to initialize guild config: %w", err)
	}

	// Sync to in-memory store
	return d.SyncGuildStateFromDB(guildID)
}

// EnsureGuildConfigExists ensures a guild has a config, creating default if needed
func (d *Database) EnsureGuildConfigExists(guildID string) error {
	_, err := d.GetGuildConfig(guildID)
	if err != nil {
		return err
	}

	// Just sync existing config to memory - DO NOT auto-enable events
	// User must explicitly enable events via /antinuke enable command
	return d.SyncGuildStateFromDB(guildID)
}

// SyncWhitelistToMemory syncs whitelist from database to in-memory profile store
func (d *Database) SyncWhitelistToMemory(guildID string) error {
	guildIDNum, err := util.StringToUint64(guildID)
	if err != nil {
		return fmt.Errorf("invalid guild ID: %w", err)
	}

	// Get all whitelist entries for this guild
	whitelists, err := d.GetAllWhitelist(guildID)
	if err != nil {
		return fmt.Errorf("failed to get whitelist: %w", err)
	}

	// Get profile store
	store := config.GetProfileStore()
	profile := store.GetOrCreate(guildIDNum)

	// Clear existing whitelist
	profile.Whitelist = make([]uint64, 0)

	// Add all whitelisted users/roles
	for _, w := range whitelists {
		targetIDNum, err := util.StringToUint64(w.TargetID)
		if err != nil {
			continue
		}
		// Add to in-memory whitelist
		profile.Whitelist = append(profile.Whitelist, targetIDNum)
	}

	// Update store
	store.Set(profile)
	return nil
}

// SyncThresholdsToMemory syncs event limits/thresholds to correlator for real-time enforcement
func (d *Database) SyncThresholdsToMemory(guildID string) error {
	// Get all event limits for this guild
	limits, err := d.GetAllEventLimits(guildID)
	if err != nil {
		return fmt.Errorf("failed to get event limits: %w", err)
	}

	guildIDNum, err := util.StringToUint64(guildID)
	if err != nil {
		return fmt.Errorf("invalid guild ID: %w", err)
	}

	// Get profile store
	store := config.GetProfileStore()
	profile := store.GetOrCreate(guildIDNum)

	// Update custom thresholds if configured
	if len(limits) > 0 {
		// Create custom threshold matrix if needed
		if profile.CustomThresholds == nil {
			profile.CustomThresholds = &config.ThresholdMatrix{}
		}
		// Thresholds are automatically loaded by correlator from database
		// This sync just ensures the profile is aware of custom limits
	}

	store.Set(profile)
	return nil
}
