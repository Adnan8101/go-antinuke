package commands

import (
	"fmt"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

// handleLogsEnable handles the /logs enable command
func handleLogsEnable(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
		Title:       "✅ Anti-Nuke Logging Enabled",
		Description: "This channel will now receive all anti-nuke event logs",
		Color:       0x57F287, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Log Format",
				Value:  "Each log will include:\n• Event type and emoji\n• User information\n• Detection speed (microseconds)\n• Action taken",
				Inline: false,
			},
			{
				Name:   "⚡ Performance",
				Value:  "Target detection speed: <1µs",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Ultra-Low-Latency Anti-Nuke System",
		},
	}

	_, err = s.ChannelMessageSendEmbed(channelID, testEmbed)
	if err != nil {
		return fmt.Errorf("failed to send test message (check bot permissions): %w", err)
	}

	// Send confirmation to user
	confirmEmbed := &discordgo.MessageEmbed{
		Title:       "✅ Log Channel Configured",
		Description: fmt.Sprintf("Anti-nuke logs will be sent to <#%s>", channelID),
		Color:       0x57F287, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📌 Channel",
				Value:  fmt.Sprintf("<#%s>", channelID),
				Inline: false,
			},
			{
				Name:   "✅ Status",
				Value:  "Test message sent successfully",
				Inline: false,
			},
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{confirmEmbed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
