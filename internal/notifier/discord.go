package notifier

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var discordSession *discordgo.Session

// SetSession sets the Discord session for the notifier
func SetSession(session *discordgo.Session) {
	discordSession = session
}

// SendEventLogWithBanTime sends an event log to a Discord channel with detection and ban timing
func SendEventLogWithBanTime(channelID, emoji, eventName, actorID, actionTaken string, detectionSpeedUS, banSpeedUS int64) {
	if discordSession == nil || channelID == "" {
		return
	}

	banSpeedMS := banSpeedUS / 1000
	speedLabel := "🔨 Ban Execution"
	speedValue := fmt.Sprintf("**%d ms** (API response time)", banSpeedMS)

	if banSpeedUS < 100000 {
		speedValue = fmt.Sprintf("**%d µs**", banSpeedUS)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s Detected", emoji, eventName),
		Color:       0xED4245,
		Description: fmt.Sprintf("**Action Taken:** %s", actionTaken),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "👤 Actor",
				Value:  fmt.Sprintf("<@%s> (`%s`)", actorID, actorID),
				Inline: true,
			},
			{
				Name:   "⚡ Detection Speed",
				Value:  fmt.Sprintf("**%d µs**", detectionSpeedUS),
				Inline: true,
			},
			{
				Name:   speedLabel,
				Value:  speedValue,
				Inline: true,
			},
			{
				Name:   "🕐 Timestamp",
				Value:  fmt.Sprintf("<t:%d:F>", time.Now().Unix()),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Ultra-Low-Latency Anti-Nuke System",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	go discordSession.ChannelMessageSendEmbed(channelID, embed)
}

func SendEventLog(channelID, emoji, eventName, actorID, actionTaken string, detectionSpeedUS int64) {
	SendEventLogWithBanTime(channelID, emoji, eventName, actorID, actionTaken, detectionSpeedUS, 0)
}
