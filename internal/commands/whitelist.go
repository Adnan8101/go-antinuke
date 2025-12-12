package commands

import (
	"fmt"
	"strings"
	"time"

	"go-antinuke-2.0/internal/database"

	"github.com/bwmarrin/discordgo"
)

// handleWhitelistAdd handles /antinuke whitelist add
func handleWhitelistAdd(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check permissions
	allowed, err := checkPermissions(s, i)
	if err != nil {
		return err
	}
	if !allowed {
		respondPermissionError(s, i, "You need Administrator permission and a role higher than the bot.")
		return nil
	}

	options := i.ApplicationCommandData().Options[0].Options // whitelist subcommand group

	// Find the "add" subcommand
	var addOptions []*discordgo.ApplicationCommandInteractionDataOption
	for _, opt := range options {
		if opt.Name == "add" {
			addOptions = opt.Options
			break
		}
	}

	if addOptions == nil {
		return fmt.Errorf("add subcommand not found")
	}

	var targetID string
	var targetType string
	var targetName string

	// Get user or role
	for _, opt := range addOptions {
		if opt.Name == "user" {
			targetID = opt.UserValue(s).ID
			targetType = "user"
			targetName = opt.UserValue(s).Username
		} else if opt.Name == "role" {
			targetID = opt.RoleValue(s, i.GuildID).ID
			targetType = "role"
			targetName = opt.RoleValue(s, i.GuildID).Name
		}
	}

	if targetID == "" {
		return fmt.Errorf("no user or role specified")
	}

	// Shorten custom IDs to avoid 100 char limit
	// whitelist_add_all_user_123456789012345678 -> wl_add_all_u_123...
	// whitelist_select_user_123456789012345678 -> wl_sel_u_123...

	shortType := "u"
	if targetType == "role" {
		shortType = "r"
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Whitelist for All Events",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("wl_add_all_%s_%s", shortType, targetID),
				},
			},
		},
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Whitelist Configuration",
		Description: fmt.Sprintf("Configure security exceptions for **%s** (%s).", targetName, targetType),
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Instructions",
				Value:  "Click **Whitelist for All Events** below to grant full immunity.",
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
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleWhitelistRemove handles /antinuke whitelist remove
func handleWhitelistRemove(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check permissions
	allowed, err := checkPermissions(s, i)
	if err != nil {
		return err
	}
	if !allowed {
		respondPermissionError(s, i, "You need Administrator permission and a role higher than the bot.")
		return nil
	}

	options := i.ApplicationCommandData().Options[0].Options // whitelist subcommand group

	// Find the "remove" subcommand
	var removeOptions []*discordgo.ApplicationCommandInteractionDataOption
	for _, opt := range options {
		if opt.Name == "remove" {
			removeOptions = opt.Options
			break
		}
	}

	if removeOptions == nil {
		return fmt.Errorf("remove subcommand not found")
	}

	var targetID string
	var targetType string
	var targetName string

	for _, opt := range removeOptions {
		if opt.Name == "user" {
			targetID = opt.UserValue(s).ID
			targetType = "user"
			targetName = opt.UserValue(s).Username
		} else if opt.Name == "role" {
			targetID = opt.RoleValue(s, i.GuildID).ID
			targetType = "role"
			targetName = opt.RoleValue(s, i.GuildID).Name
		}
	}

	if targetID == "" {
		return fmt.Errorf("no user or role specified")
	}

	db := database.GetDB()

	// Get current whitelists for this target
	whitelists, err := db.GetWhitelistByTarget(i.GuildID, targetID)
	if err != nil {
		return err
	}

	if len(whitelists) == 0 {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("**%s** is not whitelisted for any events.", targetName),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Remove All Whitelists",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("whitelist_remove_all_%s_%s", targetType, targetID),
				},
			},
		},
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Remove Whitelist",
		Description: fmt.Sprintf("Revoke security exceptions for **%s**.", targetName),
		Color:       0x2B2D31,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Instructions",
				Value:  "Click **Remove All Whitelists** below to revoke all immunities.",
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
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleWhitelistAddAll handles the button to whitelist for all events
func handleWhitelistAddAll(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	if len(parts) < 5 {
		return fmt.Errorf("invalid custom ID")
	}

	// wl_add_all_u_123456
	shortType := parts[3]
	targetID := parts[4]

	targetType := "user"
	if shortType == "r" {
		targetType = "role"
	}

	db := database.GetDB()
	// EventType 0 means all events
	if err := db.AddWhitelist(i.GuildID, targetID, targetType, 0); err != nil {
		return err
	}

	// Sync whitelist to in-memory store for real-time checking
	if err := db.SyncWhitelistToMemory(i.GuildID); err != nil {
		return fmt.Errorf("failed to sync whitelist: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Whitelist Updated",
		Description: "Target has been successfully whitelisted for **ALL** security events.",
		Color:       0x2B2D31,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{},
		},
	})
}

// handleWhitelistRemoveAll handles the button to remove all whitelists
func handleWhitelistRemoveAll(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	if len(parts) < 5 {
		return fmt.Errorf("invalid custom ID")
	}
	targetID := parts[4]

	db := database.GetDB()

	// We need to remove all entries for this target
	// The database method RemoveWhitelist takes an event type.
	// We might need a new method to remove all, or iterate.
	// For now, let's assume we can remove all by passing a special flag or iterating.

	// Let's get all whitelists first
	whitelists, err := db.GetWhitelistByTarget(i.GuildID, targetID)
	if err != nil {
		return err
	}

	for _, w := range whitelists {
		if err := db.RemoveWhitelist(i.GuildID, targetID, w.EventType); err != nil {
			return err
		}
	}

	// Sync whitelist to in-memory store after removal
	if err := db.SyncWhitelistToMemory(i.GuildID); err != nil {
		return fmt.Errorf("failed to sync whitelist: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Whitelist Cleared",
		Description: "All security exceptions for this target have been revoked.",
		Color:       0x2B2D31,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{},
		},
	})
}
