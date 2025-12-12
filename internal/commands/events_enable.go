package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

// handleEventsEnable handles the /antinuke enable command
func handleEventsEnable(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check if user is owner only
	isOwner, err := checkOwnerOnly(s, i)
	if err != nil {
		return err
	}

	if !isOwner {
		respondPermissionError(s, i, "Only the server owner can enable/disable events")
		return nil
	}

	guildID := i.GuildID

	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Ensure guild config exists
	if err := db.EnsureGuildConfigExists(guildID); err != nil {
		return fmt.Errorf("failed to ensure guild configuration: %w", err)
	}

	// Get current guild config to see which events are enabled
	_, err = db.GetGuildConfig(guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch guild configuration: %w", err)
	}

	// Create checkboxes for each event
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Add All",
					Style:    discordgo.PrimaryButton,
					CustomID: "events_enable_add_all",
				},
			},
		},
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Security Module Configuration",
		Description: "Enable all protection modules to secure your server.",
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Instructions",
				Value:  "Click **Add All** below to instantly enable maximum protection.",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})

	return err
}

// handleEventsEnableAddAll enables all events
func handleEventsEnableAddAll(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID

	// Initial response to acknowledge interaction immediately
	initialEmbed := &discordgo.MessageEmbed{
		Title:       "Initializing Security Protocols",
		Description: "Setting up all protection modules...",
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems",
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{initialEmbed},
			Components: []discordgo.MessageComponent{},
		},
	})
	if err != nil {
		return err
	}

	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Get all event types
	eventTypes, err := db.GetEventTypes()
	if err != nil {
		return fmt.Errorf("failed to fetch event types: %w", err)
	}

	// Create enabled events string with all IDs
	enabledIDs := make([]string, len(eventTypes))
	for i, et := range eventTypes {
		enabledIDs[i] = strconv.Itoa(et.ID)
	}

	guildConfig, err := db.GetGuildConfig(guildID)
	if err != nil {
		return err
	}

	guildConfig.EnabledEvents = strings.Join(enabledIDs, ",")
	if err := db.UpsertGuildConfig(guildConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Sync to in-memory store for real-time detection
	if err := db.SyncGuildStateFromDB(guildID); err != nil {
		return fmt.Errorf("failed to sync state: %w", err)
	}

	// Build dynamic progress update
	var eventList []string
	for _, et := range eventTypes {
		eventList = append(eventList, fmt.Sprintf("✅ Enabled **%s**", et.Description))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Maximum Protection Enabled",
		Description: "All security modules have been successfully activated.",
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Active Modules",
				Value:  strings.Join(eventList, "\n"),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Update the message with final state
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}
