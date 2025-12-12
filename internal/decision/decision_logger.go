package decision

import (
	"encoding/json"
	"os"
	"time"
)

type DecisionLog struct {
	Timestamp  int64  `json:"timestamp"`
	GuildID    uint64 `json:"guild_id"`
	ActorID    uint64 `json:"actor_id"`
	Decision   string `json:"decision"`
	Reason     string `json:"reason"`
	Confidence uint8  `json:"confidence"`
}

type DecisionLogger struct {
	file *os.File
}

func NewDecisionLogger(path string) (*DecisionLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &DecisionLogger{file: file}, nil
}

func (dl *DecisionLogger) LogDecision(guildID, actorID uint64, decision, reason string, confidence uint8) error {
	entry := &DecisionLog{
		Timestamp:  time.Now().UnixNano(),
		GuildID:    guildID,
		ActorID:    actorID,
		Decision:   decision,
		Reason:     reason,
		Confidence: confidence,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = dl.file.Write(data)
	return err
}

func (dl *DecisionLogger) Close() error {
	return dl.file.Close()
}
