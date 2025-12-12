package forensics

import (
	"encoding/json"
	"os"
)

type ForensicIndex struct {
	entries map[uint64][]int64
	path    string
}

func NewForensicIndex(path string) *ForensicIndex {
	return &ForensicIndex{
		entries: make(map[uint64][]int64),
		path:    path,
	}
}

func (fi *ForensicIndex) AddEntry(guildID uint64, timestamp int64) {
	if _, exists := fi.entries[guildID]; !exists {
		fi.entries[guildID] = make([]int64, 0)
	}
	fi.entries[guildID] = append(fi.entries[guildID], timestamp)
}

func (fi *ForensicIndex) Query(guildID uint64, startTime, endTime int64) []int64 {
	entries, exists := fi.entries[guildID]
	if !exists {
		return nil
	}

	result := make([]int64, 0)
	for _, ts := range entries {
		if ts >= startTime && ts <= endTime {
			result = append(result, ts)
		}
	}

	return result
}

func (fi *ForensicIndex) Save() error {
	data, err := json.Marshal(fi.entries)
	if err != nil {
		return err
	}

	return os.WriteFile(fi.path, data, 0644)
}

func (fi *ForensicIndex) Load() error {
	data, err := os.ReadFile(fi.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &fi.entries)
}
