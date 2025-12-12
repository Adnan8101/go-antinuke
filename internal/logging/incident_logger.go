package logging

import (
	"encoding/json"
	"os"
	"time"
)

type IncidentLogEntry struct {
	Timestamp  int64  `json:"timestamp"`
	GuildID    uint64 `json:"guild_id"`
	ActorID    uint64 `json:"actor_id"`
	EventType  string `json:"event_type"`
	Severity   uint8  `json:"severity"`
	Action     string `json:"action"`
	Confidence uint8  `json:"confidence"`
}

type IncidentLogger struct {
	file *os.File
}

func NewIncidentLogger(path string) (*IncidentLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &IncidentLogger{file: file}, nil
}

func (il *IncidentLogger) Log(entry *IncidentLogEntry) error {
	entry.Timestamp = time.Now().UnixNano()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = il.file.Write(data)
	return err
}

func (il *IncidentLogger) Close() error {
	return il.file.Close()
}
