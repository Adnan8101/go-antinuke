package forensics

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	GuildID     uint64                 `json:"guild_id"`
	Timestamp   int64                  `json:"timestamp"`
	Channels    []ChannelSnapshot      `json:"channels"`
	Roles       []RoleSnapshot         `json:"roles"`
	Permissions map[string]uint64      `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ChannelSnapshot struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	Position int    `json:"position"`
}

type RoleSnapshot struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Permissions uint64 `json:"permissions"`
	Position    int    `json:"position"`
}

type SnapshotStore struct {
	snapshots map[uint64]*Snapshot
}

func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		snapshots: make(map[uint64]*Snapshot),
	}
}

func (ss *SnapshotStore) Capture(guildID uint64) *Snapshot {
	snapshot := &Snapshot{
		GuildID:     guildID,
		Timestamp:   time.Now().UnixNano(),
		Channels:    make([]ChannelSnapshot, 0),
		Roles:       make([]RoleSnapshot, 0),
		Permissions: make(map[string]uint64),
		Metadata:    make(map[string]interface{}),
	}

	ss.snapshots[guildID] = snapshot
	return snapshot
}

func (ss *SnapshotStore) Get(guildID uint64) *Snapshot {
	return ss.snapshots[guildID]
}

func (ss *SnapshotStore) Serialize(snapshot *Snapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}

func (ss *SnapshotStore) Deserialize(data []byte) (*Snapshot, error) {
	var snapshot Snapshot
	err := json.Unmarshal(data, &snapshot)
	return &snapshot, err
}
