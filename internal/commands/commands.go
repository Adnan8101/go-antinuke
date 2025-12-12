package commands

import "github.com/bwmarrin/discordgo"

// GetAllCommands returns all application commands
func GetAllCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "antinuke",
			Description: "Manage anti-nuke system",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "enable",
					Description: "Enable anti-nuke protection events",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "disable",
					Description: "Disable anti-nuke protection events",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "whitelist",
					Description: "Manage whitelist",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "add",
							Description: "Add user/role to whitelist",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Name:        "user",
									Description: "User to whitelist",
									Type:        discordgo.ApplicationCommandOptionUser,
									Required:    false,
								},
								{
									Name:        "role",
									Description: "Role to whitelist",
									Type:        discordgo.ApplicationCommandOptionRole,
									Required:    false,
								},
							},
						},
						{
							Name:        "remove",
							Description: "Remove user/role from whitelist",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Name:        "user",
									Description: "User to remove from whitelist",
									Type:        discordgo.ApplicationCommandOptionUser,
									Required:    false,
								},
								{
									Name:        "role",
									Description: "Role to remove from whitelist",
									Type:        discordgo.ApplicationCommandOptionRole,
									Required:    false,
								},
							},
						},
						{
							Name:        "view",
							Description: "View all whitelisted users and roles",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
				},
			},
		},
		{
			Name:        "set",
			Description: "Configure settings",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "limit",
					Description: "Set rate limits for an event",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "action",
							Description: "The event to configure",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							// We can't dynamically populate choices here easily without registering commands dynamically
							// For now, user has to input ID or we use autocomplete (advanced)
							// Let's assume user inputs ID for now or we provide a list in description
						},
						{
							Name:        "limit",
							Description: "Max actions allowed",
							Type:        discordgo.ApplicationCommandOptionInteger,
							Required:    true,
						},
						{
							Name:        "time",
							Description: "Time window in seconds (default 10s)",
							Type:        discordgo.ApplicationCommandOptionInteger,
							Required:    false,
						},
					},
				},
			},
		},
		{
			Name:        "setpunishment",
			Description: "Set punishment for an event",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "action",
					Description: "The event to configure",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "punishment",
					Description: "Punishment type",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Ban",
							Value: "ban",
						},
						{
							Name:  "Kick",
							Value: "kick",
						},
						{
							Name:  "Timeout",
							Value: "timeout",
						},
					},
				},
			},
		},
		{
			Name:        "panic",
			Description: "Toggle panic mode (lockdown)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "enable",
					Description: "Enable/Disable panic mode",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    true,
				},
			},
		},
		{
			Name:        "logs",
			Description: "Configure logging",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "enable",
					Description: "Set log channel",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "channel",
							Description: "Channel to send logs to",
							Type:        discordgo.ApplicationCommandOptionChannel,
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "status",
			Description: "Show system status",
		},
		{
			Name:        "ping",
			Description: "Check Discord API latency and connection quality",
		},
		{
			Name:        "stats",
			Description: "Show comprehensive VM and system statistics",
		},
	}
}
