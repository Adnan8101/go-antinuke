package bot

import (
	"fmt"
	"strconv"

	"go-antinuke-2.0/internal/logging"
	"go-antinuke-2.0/internal/state"

	"github.com/bwmarrin/discordgo"
)

type Session struct {
	discord *discordgo.Session
	token   string
	BotID   uint64
}

var globalSession *Session

// Initialize creates and initializes the Discord session
func Initialize(token string) error {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Set required intents - enable ALL intents for comprehensive event detection
	dg.Identify.Intents = discordgo.IntentsAll

	globalSession = &Session{
		discord: dg,
		token:   token,
	}

	return nil
}

// GetSession returns the global Discord session
func GetSession() *Session {
	return globalSession
}

// GetDiscord returns the underlying discordgo session
func (s *Session) GetDiscord() *discordgo.Session {
	return s.discord
}

// Connect opens the Discord websocket connection
func (s *Session) Connect() error {
	if err := s.discord.Open(); err != nil {
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}

	// Store bot ID
	if s.discord.State.User != nil {
		botID, _ := strconv.ParseUint(s.discord.State.User.ID, 10, 64)
		s.BotID = botID
		state.SetBotID(botID)
		logging.Info("Bot ID: %d", botID)
	}

	logging.Info("Discord bot connected successfully")
	return nil
}

// Close closes the Discord connection
func (s *Session) Close() error {
	if s.discord != nil {
		return s.discord.Close()
	}
	return nil
}

// RegisterCommands registers all slash commands with Discord
func (s *Session) RegisterCommands(commands []*discordgo.ApplicationCommand) error {
	logging.Info("Registering %d slash commands...", len(commands))

	for _, cmd := range commands {
		_, err := s.discord.ApplicationCommandCreate(s.discord.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
		logging.Info("Registered command: /%s", cmd.Name)
	}

	return nil
}

// AddHandler adds an event handler to the Discord session
func (s *Session) AddHandler(handler interface{}) {
	s.discord.AddHandler(handler)
}

// SyncGuildsFromDatabase syncs all guild configurations from database to in-memory store
func (s *Session) SyncGuildsFromDatabase(db interface {
	SyncAllGuildsFromDB() error
	EnsureGuildConfigExists(guildID string) error
}) error {
	logging.Info("Syncing guild configurations from database...")

	// Sync all existing guilds
	if err := db.SyncAllGuildsFromDB(); err != nil {
		logging.Warn("Failed to sync all guilds: %v", err)
	}

	// Ensure all current guilds in the bot have configs
	for _, guild := range s.discord.State.Guilds {
		if err := db.EnsureGuildConfigExists(guild.ID); err != nil {
			logging.Warn("Failed to ensure config for guild %s: %v", guild.ID, err)
		}
	}

	logging.Info("Guild sync completed")
	return nil
}
