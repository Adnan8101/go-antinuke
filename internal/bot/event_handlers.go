package bot

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/internal/ingest"
	"go-antinuke-2.0/internal/logging"
	"go-antinuke-2.0/internal/state"

	"github.com/bwmarrin/discordgo"
)

// auditLogCache stores recent audit log entries to correlate with events
type auditLogCache struct {
	mu      sync.RWMutex
	entries map[string]*auditCacheEntry
}

type auditCacheEntry struct {
	actorID   uint64
	targetID  uint64
	action    int
	timestamp time.Time
}

var (
	auditCache = &auditLogCache{
		entries: make(map[string]*auditCacheEntry),
	}
	cacheTTL = 5 * time.Second
)

func (c *auditLogCache) Store(guildID string, action int, actorID, targetID uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := guildID + ":" + strconv.Itoa(action)
	c.entries[key] = &auditCacheEntry{
		actorID:   actorID,
		targetID:  targetID,
		action:    action,
		timestamp: time.Now(),
	}

	// Cleanup old entries
	for k, v := range c.entries {
		if time.Since(v.timestamp) > cacheTTL {
			delete(c.entries, k)
		}
	}
}

func (c *auditLogCache) Get(guildID string, action int) (uint64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := guildID + ":" + strconv.Itoa(action)
	if entry, exists := c.entries[key]; exists {
		if time.Since(entry.timestamp) < cacheTTL {
			return entry.actorID, true
		}
	}
	return 0, false
}

// fetchActorFromAuditLog fetches the most recent audit log entry for a specific action
func fetchActorFromAuditLog(sess *discordgo.Session, guildID string, actionType int, targetID uint64) uint64 {
	// First check cache
	if actorID, found := auditCache.Get(guildID, actionType); found {
		return actorID
	}

	// Fetch from Discord API
	audit, err := sess.GuildAuditLog(guildID, "", "", actionType, 1)
	if err != nil {
		logging.Warn("Failed to fetch audit log for guild %s action %d: %v", guildID, actionType, err)
		return 0
	}

	if len(audit.AuditLogEntries) == 0 {
		return 0
	}

	// Get the most recent entry
	entry := audit.AuditLogEntries[0]

	// Check if the actor is a bot - skip bot actions entirely
	if len(audit.Users) > 0 {
		for _, user := range audit.Users {
			if user.ID == entry.UserID && user.Bot {
				logging.Debug("[AUDIT] Skipping action %d by bot user %s", actionType, user.Username)
				return 0
			}
		}
	}

	actorID, _ := strconv.ParseUint(entry.UserID, 10, 64)

	// Cache it for future use
	auditCache.Store(guildID, actionType, actorID, targetID)

	return actorID
}

// SetupEventHandlers configures Discord event handlers to feed the ring buffer
func (s *Session) SetupEventHandlers(ringBuffer *ingest.RingBuffer) {
	logging.Info("Setting up Discord event handlers...")

	// Handle bot joining new guilds - auto-initialize with all events enabled
	s.discord.AddHandler(func(sess *discordgo.Session, g *discordgo.GuildCreate) {
		logging.Info("Bot joined/loaded guild: %s (ID: %s)", g.Name, g.ID)

		// Clear all actor state for this guild when bot is re-added
		// This ensures previously banned users can be detected again
		guildID, _ := strconv.ParseUint(g.ID, 10, 64)
		state.ClearGuildActorStates(guildID)
		logging.Info("✓ Cleared actor state for guild %s", g.ID)

		// Store owner ID in guild profile
		ownerID, _ := strconv.ParseUint(g.OwnerID, 10, 64)
		profile := config.GetProfileStore().GetOrCreate(guildID)
		profile.OwnerID = ownerID
		logging.Info("✓ Set owner ID %d for guild %s", ownerID, g.ID)

		// Import database package at runtime
		// Auto-initialize guild config with all events enabled
		// This is handled by EnsureGuildConfigExists in the sync process
	})

	// Handle bot ready - clear state for all guilds
	s.discord.AddHandler(func(sess *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("[BOT] Ready event fired! Connected as %s\n", r.User.Username)
		fmt.Printf("[BOT] Clearing actor state for %d guilds...\n", len(r.Guilds))
		logging.Info("Bot ready! Connected as %s", r.User.Username)
		logging.Info("Clearing actor state for all guilds on reconnect...")

		// Clear state for all guilds the bot is in
		for _, guild := range r.Guilds {
			guildID, _ := strconv.ParseUint(guild.ID, 10, 64)
			state.ClearGuildActorStates(guildID)
			fmt.Printf("[BOT] ✓ Cleared actor state for guild %s\n", guild.ID)
			logging.Info("✓ Cleared actor state for guild %s", guild.ID)
		}
		fmt.Printf("[BOT] State clearing complete!\n")
	})

	// Handle Guild Ban Remove (Unban) - Clear actor state so they can be detected again if they return
	s.discord.AddHandler(func(sess *discordgo.Session, b *discordgo.GuildBanRemove) {
		if b.GuildID == "" {
			return
		}
		actorID, _ := strconv.ParseUint(b.User.ID, 10, 64)
		state.ClearActorState(actorID)
		logging.Info("[STATE] Cleared actor state for unbanned user %s in guild %s", b.User.ID, b.GuildID)
	})

	// Handle Guild Member Add (Join) - Smart bot/user rejoin logic
	s.discord.AddHandler(func(sess *discordgo.Session, m *discordgo.GuildMemberAdd) {
		if m.GuildID == "" {
			return
		}

		userID, _ := strconv.ParseUint(m.User.ID, 10, 64)
		guildID, _ := strconv.ParseUint(m.GuildID, 10, 64)

		// Check if this user was previously banned by the bot
		if db := database.GetDB(); db != nil {
			if db.IsBannedUser(m.GuildID, m.User.ID) {
				// Get the banned user record (for future use if needed)
				_, err := db.GetBannedUser(m.GuildID, m.User.ID)
				if err != nil {
					logging.Error("Failed to fetch banned user record: %v", err)
					return
				}

				// Check if this is a bot
				if m.User.Bot {
					logging.Info("[BANNED BOT REJOINED] Bot %s (%s) rejoined guild %s", m.User.Username, m.User.ID, m.GuildID)

					// Find who added this bot by checking audit logs
					audit, err := sess.GuildAuditLog(m.GuildID, "", "", 28, 5) // 28 = BOT_ADD
					if err != nil {
						logging.Error("Failed to fetch audit log for bot add: %v", err)
						// Re-ban the bot immediately as a safety measure
						go sess.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Previously banned bot - automatic re-ban", 0)
						return
					}

					// Find the most recent bot add entry for this bot
					var adderID string
					for _, entry := range audit.AuditLogEntries {
						if entry.TargetID == m.User.ID {
							adderID = entry.UserID
							break
						}
					}

					if adderID == "" {
						logging.Warn("Could not determine who added bot %s", m.User.ID)
						// Re-ban as safety measure
						go sess.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Previously banned bot - automatic re-ban", 0)
						return
					}

					adderIDNum, _ := strconv.ParseUint(adderID, 10, 64)

					// Get guild profile to check owner and whitelist
					profile := config.GetProfileStore().Get(guildID)
					isOwner := profile != nil && profile.OwnerID == adderIDNum
					isWhitelisted := profile != nil && config.GetProfileStore().IsWhitelisted(guildID, adderIDNum)

					if isOwner || isWhitelisted {
						// Owner or whitelisted user added the bot - ALLOW IT
						logging.Info("[✓ BOT ALLOWED] Bot %s added by %s %s - Clearing ban record",
							m.User.Username,
							map[bool]string{true: "owner", false: "whitelisted user"}[isOwner],
							adderID)

						// Remove from banned list and clear state for fresh start
						db.RemoveBannedUser(m.GuildID, m.User.ID)
						state.ClearActorState(userID)

						// Log this action
						logging.Info("[STATE] Bot %s allowed and given fresh start by %s", m.User.ID, adderID)
						return
					} else {
						// Unauthorized person added a previously banned bot - BAN BOTH
						logging.Warn("[❌ UNAUTHORIZED BOT ADD] Bot %s added by unauthorized user %s - Banning both", m.User.Username, adderID)

						// Ban the bot
						go func() {
							err := sess.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Previously banned bot re-added by unauthorized user", 0)
							if err != nil {
								logging.Error("Failed to re-ban bot %s: %v", m.User.ID, err)
							} else {
								logging.Info("[✓ BOT RE-BANNED] %s", m.User.ID)
							}
						}()

						// Ban the person who added it
						go func() {
							err := sess.GuildBanCreateWithReason(m.GuildID, adderID, "Added previously banned bot - security violation", 0)
							if err != nil {
								logging.Error("Failed to ban user %s who added banned bot: %v", adderID, err)
							} else {
								logging.Info("[✓ ADDER BANNED] User %s banned for adding banned bot", adderID)
								// Add to database
								if db := database.GetDB(); db != nil {
									db.AddBannedUser(m.GuildID, adderID, "Added previously banned bot", "antinuke-bot", false, "")
								}
							}
						}()

						// Try to send DM to the banned person
						go func() {
							channel, err := sess.UserChannelCreate(adderID)
							if err == nil {
								sess.ChannelMessageSend(channel.ID, fmt.Sprintf(
									"⚠️ **Security Violation in %s**\\n\\n"+
										"You were banned for adding a previously banned bot (%s).\\n"+
										"This bot was flagged by the antinuke system for malicious behavior.\\n\\n"+
										"Only the server owner or whitelisted users can re-add banned bots.",
									m.GuildID, m.User.Username))
							}
						}()

						return
					}
				} else {
					// Human user rejoined - give them a fresh start
					logging.Info("[BANNED USER REJOINED] User %s rejoined guild %s - Clearing ban record for fresh start", m.User.ID, m.GuildID)

					// Remove from banned list to give them another chance
					db.RemoveBannedUser(m.GuildID, m.User.ID)
					state.ClearActorState(userID)

					logging.Info("[✓ FRESH START] User %s given clean slate in guild %s - will be monitored for new violations", m.User.ID, m.GuildID)
					return
				}
			}

			// User not in banned list - cleanup and fresh start
			db.RemoveBannedUser(m.GuildID, m.User.ID)
		}

		// Clear actor state for fresh start
		state.ClearActorState(userID)
		logging.Info("[STATE] Cleared actor state for joining %s %s in guild %s",
			map[bool]string{true: "bot", false: "user"}[m.User.Bot],
			m.User.ID, m.GuildID)
	})

	// CRITICAL: GuildAuditLogEntryCreate - This captures WHO did the action
	s.discord.AddHandler(func(sess *discordgo.Session, audit *discordgo.GuildAuditLogEntryCreate) {
		startTime := time.Now() // Track detection latency

		if audit.GuildID == "" {
			return
		}

		actorID, _ := strconv.ParseUint(audit.UserID, 10, 64)
		targetID := uint64(0)
		if audit.TargetID != "" {
			targetID, _ = strconv.ParseUint(audit.TargetID, 10, 64)
		}

		// Get action type value
		actionType := 0
		if audit.ActionType != nil {
			actionType = int(*audit.ActionType)
		}

		// Store in cache for correlation with direct events
		auditCache.Store(audit.GuildID, actionType, actorID, targetID)

		logging.Debug("[AUDIT] Action %d by user %d in guild %s | Latency: %d µs",
			actionType, actorID, audit.GuildID, time.Since(startTime).Microseconds())
	})

	// DIRECT EVENT HANDLERS - These fire immediately, we correlate with audit logs for actor ID

	// Channel Create - Fetch audit logs immediately to get actor
	s.discord.AddHandler(func(sess *discordgo.Session, c *discordgo.ChannelCreate) {
		startTime := time.Now()

		if c.GuildID == "" {
			return
		}

		guildID, _ := strconv.ParseUint(c.GuildID, 10, 64)
		channelIDNum, _ := strconv.ParseUint(c.ID, 10, 64)

		// Fetch audit log entry for this specific action
		actorID := fetchActorFromAuditLog(sess, c.GuildID, 10, channelIDNum) // 10 = CHANNEL_CREATE

		if actorID == 0 {
			logging.Warn("[EVENT] Channel create but no actor ID: %s", c.ID)
			return
		}

		event := ingest.CreateEvent(
			ingest.EventTypeChannelCreate,
			guildID,
			actorID,
			channelIDNum,
			0,
		)
		ringBuffer.Enqueue(event)

		latencyUs := time.Since(startTime).Microseconds()
		logging.Info("[EVENT] Channel create: %s by actor %d | Latency: %d µs", c.ID, actorID, latencyUs)
	})

	// Channel Delete
	s.discord.AddHandler(func(sess *discordgo.Session, c *discordgo.ChannelDelete) {
		startTime := time.Now()

		if c.GuildID == "" {
			return
		}

		guildID, _ := strconv.ParseUint(c.GuildID, 10, 64)
		channelIDNum, _ := strconv.ParseUint(c.ID, 10, 64)

		actorID := fetchActorFromAuditLog(sess, c.GuildID, 12, channelIDNum) // 12 = CHANNEL_DELETE

		if actorID == 0 {
			logging.Warn("[EVENT] Channel delete but no actor ID: %s", c.ID)
			return
		}

		event := ingest.CreateEvent(
			ingest.EventTypeChannelDelete,
			guildID,
			actorID,
			channelIDNum,
			0,
		)
		ringBuffer.Enqueue(event)

		latencyUs := time.Since(startTime).Microseconds()
		logging.Info("[EVENT] Channel delete: %s by actor %d | Latency: %d µs", c.ID, actorID, latencyUs)
	})

	// Role Create
	s.discord.AddHandler(func(sess *discordgo.Session, r *discordgo.GuildRoleCreate) {
		startTime := time.Now()

		if r.GuildID == "" {
			return
		}

		// Skip managed roles (bot roles, integration roles, booster roles, etc.)
		// These are created automatically by Discord and are not malicious
		if r.Role.Managed {
			logging.Debug("[EVENT] Skipping managed role create: %s (likely bot/integration role)", r.Role.ID)
			return
		}

		guildID, _ := strconv.ParseUint(r.GuildID, 10, 64)
		roleIDNum, _ := strconv.ParseUint(r.Role.ID, 10, 64)

		actorID := fetchActorFromAuditLog(sess, r.GuildID, 30, roleIDNum) // 30 = ROLE_CREATE

		if actorID == 0 {
			logging.Warn("[EVENT] Role create but no actor ID: %s", r.Role.ID)
			return
		}

		event := ingest.CreateEvent(
			ingest.EventTypeRoleCreate,
			guildID,
			actorID,
			roleIDNum,
			0,
		)
		ringBuffer.Enqueue(event)

		latencyUs := time.Since(startTime).Microseconds()
		logging.Info("[EVENT] Role create: %s by actor %d | Latency: %d µs", r.Role.ID, actorID, latencyUs)
	})

	// Role Delete
	s.discord.AddHandler(func(sess *discordgo.Session, r *discordgo.GuildRoleDelete) {
		startTime := time.Now()

		if r.GuildID == "" {
			return
		}

		guildID, _ := strconv.ParseUint(r.GuildID, 10, 64)
		roleIDNum, _ := strconv.ParseUint(r.RoleID, 10, 64)

		actorID := fetchActorFromAuditLog(sess, r.GuildID, 32, roleIDNum) // 32 = ROLE_DELETE

		if actorID == 0 {
			logging.Warn("[EVENT] Role delete but no actor ID: %s", r.RoleID)
			return
		}

		event := ingest.CreateEvent(
			ingest.EventTypeRoleDelete,
			guildID,
			actorID,
			roleIDNum,
			0,
		)
		ringBuffer.Enqueue(event)

		latencyUs := time.Since(startTime).Microseconds()
		logging.Info("[EVENT] Role delete: %s by actor %d | Latency: %d µs", r.RoleID, actorID, latencyUs)
	})

	logging.Info("Discord event handlers configured successfully (Direct Events + Audit Log Fetch)")
}

// mapAuditActionToEventType maps Discord audit log action types to internal event types
func mapAuditActionToEventType(action int) uint8 {
	switch action {
	case 10: // CHANNEL_CREATE
		return ingest.EventTypeChannelCreate
	case 12: // CHANNEL_DELETE
		return ingest.EventTypeChannelDelete
	case 11: // CHANNEL_UPDATE
		return ingest.EventTypeChannelUpdate
	case 30: // ROLE_CREATE
		return ingest.EventTypeRoleCreate
	case 32: // ROLE_DELETE
		return ingest.EventTypeRoleDelete
	case 31: // ROLE_UPDATE
		return ingest.EventTypeRoleUpdate
	case 20: // MEMBER_KICK
		return ingest.EventTypeKick
	case 22: // MEMBER_BAN_ADD
		return ingest.EventTypeBan
	case 23: // MEMBER_BAN_REMOVE
		return ingest.EventTypeUnban
	case 24: // MEMBER_UPDATE
		return ingest.EventTypeMemberUpdate
	case 50: // WEBHOOK_CREATE
		return ingest.EventTypeWebhook
	case 51: // WEBHOOK_UPDATE
		return ingest.EventTypeWebhook
	case 52: // WEBHOOK_DELETE
		return ingest.EventTypeWebhook
	case 60: // EMOJI_CREATE
		return ingest.EventTypeEmojiStickerCreate
	case 61: // EMOJI_UPDATE
		return ingest.EventTypeEmojiStickerUpdate
	case 62: // EMOJI_DELETE
		return ingest.EventTypeEmojiStickerDelete
	case 1: // GUILD_UPDATE
		return ingest.EventTypeServerUpdate
	case 80: // INTEGRATION_CREATE
		return ingest.EventTypeIntegration
	case 81: // INTEGRATION_UPDATE
		return ingest.EventTypeIntegration
	case 82: // INTEGRATION_DELETE
		return ingest.EventTypeIntegration
	default:
		return 0 // Unknown/unsupported event type
	}
}
