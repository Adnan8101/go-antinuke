package commands

import (
	"fmt"
	"time"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

// handleLogsEnable handles the /logs enable command
func handleLogsEnable(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check permissions
	allowed, err := checkPermissions(s, i)
	if err != nil {
		return err
	}
	if !allowed {
		respondPermissionError(s, i, "You need Administrator permission and a role higher than the bot.")
		return nil
	}

	data := i.ApplicationCommandData()
	guildID := i.GuildID

	// Get the channel option from the subcommand
	options := data.Options[0].Options // "enable" subcommand options
	channelID := options[0].ChannelValue(s).ID

	// Save to database
	db := database.GetDB()
	config, err := db.GetGuildConfig(guildID)
	if err != nil {
		return err
	}

	config.LogChannelID = channelID
	if err := db.UpsertGuildConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Send test message to the log channel
	testEmbed := &discordgo.MessageEmbed{
		Title:       "Logging System Active",
		Description: "This channel has been successfully configured as the primary audit log destination.",
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Log Format",
				Value:  "• Event Classification\n• Actor Identification\n• Detection Latency (µs)\n• Automated Response",
				Inline: false,
			},
			{
				Name:   "Performance Metrics",
				Value:  "Target Detection Latency: <1µs",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = s.ChannelMessageSendEmbed(channelID, testEmbed)
	if err != nil {
		return fmt.Errorf("failed to send test message (check bot permissions): %w", err)
	}

	// Send confirmation to user
	confirmEmbed := &discordgo.MessageEmbed{
		Title:       "Log Channel Configured",
		Description: fmt.Sprintf("Security audit logs will now be streamed to <#%s>.", channelID),
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Destination",
				Value:  fmt.Sprintf("<#%s>", channelID),
				Inline: false,
			},
			{
				Name:   "Status",
				Value:  "Verification message sent successfully.",
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
			Embeds: []*discordgo.MessageEmbed{confirmEmbed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
