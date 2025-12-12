package commands

import (
	"fmt"
	"strings"

	"go-antinuke-2.0/internal/bot"
	"go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/internal/logging"
	"go-antinuke-2.0/pkg/util"

	"github.com/bwmarrin/discordgo"
)

// Handler manages all command interactions
type Handler struct {
	session *bot.Session
}

var globalHandler *Handler

// Initialize creates and initializes the command handler
func Initialize(session *bot.Session) error {
	globalHandler = &Handler{
		session: session,
	}

	// Register interaction handler
	session.AddHandler(globalHandler.handleInteraction)

	// Register all commands
	commands := GetAllCommands()
	if err := session.RegisterCommands(commands); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	logging.Info("Command handler initialized with %d commands", len(commands))
	return nil
}

// GetHandler returns the global command handler
func GetHandler() *Handler {
	return globalHandler
}

// handleInteraction routes all interactions (commands, buttons, dropdowns)
func (h *Handler) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		h.handleComponent(s, i)
	}
}

// handleCommand routes slash commands to their handlers
func (h *Handler) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	var err error
	switch data.Name {
	case "antinuke":
		err = handleAntiNukeEnable(s, i)
	case "panic":
		err = handlePanicMode(s, i)
	case "logs":
		err = handleLogsEnable(s, i)
	case "status":
		err = handleStatus(s, i)
	default:
		err = fmt.Errorf("unknown command: %s", data.Name)
	}

	if err != nil {
		logging.Error("Command error [%s]: %v", data.Name, err)
		respondError(s, i, err.Error())
	}
}

// handleComponent routes component interactions (buttons, dropdowns)
func (h *Handler) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()

	var err error
	switch {
	case data.CustomID == "antinuke_event_select":
		err = handleEventSelect(s, i)
	case data.CustomID == "antinuke_enable_all":
		err = handleEnableAll(s, i)
	default:
		err = fmt.Errorf("unknown component: %s", data.CustomID)
	}

	if err != nil {
		logging.Error("Component error [%s]: %v", data.CustomID, err)
		respondError(s, i, err.Error())
	}
}

// GetAllCommands returns all slash command definitions
func GetAllCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "antinuke",
			Description: "Configure anti-nuke event protection",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "enable",
					Description: "Enable and configure anti-nuke events",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:        "panic",
			Description: "Configure anti-nuke panic mode",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "mode",
					Description: "Set panic mode (instant ban on any detection)",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "state",
							Description: "Enable or disable panic mode",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Enable", Value: "enable"},
								{Name: "Disable", Value: "disable"},
							},
						},
					},
				},
			},
		},
		{
			Name:        "logs",
			Description: "Configure anti-nuke logging",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "enable",
					Description: "Set the channel for anti-nuke logs",
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
			Description: "View current anti-nuke system status and configuration",
		},
	}
}

// respondError sends an error response to the user
func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("❌ Error: %s", message),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// Helper function to parse enabled events from comma-separated string
func parseEnabledEvents(enabledStr string) map[int]bool {
	enabled := make(map[int]bool)
	if enabledStr == "" {
		return enabled
	}

	parts := strings.Split(enabledStr, ",")
	for _, part := range parts {
		var id int
		fmt.Sscanf(part, "%d", &id)
		enabled[id] = true
	}
	return enabled
}

// Helper function to serialize enabled events to comma-separated string
func serializeEnabledEvents(enabled map[int]bool) string {
	var ids []string
	for id := range enabled {
		ids = append(ids, fmt.Sprintf("%d", id))
	}
	return strings.Join(ids, ",")
}

// IsEventEnabled checks if an event type is enabled for a guild
func IsEventEnabled(guildID string, eventType int) bool {
	db := database.GetDB()
	if db == nil {
		return false
	}

	config, err := db.GetGuildConfig(guildID)
	if err != nil {
		return false
	}

	enabled := parseEnabledEvents(config.EnabledEvents)
	return enabled[eventType]
}

// IsPanicMode checks if panic mode is enabled for a guild
func IsPanicMode(guildID string) bool {
	// Use in-memory store for zero-latency check
	id, err := util.StringToUint64(guildID)
	if err != nil {
		return false
	}

	store := config.GetProfileStore()
	return store.IsPanicMode(id)
}

// GetLogChannel returns the log channel ID for a guild
func GetLogChannel(guildID string) string {
	db := database.GetDB()
	if db == nil {
		return ""
	}

	config, err := db.GetGuildConfig(guildID)
	if err != nil {
		return ""
	}

	return config.LogChannelID
}
