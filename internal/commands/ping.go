package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// handlePing shows the actual latency to Discord API
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

	// Measure Discord API latency
	apiStart := time.Now()
	_, err = s.Channel(i.ChannelID)
	apiLatency := time.Since(apiStart)

	// Calculate response latency
	responseLatency := time.Since(startTime)

	// Get WebSocket heartbeat latency
	wsLatency := s.HeartbeatLatency()

	// Create embed with ping information
	embed := &discordgo.MessageEmbed{
		Title:       " Pong!",
		Color:       0x00FF00,
		Description: "Discord API Latency Measurements",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "⚡ WebSocket Heartbeat",
				Value:  fmt.Sprintf("`%dms` (%dµs)", wsLatency.Milliseconds(), wsLatency.Microseconds()),
				Inline: true,
			},
			{
				Name:   "📡 API Round-Trip",
				Value:  fmt.Sprintf("`%dms` (%dµs)", apiLatency.Milliseconds(), apiLatency.Microseconds()),
				Inline: true,
			},
			{
				Name:   "🔄 Response Time",
				Value:  fmt.Sprintf("`%dms` (%dµs)", responseLatency.Milliseconds(), responseLatency.Microseconds()),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Ultra-Low-Latency Antinuke Engine",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Determine quality indicator
	avgLatency := (wsLatency.Milliseconds() + apiLatency.Milliseconds()) / 2
	var quality string
	var statusColor int

	switch {
	case avgLatency < 50:
		quality = "🟢 Excellent"
		statusColor = 0x00FF00
	case avgLatency < 100:
		quality = "🟡 Good"
		statusColor = 0xFFFF00
	case avgLatency < 200:
		quality = "🟠 Fair"
		statusColor = 0xFFA500
	default:
		quality = "🔴 Poor"
		statusColor = 0xFF0000
	}

	embed.Color = statusColor
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📊 Connection Quality",
		Value:  quality,
		Inline: false,
	})

	// Add nanosecond precision for ultra-low-latency monitoring
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:  "⚙️ Precision Metrics",
		Value: fmt.Sprintf("API: `%dns`\nWS: `%dns`", apiLatency.Nanoseconds(), wsLatency.Nanoseconds()),
		Inline: false,
	})

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})

	return err
}
