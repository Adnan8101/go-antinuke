package commands

import (
	"fmt"
	"strconv"
	"strings"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

// handleAntiNukeEnable handles the /antinuke enable command
func handleAntiNukeEnable(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get all event types from database
	db := database.GetDB()
	eventTypes, err := db.GetEventTypes()
	if err != nil {
		return fmt.Errorf("failed to get event types: %w", err)
	}

	// Create dropdown options (max 25 per Discord limit, we have 26 so split if needed)
	options := make([]discordgo.SelectMenuOption, 0)
	for _, et := range eventTypes {
		if len(options) >= 25 {
			break // Discord limit
		}
		options = append(options, discordgo.SelectMenuOption{
			Label:       et.Description,
			Value:       strconv.Itoa(et.ID),
			Description: fmt.Sprintf("Toggle %s protection", et.Description),
			Emoji: &discordgo.ComponentEmoji{
				Name: "✅",
			},
		})
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "🛡️ Anti-Nuke Event Configuration",
		Description: "Select the events you want to protect against, or enable all events at once.",
		Color:       0x5865F2, // Discord blurple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📋 Available Events",
				Value:  fmt.Sprintf("Total: %d protection modules", len(eventTypes)),
				Inline: false,
			},
			{
				Name:   "⚙️ Configuration",
				Value:  "Use the dropdown below to select events or click 'Enable All Events'",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Ultra-Low-Latency Anti-Nuke System",
		},
	}

	// Get current config to show enabled events
	guildID := i.GuildID
	config, err := db.GetGuildConfig(guildID)
	if err == nil && config.EnabledEvents != "" {
		enabled := parseEnabledEvents(config.EnabledEvents)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "✅ Currently Enabled",
			Value:  fmt.Sprintf("%d events are currently enabled", len(enabled)),
			Inline: false,
		})
	}

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "antinuke_event_select",
					Placeholder: "Select events to enable...",
					MinValues:   new(int), // 0 minimum
					MaxValues:   len(options),
					Options:     options,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Enable All Events",
					Style:    discordgo.SuccessButton,
					CustomID: "antinuke_enable_all",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🛡️",
					},
				},
			},
		},
	}

	// Send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleEventSelect handles dropdown selection for events
func handleEventSelect(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.MessageComponentData()
	guildID := i.GuildID

	// Get selected event IDs
	enabled := make(map[int]bool)
	for _, value := range data.Values {
		id, _ := strconv.Atoi(value)
		enabled[id] = true
	}

	// Save to database
	db := database.GetDB()
	config, err := db.GetGuildConfig(guildID)
	if err != nil {
		return err
	}

	config.EnabledEvents = serializeEnabledEvents(enabled)
	if err := db.UpsertGuildConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Get event names for confirmation
	eventTypes, _ := db.GetEventTypes()
	eventNames := make([]string, 0)
	for id := range enabled {
		for _, et := range eventTypes {
			if et.ID == id {
				eventNames = append(eventNames, et.Description)
				break
			}
		}
	}

	// Create confirmation embed
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Events Updated",
		Description: fmt.Sprintf("Successfully enabled %d event(s)", len(enabled)),
		Color:       0x57F287, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Enabled Events",
				Value:  strings.Join(eventNames, "\n"),
				Inline: false,
			},
		},
	}

	// Update the message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{}, // Remove components after selection
		},
	})
}

// handleEnableAll handles the "Enable All Events" button
func handleEnableAll(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID

	// Get all event types
	db := database.GetDB()
	eventTypes, err := db.GetEventTypes()
	if err != nil {
		return err
	}

	// Enable all events
	enabled := make(map[int]bool)
	for _, et := range eventTypes {
		enabled[et.ID] = true
	}

	// Save to database
	config, err := db.GetGuildConfig(guildID)
	if err != nil {
		return err
	}

	config.EnabledEvents = serializeEnabledEvents(enabled)
	if err := db.UpsertGuildConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create confirmation embed
	embed := &discordgo.MessageEmbed{
		Title:       "✅ All Events Enabled",
		Description: fmt.Sprintf("Successfully enabled all %d anti-nuke events!", len(eventTypes)),
		Color:       0x57F287, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🛡️ Protection Status",
				Value:  "**Maximum Protection Activated**\nAll events are now being monitored",
				Inline: false,
			},
			{
				Name:   "⚡ Detection Speed",
				Value:  "Target: <1µs detection latency",
				Inline: false,
			},
		},
	}

	// Update the message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{}, // Remove components after selection
		},
	})
}
