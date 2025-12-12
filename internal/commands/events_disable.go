package commands

import (
	"fmt"
	"time"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

// handleEventsDisable handles the /antinuke disable command
func handleEventsDisable(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

	// Get current guild config
	_, err = db.GetGuildConfig(guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch guild configuration: %w", err)
	}

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Disable All",
					Style:    discordgo.DangerButton,
					CustomID: "events_disable_all",
				},
			},
		},
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Disable Security Modules",
		Description: "Deactivate all protection modules.",
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Instructions",
				Value:  "Click **Disable All** below to turn off all protection systems.",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleEventsDisableAll disables all events
func handleEventsDisableAll(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID

	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	guildConfig, err := db.GetGuildConfig(guildID)
	if err != nil {
		return err
	}

	guildConfig.EnabledEvents = ""
	if err := db.UpsertGuildConfig(guildConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Sync to in-memory store for real-time detection
	if err := db.SyncGuildStateFromDB(guildID); err != nil {
		return fmt.Errorf("failed to sync state: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Protection Disabled",
		Description: "All security modules have been deactivated.",
		Color:       0x2B2D31,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{},
		},
	})
}
