package forensics

import (
	"encoding/json"
	"os"
	"time"
)

type ForensicLog struct {
	Timestamp int64                  `json:"timestamp"`
	EventType string                 `json:"event_type"`
	GuildID   uint64                 `json:"guild_id"`
	ActorID   uint64                 `json:"actor_id"`
	TargetID  uint64                 `json:"target_id"`
	Severity  uint8                  `json:"severity"`
	Data      map[string]interface{} `json:"data"`
}

type ForensicLogger struct {
	file *os.File
	path string
}

func NewForensicLogger(path string) (*ForensicLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &ForensicLogger{
		file: file,
		path: path,
	}, nil
}

func (fl *ForensicLogger) Log(entry *ForensicLog) error {
	entry.Timestamp = time.Now().UnixNano()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = fl.file.Write(data)
	return err
}

func (fl *ForensicLogger) Close() error {
	return fl.file.Close()
}
