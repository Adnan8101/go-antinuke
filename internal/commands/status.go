package commands

import (
	"fmt"
	"strings"
	"time"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

func handleStatus(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check permissions
	allowed, err := checkPermissions(s, i)
	if err != nil {
		return err
	}
	if !allowed {
		respondPermissionError(s, i, "You need Administrator permission and a role higher than the bot.")
		return nil
	}

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

	// Security level based on enabled events and panic mode
	securityLevel := "Disabled"
	if guildConfig.EnabledEvents != "" {
		if guildConfig.PanicMode {
			securityLevel = "**MAXIMUM** (Panic Mode Active)"
		} else {
			securityLevel = "**Enabled**"
		}
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
		Title:       "System Status Overview",
		Description: "Real-time security configuration and operational status.",
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Security Level",
				Value:  securityLevel,
				Inline: false,
			},
			{
				Name:   "Audit Logging",
				Value:  logChannelText,
				Inline: false,
			},
			{
				Name:   "Active Protection Modules",
				Value:  enabledText,
				Inline: true,
			},
			{
				Name:   "Disabled Modules",
				Value:  disabledText,
				Inline: true,
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
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Refresh Status",
							Style:    discordgo.PrimaryButton,
							CustomID: "status_refresh",
						},
					},
				},
			},
		},
	})
}
