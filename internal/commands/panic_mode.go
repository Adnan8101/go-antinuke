package commands

import (
	"fmt"

	cfg "go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/pkg/util"

	"github.com/bwmarrin/discordgo"
)

// handlePanicMode handles the /panic mode command
func handlePanicMode(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	guildID := i.GuildID

	// Get the state option from the subcommand
	options := data.Options[0].Options // "mode" subcommand options
	state := options[0].StringValue()  // "enable" or "disable"

	panicMode := state == "enable"

	// Save to database
	db := database.GetDB()
	config, err := db.GetGuildConfig(guildID)
	if err != nil {
		return err
	}

	config.PanicMode = panicMode
	if err := db.UpsertGuildConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Sync with in-memory store for zero-latency checks
	if id, err := util.StringToUint64(guildID); err == nil {
		store := cfg.GetProfileStore()
		store.SetPanicMode(id, panicMode)
	}

	// Create response embed based on mode
	var embed *discordgo.MessageEmbed
	if panicMode {
		embed = &discordgo.MessageEmbed{
			Title:       "⚠️ PANIC MODE ENABLED",
			Description: "**Ultra-Aggressive Protection Activated**",
			Color:       0xED4245, // Red
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "🚨 Enforcement Policy",
					Value:  "**INSTANT BAN** on ANY detection (even 1 count)\nNo warnings, no thresholds, immediate action",
					Inline: false,
				},
				{
					Name:   "⚡ Detection Speed",
					Value:  "Target: <1µs detection → <200ms execution",
					Inline: false,
				},
				{
					Name:   "⚠️ Warning",
					Value:  "This mode is extremely aggressive. Use with caution.",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Panic Mode Active",
			},
		}
	} else {
		embed = &discordgo.MessageEmbed{
			Title:       "✅ Normal Mode Enabled",
			Description: "Standard protection with configured thresholds",
			Color:       0x57F287, // Green
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "📊 Enforcement Policy",
					Value:  "Following configured rate limits and thresholds\nWarnings and gradual escalation enabled",
					Inline: false,
				},
				{
					Name:   "⚙️ Configuration",
					Value:  "Use `/antinuke enable` to configure event protection",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Normal Mode Active",
			},
		}
	}

	// Send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
