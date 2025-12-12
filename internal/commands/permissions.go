package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// checkPermissions checks if the user has permission to run the command
// Returns true if:
// 1. User is the server owner, OR
// 2. User has Administrator permission AND their highest role is higher than the bot's
func checkPermissions(s *discordgo.Session, i *discordgo.InteractionCreate) (bool, error) {
	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		guild, err = s.Guild(i.GuildID)
		if err != nil {
			return false, fmt.Errorf("failed to get guild: %w", err)
		}
	}

	// Check if user is owner
	if i.Member.User.ID == guild.OwnerID {
		return true, nil
	}

	// Check if user has admin permission
	permissions, err := s.State.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		permissions, err = s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
		if err != nil {
			return false, fmt.Errorf("failed to get permissions: %w", err)
		}
	}

	hasAdmin := permissions&discordgo.PermissionAdministrator != 0
	if !hasAdmin {
		return false, nil
	}

	// Get bot's highest role
	botMember, err := s.State.Member(i.GuildID, s.State.User.ID)
	if err != nil {
		botMember, err = s.GuildMember(i.GuildID, s.State.User.ID)
		if err != nil {
			return false, fmt.Errorf("failed to get bot member: %w", err)
		}
	}

	botHighestRole := getHighestRole(guild, botMember.Roles)
	userHighestRole := getHighestRole(guild, i.Member.Roles)

	// User must have a higher role than bot
	if userHighestRole != nil && botHighestRole != nil {
		return userHighestRole.Position > botHighestRole.Position, nil
	}

	return hasAdmin, nil
}

// checkOwnerOnly checks if the user is the server owner
func checkOwnerOnly(s *discordgo.Session, i *discordgo.InteractionCreate) (bool, error) {
	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		guild, err = s.Guild(i.GuildID)
		if err != nil {
			return false, fmt.Errorf("failed to get guild: %w", err)
		}
	}

	return i.Member.User.ID == guild.OwnerID, nil
}

// getHighestRole returns the highest role from a list of role IDs
func getHighestRole(guild *discordgo.Guild, roleIDs []string) *discordgo.Role {
	var highest *discordgo.Role
	for _, roleID := range roleIDs {
		for _, role := range guild.Roles {
			if role.ID == roleID {
				if highest == nil || role.Position > highest.Position {
					highest = role
				}
			}
		}
	}
	return highest
}

// respondPermissionError sends a permission denied error response
func respondPermissionError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := &discordgo.MessageEmbed{
		Title:       "Access Denied",
		Description: message,
		Color:       0x2B2D31,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Anti-Nuke Security Systems • Enterprise Grade Protection",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
