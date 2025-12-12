package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

// handleSetLimit handles /set limit
func handleSetLimit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check permissions
	allowed, err := checkPermissions(s, i)
	if err != nil {
		return err
	}
	if !allowed {
		respondPermissionError(s, i, "You need Administrator permission and a role higher than the bot.")
		return nil
	}

	options := i.ApplicationCommandData().Options

	var action string
	var limit int64
	var timeWindow int64 = 10 // Default 10s

	for _, opt := range options {
		switch opt.Name {
		case "action":
			action = opt.StringValue()
		case "limit":
			limit = opt.IntValue()
		case "time":
			timeWindow = opt.IntValue()
		}
	}

	// Parse action to event ID
	eventID, err := strconv.Atoi(action)
	if err != nil {
		return fmt.Errorf("invalid action ID")
	}

	db := database.GetDB()

	// Get event details for display
	eventTypes, err := db.GetEventTypes()
	if err != nil {
		return err
	}

	var eventName string
	for _, et := range eventTypes {
		if et.ID == eventID {
			eventName = et.Description
			break
		}
	}

	if eventName == "" {
		return fmt.Errorf("unknown event type")
	}

	// Get existing limit to preserve punishment setting
	existingLimit, err := db.GetEventLimit(i.GuildID, eventID)
	punishment := "ban" // Default
	if err == nil && existingLimit != nil {
		punishment = existingLimit.Punishment
	}

	// Save new limit
	newLimit := &database.EventLimit{
		GuildID:    i.GuildID,
		EventType:  eventID,
		MaxActions: int(limit),
		TimeWindow: int(timeWindow),
		Punishment: punishment,
	}

	if err := db.UpsertEventLimit(newLimit); err != nil {
		return fmt.Errorf("failed to save limit: %w", err)
	}

	// Sync limits to correlator for real-time enforcement
	if err := db.SyncThresholdsToMemory(i.GuildID); err != nil {
		return fmt.Errorf("failed to sync thresholds: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Configuration Updated",
		Description: fmt.Sprintf("The security parameters for **%s** have been successfully modified.", eventName),
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Action Limit",
				Value:  fmt.Sprintf("`%d` actions", limit),
				Inline: true,
			},
			{
				Name:   "Time Window",
				Value:  fmt.Sprintf("`%d` seconds", timeWindow),
				Inline: true,
			},
			{
				Name:   "Punishment Type",
				Value:  fmt.Sprintf("`%s`", strings.Title(punishment)),
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
							Label:    "Dismiss",
							Style:    discordgo.SecondaryButton,
							CustomID: "dismiss_message",
						},
					},
				},
			},
		},
	})
}

// handleSetPunishment handles /setpunishment
func handleSetPunishment(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check permissions
	allowed, err := checkPermissions(s, i)
	if err != nil {
		return err
	}
	if !allowed {
		respondPermissionError(s, i, "You need Administrator permission and a role higher than the bot.")
		return nil
	}

	options := i.ApplicationCommandData().Options

	var action string
	var punishment string

	for _, opt := range options {
		switch opt.Name {
		case "action":
			action = opt.StringValue()
		case "punishment":
			punishment = opt.StringValue()
		}
	}

	// Validate punishment
	validPunishments := map[string]bool{
		"ban":     true,
		"kick":    true,
		"timeout": true,
	}

	if !validPunishments[punishment] {
		return fmt.Errorf("invalid punishment type. Must be ban, kick, or timeout")
	}

	// Parse action to event ID
	eventID, err := strconv.Atoi(action)
	if err != nil {
		return fmt.Errorf("invalid action ID")
	}

	db := database.GetDB()

	// Get event details
	eventTypes, err := db.GetEventTypes()
	if err != nil {
		return err
	}

	var eventName string
	for _, et := range eventTypes {
		if et.ID == eventID {
			eventName = et.Description
			break
		}
	}

	// Get existing limit to preserve limit settings
	existingLimit, err := db.GetEventLimit(i.GuildID, eventID)
	maxActions := 3
	timeWindow := 10

	if err == nil && existingLimit != nil {
		maxActions = existingLimit.MaxActions
		timeWindow = existingLimit.TimeWindow
	}

	// Save new punishment
	newLimit := &database.EventLimit{
		GuildID:    i.GuildID,
		EventType:  eventID,
		MaxActions: maxActions,
		TimeWindow: timeWindow,
		Punishment: punishment,
	}

	if err := db.UpsertEventLimit(newLimit); err != nil {
		return fmt.Errorf("failed to save punishment: %w", err)
	}

	// Sync thresholds to memory for real-time enforcement
	if err := db.SyncThresholdsToMemory(i.GuildID); err != nil {
		return fmt.Errorf("failed to sync thresholds: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Punishment Updated",
		Description: fmt.Sprintf("The enforcement policy for **%s** has been updated.", eventName),
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "New Punishment",
				Value:  fmt.Sprintf("`%s`", strings.Title(punishment)),
				Inline: true,
			},
			{
				Name:   "Current Limit",
				Value:  fmt.Sprintf("`%d` in `%ds`", maxActions, timeWindow),
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
							Label:    "Dismiss",
							Style:    discordgo.SecondaryButton,
							CustomID: "dismiss_message",
						},
					},
				},
			},
		},
	})
}
