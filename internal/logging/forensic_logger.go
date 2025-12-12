package logging

type ForensicLogEntry struct {
	Timestamp int64                  `json:"timestamp"`
	GuildID   uint64                 `json:"guild_id"`
	EventType string                 `json:"event_type"`
	ActorID   uint64                 `json:"actor_id"`
	TargetID  uint64                 `json:"target_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}
