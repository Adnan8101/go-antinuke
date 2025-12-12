package forensics

import (
	"sync"
	"time"
)

type EntityChange struct {
	GuildID    uint64
	EntityID   uint64
	EntityType string
	Action     string
	ActorID    uint64
	Timestamp  int64
	Name       string
	Properties map[string]interface{}
}

type RecoveryTracker struct {
	mu      sync.RWMutex
	changes map[uint64][]*EntityChange
}

var globalRecoveryTracker *RecoveryTracker

func InitRecoveryTracker() {
	globalRecoveryTracker = &RecoveryTracker{
		changes: make(map[uint64][]*EntityChange),
	}
}

func GetRecoveryTracker() *RecoveryTracker {
	if globalRecoveryTracker == nil {
		InitRecoveryTracker()
	}
	return globalRecoveryTracker
}

func (rt *RecoveryTracker) TrackChannelDelete(guildID, channelID, actorID uint64, name string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	change := &EntityChange{
		GuildID:    guildID,
		EntityID:   channelID,
		EntityType: "channel",
		Action:     "delete",
		ActorID:    actorID,
		Timestamp:  time.Now().UnixNano(),
		Name:       name,
	}

	rt.changes[guildID] = append(rt.changes[guildID], change)
}

func (rt *RecoveryTracker) TrackChannelCreate(guildID, channelID, actorID uint64, name string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	change := &EntityChange{
		GuildID:    guildID,
		EntityID:   channelID,
		EntityType: "channel",
		Action:     "create",
		ActorID:    actorID,
		Timestamp:  time.Now().UnixNano(),
		Name:       name,
	}

	rt.changes[guildID] = append(rt.changes[guildID], change)
}

func (rt *RecoveryTracker) TrackRoleDelete(guildID, roleID, actorID uint64, name string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	change := &EntityChange{
		GuildID:    guildID,
		EntityID:   roleID,
		EntityType: "role",
		Action:     "delete",
		ActorID:    actorID,
		Timestamp:  time.Now().UnixNano(),
		Name:       name,
	}

	rt.changes[guildID] = append(rt.changes[guildID], change)
}

func (rt *RecoveryTracker) TrackRoleCreate(guildID, roleID, actorID uint64, name string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	change := &EntityChange{
		GuildID:    guildID,
		EntityID:   roleID,
		EntityType: "role",
		Action:     "create",
		ActorID:    actorID,
		Timestamp:  time.Now().UnixNano(),
		Name:       name,
	}

	rt.changes[guildID] = append(rt.changes[guildID], change)
}

func (rt *RecoveryTracker) GetMaliciousChanges(guildID, actorID uint64, since int64) []*EntityChange {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	var result []*EntityChange
	changes := rt.changes[guildID]

	for _, change := range changes {
		if change.ActorID == actorID && change.Timestamp >= since {
			result = append(result, change)
		}
	}

	return result
}

func (rt *RecoveryTracker) GetDeletedChannels(guildID, actorID uint64, since int64) []uint64 {
	changes := rt.GetMaliciousChanges(guildID, actorID, since)
	var channelIDs []uint64

	for _, change := range changes {
		if change.EntityType == "channel" && change.Action == "delete" {
			channelIDs = append(channelIDs, change.EntityID)
		}
	}

	return channelIDs
}

func (rt *RecoveryTracker) GetCreatedChannels(guildID, actorID uint64, since int64) []uint64 {
	changes := rt.GetMaliciousChanges(guildID, actorID, since)
	var channelIDs []uint64

	for _, change := range changes {
		if change.EntityType == "channel" && change.Action == "create" {
			channelIDs = append(channelIDs, change.EntityID)
		}
	}

	return channelIDs
}

func (rt *RecoveryTracker) ClearGuildChanges(guildID uint64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	delete(rt.changes, guildID)
}
