package commands

import (
	"fmt"
	"time"

	cfg "go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/pkg/util"

	"github.com/bwmarrin/discordgo"
)

// handlePanicMode handles the /panic command
func handlePanicMode(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check permissions - Owner only for panic mode
	isOwner, err := checkOwnerOnly(s, i)
	if err != nil {
		return err
	}
	if !isOwner {
		respondPermissionError(s, i, "Only the server owner can toggle panic mode")
		return nil
	}

	data := i.ApplicationCommandData()
	guildID := i.GuildID

	// Get the enable option directly
	panicMode := data.Options[0].BoolValue()

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
			Title:       "⚠️ PANIC MODE ACTIVATED",
			Description: "**Maximum Security Protocol Engaged**",
			Color:       0xED4245,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Enforcement Policy",
					Value:  "**ZERO TOLERANCE**\nImmediate ban on any detected anomaly.\nAll thresholds bypassed.",
					Inline: false,
				},
				{
					Name:   "System Latency",
					Value:  "Optimized for <1µs detection",
					Inline: false,
				},
				{
					Name:   "Warning",
					Value:  "This mode is extremely aggressive. Legitimate actions may be flagged.",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
	} else {
		embed = &discordgo.MessageEmbed{
			Title:       "Standard Protection Active",
			Description: "System returned to standard operational parameters.",
			Color:       0x2B2D31,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Enforcement Policy",
					Value:  "Standard rate limits and thresholds applied.\nGradual escalation enabled.",
					Inline: false,
				},
				{
					Name:   "Configuration",
					Value:  "Use `/antinuke enable` to manage specific modules.",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
			},
			Timestamp: time.Now().Format(time.RFC3339),
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
