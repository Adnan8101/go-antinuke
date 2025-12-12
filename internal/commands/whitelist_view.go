package commands

import (
	"fmt"
	"strings"
	"time"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

// handleWhitelistView handles /antinuke whitelist view command
func handleWhitelistView(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

	// Get all whitelist entries for this guild
	whitelists, err := db.GetAllWhitelist(guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch whitelist: %w", err)
	}

	if len(whitelists) == 0 {
		embed := &discordgo.MessageEmbed{
			Title:       "Whitelist Registry",
			Description: "No users or roles are currently whitelisted.",
			Color:       0x2B2D31,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get event types for descriptions
	eventTypes, err := db.GetEventTypes()
	if err != nil {
		return fmt.Errorf("failed to fetch event types: %w", err)
	}

	eventTypeMap := make(map[int]string)
	for _, et := range eventTypes {
		eventTypeMap[et.ID] = et.Description
	}

	// Group by target
	targetMap := make(map[string][]string)
	targetTypeMap := make(map[string]string)

	for _, w := range whitelists {
		eventDesc := "All Events"
		if w.EventType != 0 {
			if desc, ok := eventTypeMap[w.EventType]; ok {
				eventDesc = desc
			} else {
				eventDesc = fmt.Sprintf("Event ID %d", w.EventType)
			}
		}

		targetMap[w.TargetID] = append(targetMap[w.TargetID], eventDesc)
		targetTypeMap[w.TargetID] = w.TargetType
	}

	// Build the whitelist display
	var userList []string
	var roleList []string

	for targetID, events := range targetMap {
		targetType := targetTypeMap[targetID]
		mention := fmt.Sprintf("<@%s>", targetID)
		if targetType == "role" {
			mention = fmt.Sprintf("<@&%s>", targetID)
		}

		eventsText := strings.Join(events, ", ")
		entry := fmt.Sprintf("%s\n└ **Scope:** %s", mention, eventsText)

		if targetType == "user" {
			userList = append(userList, entry)
		} else {
			roleList = append(roleList, entry)
		}
	}

	fields := []*discordgo.MessageEmbedField{}

	if len(userList) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Whitelisted Users",
			Value:  strings.Join(userList, "\n\n"),
			Inline: false,
		})
	}

	if len(roleList) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Whitelisted Roles",
			Value:  strings.Join(roleList, "\n\n"),
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Whitelist Registry",
		Description: fmt.Sprintf("**%d** entries configured with security exceptions.", len(whitelists)),
		Color:       0x2B2D31,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
