package logging

import (
	"encoding/json"
	"os"
	"time"
)

type PerformanceEntry struct {
	Timestamp       int64  `json:"timestamp"`
	Component       string `json:"component"`
	OperationNs     int64  `json:"operation_ns"`
	OperationUs     int64  `json:"operation_us"`
	QueueDepth      uint32 `json:"queue_depth"`
	ProcessedEvents uint64 `json:"processed_events"`
}

type PerformanceLogger struct {
	file *os.File
}

func NewPerformanceLogger(path string) (*PerformanceLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &PerformanceLogger{file: file}, nil
}

func (pl *PerformanceLogger) LogLatency(component string, latencyNs int64) error {
	entry := &PerformanceEntry{
		Timestamp:   time.Now().UnixNano(),
		Component:   component,
		OperationNs: latencyNs,
		OperationUs: latencyNs / 1000,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = pl.file.Write(data)
	return err
}

func (pl *PerformanceLogger) Close() error {
	return pl.file.Close()
}
