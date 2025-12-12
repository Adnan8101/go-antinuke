package commands

import (
	"fmt"
	"strings"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

func handleStatus(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID

	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Ensure guild config exists (creates default with all events enabled if needed)
	if err := db.EnsureGuildConfigExists(guildID); err != nil {
		return fmt.Errorf("failed to ensure guild configuration: %w", err)
	}

	// Fetch guild config from database (single source of truth)
	guildConfig, err := db.GetGuildConfig(guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch guild configuration: %w", err)
	}

	// Panic mode status directly from database
	panicStatus := "Disabled"
	if guildConfig.PanicMode {
		panicStatus = "**ENABLED** (Instant ban on first detection)"
	}

	// Get event types
	eventTypes, err := db.GetEventTypes()
	if err != nil {
		return fmt.Errorf("failed to fetch event types: %w", err)
	}

	// Parse enabled events from database
	enabledEvents := parseEnabledEvents(guildConfig.EnabledEvents)

	var enabledList []string
	var disabledList []string

	for _, et := range eventTypes {
		name := et.Description
		if enabledEvents[et.ID] {
			enabledList = append(enabledList, fmt.Sprintf("• %s", name))
		} else {
			disabledList = append(disabledList, fmt.Sprintf("• %s", name))
		}
	}

	enabledText := "None"
	if len(enabledList) > 0 {
		enabledText = strings.Join(enabledList, "\n")
	}

	disabledText := "None"
	if len(disabledList) > 0 {
		disabledText = strings.Join(disabledList, "\n")
	}

	logChannel := guildConfig.LogChannelID
	logChannelText := "Not configured"
	if logChannel != "" {
		logChannelText = fmt.Sprintf("<#%s>", logChannel)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Anti-Nuke System Status",
		Description: "Current configuration and protection status",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Panic Mode",
				Value:  panicStatus,
				Inline: false,
			},
			{
				Name:   "Log Channel",
				Value:  logChannelText,
				Inline: false,
			},
			{
				Name:   "Enabled Events",
				Value:  enabledText,
				Inline: true,
			},
			{
				Name:   "Disabled Events",
				Value:  disabledText,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Ultra-Low-Latency Anti-Nuke System",
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
