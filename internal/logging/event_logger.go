package logging

import (
	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/internal/notifier"
)

// LogEventToDiscordAndDB logs an event to both Discord and the database
func LogEventToDiscordAndDB(guildID string, eventType int, eventName string, actorID string, detectionSpeedUS int64, actionTaken string) {
	// Get database and log channel
	db := database.GetDB()
	if db == nil {
		return
	}

	config, err := db.GetGuildConfig(guildID)
	if err != nil {
		return
	}

	// Get event type info for emoji
	eventTypes, err := db.GetEventTypes()
	if err != nil {
		return
	}

	var emoji string
	var description string
	for _, et := range eventTypes {
		if et.ID == eventType {
			emoji = et.Emoji
			description = et.Description
			break
		}
	}
	if description == "" {
		description = eventName
	}

	// Send to Discord if log channel is configured
	if config.LogChannelID != "" {
		notifier.SendEventLog(config.LogChannelID, emoji, description, actorID, actionTaken, detectionSpeedUS)
	}

	// Log to database
	logEntry := &database.EventLog{
		GuildID:          guildID,
		EventType:        eventType,
		ActorID:          actorID,
		DetectionSpeedUS: detectionSpeedUS,
		ActionTaken:      actionTaken,
	}
	db.LogEvent(logEntry)
}
