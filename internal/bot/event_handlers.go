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
	cacheTTL = 10 * time.Second
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

	// Only cleanup every 100th entry to reduce overhead
	if len(c.entries)%100 == 0 {
		now := time.Now()
		for k, v := range c.entries {
			if now.Sub(v.timestamp) > cacheTTL {
				delete(c.entries, k)
			}
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

	// FAKE EVENT DETECTION: Detect fake audit log events
	isFakeEvent := false
	eventTypeName := ""

	// Check for fake member-targeted events (kick, ban, unban)
	if actionType == 20 || actionType == 22 || actionType == 23 { // KICK=20, BAN=22, UNBAN=23
		if entry.TargetID != "" {
			// Check if the target user actually exists in the server
			_, err := sess.GuildMember(guildID, entry.TargetID)
			if err != nil {
				isFakeEvent = true
				eventTypeName = map[int]string{20: "kick", 22: "ban", 23: "unban"}[actionType]
			}
		}
	}

	// Check for fake role events (role create, delete, update)
	if actionType == 30 || actionType == 32 || actionType == 31 { // ROLE_CREATE=30, ROLE_DELETE=32, ROLE_UPDATE=31
		if entry.TargetID != "" {
			// Check if the role actually exists in the server
			roles, err := sess.GuildRoles(guildID)
			if err == nil {
				roleExists := false
				for _, role := range roles {
					if role.ID == entry.TargetID {
						roleExists = true
						break
					}
				}
				if !roleExists {
					isFakeEvent = true
					eventTypeName = map[int]string{30: "role create", 32: "role delete", 31: "role update"}[actionType]
				}
			}
		}
	}

	// Check for fake channel events (channel create, delete, update)
	if actionType == 10 || actionType == 12 || actionType == 11 { // CHANNEL_CREATE=10, CHANNEL_DELETE=12, CHANNEL_UPDATE=11
		if entry.TargetID != "" {
			// Check if the channel actually exists
			_, err := sess.Channel(entry.TargetID)
			if err != nil {
				isFakeEvent = true
				eventTypeName = map[int]string{10: "channel create", 12: "channel delete", 11: "channel update"}[actionType]
			}
		}
	}

	// Check for fake webhook events
	if actionType == 50 || actionType == 51 || actionType == 52 { // WEBHOOK_CREATE=50, WEBHOOK_UPDATE=51, WEBHOOK_DELETE=52
		if entry.TargetID != "" {
			// Webhooks are harder to verify, but we can try to fetch it
			// For now, we'll log but not auto-ban (webhooks get deleted legitimately)
			// Only flag as fake if it's a webhook create/update for non-existent webhook
			if actionType == 50 || actionType == 51 {
				// Try to get webhook (will fail if doesn't exist)
				_, err := sess.Webhook(entry.TargetID)
				if err != nil {
					isFakeEvent = true
					eventTypeName = map[int]string{50: "webhook create", 51: "webhook update", 52: "webhook delete"}[actionType]
				}
			}
		}
	}

	// If fake event detected, ban the perpetrator
	if isFakeEvent {
		logging.Warn("[FAKE EVENT DETECTED] %s (action %d) on non-existent target %s by %s in guild %s",
			eventTypeName, actionType, entry.TargetID, entry.UserID, guildID)

		// Ban the perpetrator immediately
		go func() {
			err := sess.GuildBanCreateWithReason(guildID, entry.UserID,
				fmt.Sprintf("Fake %s event detected - security violation", eventTypeName), 0)
			if err != nil {
				logging.Error("Failed to ban fake event perpetrator %s: %v", entry.UserID, err)
			} else {
				logging.Info("[✓ FAKE EVENT BAN] Banned %s for creating fake %s event", entry.UserID, eventTypeName)

				// Add to database
				if db := database.GetDB(); db != nil {
					db.AddBannedUser(guildID, entry.UserID,
						fmt.Sprintf("Fake %s event creator", eventTypeName),
						"antinuke-bot", false, "")
				}

				// Send log to guild's log channel
				guildIDNum, _ := strconv.ParseUint(guildID, 10, 64)
				_ = guildIDNum // Use database for log channel
				if db := database.GetDB(); db != nil {
					guildConfig, err := db.GetGuildConfig(guildID)
					if err == nil && guildConfig != nil && guildConfig.LogChannelID != "" {
						sess.ChannelMessageSendEmbed(guildConfig.LogChannelID, &discordgo.MessageEmbed{
							Title: "🚨 FAKE EVENT DETECTED & BANNED",
							Description: fmt.Sprintf("User <@%s> created a fake **%s** event targeting non-existent target `%s`\n\n**Action Taken:** Permanently banned\n\n**Reason:** Creating fake audit log events is a serious security violation.",
								entry.UserID,
								eventTypeName,
								entry.TargetID),
							Color:     0xFF0000,
							Timestamp: time.Now().Format(time.RFC3339),
							Footer: &discordgo.MessageEmbedFooter{
								Text: "Antinuke Security System",
							},
						})
					}
				}
			}
		}()

		return 0 // Don't process this fake event further
	}

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

	// Handle Guild Member Add (Join) - Panic mode rejoin logic
	s.discord.AddHandler(func(sess *discordgo.Session, m *discordgo.GuildMemberAdd) {
		if m.GuildID == "" {
			return
		}

		userID, _ := strconv.ParseUint(m.User.ID, 10, 64)
		guildID, _ := strconv.ParseUint(m.GuildID, 10, 64)

		// Check if this is a bot joining
		if m.User.Bot {
			logging.Info("[BOT JOINED] Bot %s (%s) joined guild %s", m.User.Username, m.User.ID, m.GuildID)

			// Find who added this bot by checking audit logs
			audit, err := sess.GuildAuditLog(m.GuildID, "", "", 28, 5) // 28 = BOT_ADD
			if err != nil {
				logging.Error("Failed to fetch audit log for bot add: %v", err)
				// Safety measure: check if bot was previously banned
				if db := database.GetDB(); db != nil && db.IsBannedUser(m.GuildID, m.User.ID) {
					go sess.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Previously banned bot - automatic re-ban", 0)
				}
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
				// Safety measure: check if bot was previously banned
				if db := database.GetDB(); db != nil && db.IsBannedUser(m.GuildID, m.User.ID) {
					go sess.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Previously banned bot - automatic re-ban", 0)
				}
				return
			}

			adderIDNum, _ := strconv.ParseUint(adderID, 10, 64)

			// Get guild profile to check owner and whitelist
			profile := config.GetProfileStore().Get(guildID)
			isOwner := profile != nil && profile.OwnerID == adderIDNum
			isWhitelisted := profile != nil && config.GetProfileStore().IsWhitelisted(guildID, adderIDNum)

			if isOwner || isWhitelisted {
				// Owner or whitelisted user added the bot - ALLOW IT
				logging.Info("[✓ BOT ALLOWED] Bot %s added by %s %s - Allowing and clearing state",
					m.User.Username,
					map[bool]string{true: "owner", false: "whitelisted user"}[isOwner],
					adderID)

				// Remove from banned list if present and clear state for fresh start
				if db := database.GetDB(); db != nil {
					db.RemoveBannedUser(m.GuildID, m.User.ID)
				}
				state.ClearActorState(userID)

				// Log this action
				logging.Info("[STATE] Bot %s allowed and given fresh start by %s", m.User.ID, adderID)
				return
			} else {
				// Unauthorized person added a bot - BAN THE BOT
				logging.Warn("[❌ UNAUTHORIZED BOT ADD] Bot %s added by unauthorized user %s - Banning bot", m.User.Username, adderID)

				// Ban the bot
				go func() {
					err := sess.GuildBanCreateWithReason(m.GuildID, m.User.ID, "Bot added by non-whitelisted user - security policy", 0)
					if err != nil {
						logging.Error("Failed to ban unauthorized bot %s: %v", m.User.ID, err)
					} else {
						logging.Info("[✓ BOT BANNED] %s", m.User.ID)
						// Add to database
						if db := database.GetDB(); db != nil {
							db.AddBannedUser(m.GuildID, m.User.ID, "Added by non-whitelisted user", "antinuke-bot", true, adderID)
						}
					}
				}()

				// Send log to guild's log channel
				go func() {
					// Get log channel from database
					if db := database.GetDB(); db != nil {
						guildConfig, err := db.GetGuildConfig(m.GuildID)
						if err == nil && guildConfig != nil && guildConfig.LogChannelID != "" {
							sess.ChannelMessageSendEmbed(guildConfig.LogChannelID, &discordgo.MessageEmbed{
								Title: "🚨 UNAUTHORIZED BOT DETECTED & BANNED",
								Description: fmt.Sprintf("Bot <@%s> was added by <@%s> who is not authorized.\\n\\n**Action Taken:** Bot banned\\n\\n**Note:** Only the server owner or whitelisted users can add bots.",
									m.User.ID,
									adderID),
								Color:     0xFF0000,
								Timestamp: time.Now().Format(time.RFC3339),
							})
						}
					}
				}()

				return
			}
		}

		// Human user joined - ALWAYS give fresh start, never re-ban
		logging.Info("[USER JOINED] User %s joined guild %s - Giving fresh start", m.User.ID, m.GuildID)

		// Remove from banned list if they were previously banned (panic mode rejoin)
		if db := database.GetDB(); db != nil {
			if db.IsBannedUser(m.GuildID, m.User.ID) {
				logging.Info("[PANIC MODE REJOIN] User %s was previously banned, removing record and resetting tracking", m.User.ID)
				db.RemoveBannedUser(m.GuildID, m.User.ID)
			}
		}

		// Clear actor state for fresh start - they can be banned again if they violate
		state.ClearActorState(userID)
		logging.Info("[✓ FRESH START] User %s given clean slate in guild %s - tracking reset", m.User.ID, m.GuildID)
	}) // CRITICAL: GuildAuditLogEntryCreate - This captures WHO did the action
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
