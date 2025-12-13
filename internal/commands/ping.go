package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// handlePing shows the actual latency to Discord API with FastHTTP optimization
func handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Measure time before responding
	startTime := time.Now()

	// Send initial response
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return err
	}

	// Measure Discord API latency using FastHTTP-optimized request
	apiStart := time.Now()
	_, err = s.Channel(i.ChannelID)
	apiLatency := time.Since(apiStart)

	// Calculate response latency
	responseLatency := time.Since(startTime)

	// Get WebSocket heartbeat latency
	wsLatency := s.HeartbeatLatency()

	// Determine quality indicator
	avgLatency := (wsLatency.Milliseconds() + apiLatency.Milliseconds()) / 2
	var statusColor int

	switch {
	case avgLatency < 30:
		statusColor = 0x00FF00 // Green
	case avgLatency < 60:
		statusColor = 0xFFFF00 // Yellow
	case avgLatency < 120:
		statusColor = 0xFFA500 // Orange
	default:
		statusColor = 0xFF0000 // Red
	}

	// Create minimal embed with only essential metrics
	embed := &discordgo.MessageEmbed{
		Title: "🚀 Pong!",
		Color: statusColor,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "⚡ WebSocket",
				Value:  fmt.Sprintf("`%dms`", wsLatency.Milliseconds()),
				Inline: true,
			},
			{
				Name:   "📡 API",
				Value:  fmt.Sprintf("`%dms`", apiLatency.Milliseconds()),
				Inline: true,
			},
			{
				Name:   "🔄 Response",
				Value:  fmt.Sprintf("`%dms`", responseLatency.Milliseconds()),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "FastHTTP Optimized",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})

	return err
}
