package ingest

import (
	"encoding/json"
	"strconv"

	"go-antinuke-2.0/pkg/util"
)

func SliceEvent(eventType string, rawData json.RawMessage) *Event {
	data := []byte(rawData)

	switch eventType {
	case "GUILD_BAN_ADD":
		return sliceBanAdd(data)
	case "GUILD_MEMBER_REMOVE":
		return sliceMemberRemove(data)
	case "CHANNEL_DELETE":
		return sliceChannelDelete(data)
	case "GUILD_ROLE_DELETE":
		return sliceRoleDelete(data)
	case "WEBHOOKS_UPDATE":
		return sliceWebhookUpdate(data)
	case "GUILD_ROLE_UPDATE":
		return sliceRoleUpdate(data)
	default:
		return nil
	}
}

func sliceBanAdd(data []byte) *Event {
	guildID := extractU64Field(data, "guild_id")
	userID := extractU64Field(data, "user", "id")

	if guildID == 0 || userID == 0 {
		return nil
	}

	return &Event{
		EventType: EventTypeBan,
		GuildID:   guildID,
		ActorID:   userID,
		TargetID:  userID,
		Timestamp: util.NowMono(),
		Priority:  3,
		Flags:     0,
	}
}

func sliceMemberRemove(data []byte) *Event {
	guildID := extractU64Field(data, "guild_id")
	userID := extractU64Field(data, "user", "id")

	if guildID == 0 || userID == 0 {
		return nil
	}

	return &Event{
		EventType: EventTypeKick,
		GuildID:   guildID,
		ActorID:   0,
		TargetID:  userID,
		Timestamp: util.NowMono(),
		Priority:  2,
		Flags:     0,
	}
}

func sliceChannelDelete(data []byte) *Event {
	guildID := extractU64Field(data, "guild_id")
	channelID := extractU64Field(data, "id")

	if guildID == 0 || channelID == 0 {
		return nil
	}

	return &Event{
		EventType: EventTypeChannelDelete,
		GuildID:   guildID,
		ActorID:   0,
		TargetID:  channelID,
		Timestamp: util.NowMono(),
		Priority:  3,
		Flags:     0,
	}
}

func sliceRoleDelete(data []byte) *Event {
	guildID := extractU64Field(data, "guild_id")
	roleID := extractU64Field(data, "role_id")

	if guildID == 0 || roleID == 0 {
		return nil
	}

	return &Event{
		EventType: EventTypeRoleDelete,
		GuildID:   guildID,
		ActorID:   0,
		TargetID:  roleID,
		Timestamp: util.NowMono(),
		Priority:  3,
		Flags:     0,
	}
}

func sliceWebhookUpdate(data []byte) *Event {
	guildID := extractU64Field(data, "guild_id")
	channelID := extractU64Field(data, "channel_id")

	if guildID == 0 {
		return nil
	}

	return &Event{
		EventType: EventTypeWebhook,
		GuildID:   guildID,
		ActorID:   0,
		TargetID:  channelID,
		Timestamp: util.NowMono(),
		Priority:  2,
		Flags:     0,
	}
}

func sliceRoleUpdate(data []byte) *Event {
	guildID := extractU64Field(data, "guild_id")
	roleID := extractU64Field(data, "role", "id")
	permissions := extractU64Field(data, "role", "permissions")

	if guildID == 0 || roleID == 0 {
		return nil
	}

	return &Event{
		EventType: EventTypePermChange,
		GuildID:   guildID,
		ActorID:   0,
		TargetID:  roleID,
		Metadata:  permissions,
		Timestamp: util.NowMono(),
		Priority:  2,
		Flags:     0,
	}
}

func extractU64Field(data []byte, keys ...string) uint64 {
	s := findJSONValue(data, keys)
	if s == "" {
		return 0
	}
	if s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}

func findJSONValue(data []byte, keys []string) string {
	current := data

	for _, key := range keys {
		searchKey := []byte(`"` + key + `":`)
		idx := findBytes(current, searchKey)
		if idx == -1 {
			return ""
		}

		current = current[idx+len(searchKey):]

		for len(current) > 0 && (current[0] == ' ' || current[0] == '\t' || current[0] == '\n') {
			current = current[1:]
		}

		if len(current) == 0 {
			return ""
		}

		if current[0] == '{' {
			end := findMatchingBrace(current)
			if end == -1 {
				return ""
			}
			current = current[:end+1]
			continue
		}
	}

	return extractValue(current)
}

func findBytes(data, search []byte) int {
	for i := 0; i <= len(data)-len(search); i++ {
		match := true
		for j := 0; j < len(search); j++ {
			if data[i+j] != search[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func findMatchingBrace(data []byte) int {
	depth := 0
	for i, b := range data {
		if b == '{' {
			depth++
		} else if b == '}' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func extractValue(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	if data[0] == '"' {
		end := 1
		for end < len(data) && data[end] != '"' {
			if data[end] == '\\' {
				end++
			}
			end++
		}
		if end < len(data) {
			return util.BytesToString(data[:end+1])
		}
		return ""
	}

	end := 0
	for end < len(data) {
		c := data[end]
		if c == ',' || c == '}' || c == ']' || c == ' ' || c == '\t' || c == '\n' {
			break
		}
		end++
	}

	return util.BytesToString(data[:end])
}
