package config

import (
	"sync"
)

type GuildProfile struct {
	GuildID          uint64
	Name             string
	MemberCount      uint32
	Enabled          bool
	SafetyMode       SafetyMode
	PanicMode        bool
	OwnerID          uint64
	Whitelist        []uint64
	TrustedRoles     []uint64
	CustomThresholds *ThresholdMatrix
}

type ProfileStore struct {
	mu       sync.RWMutex
	profiles map[uint64]*GuildProfile
}

var GlobalProfiles *ProfileStore

func InitGuildProfiles() {
	GlobalProfiles = &ProfileStore{
		profiles: make(map[uint64]*GuildProfile),
	}
}

func GetProfileStore() *ProfileStore {
	if GlobalProfiles == nil {
		InitGuildProfiles()
	}
	return GlobalProfiles
}

func (ps *ProfileStore) Get(guildID uint64) *GuildProfile {
	ps.mu.RLock()
	profile := ps.profiles[guildID]
	ps.mu.RUnlock()

	if profile == nil {
		return ps.GetOrCreate(guildID)
	}
	return profile
}

func (ps *ProfileStore) Set(profile *GuildProfile) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.profiles[profile.GuildID] = profile
}

func (ps *ProfileStore) GetOrCreate(guildID uint64) *GuildProfile {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if profile, exists := ps.profiles[guildID]; exists {
		return profile
	}

	profile := &GuildProfile{
		GuildID:      guildID,
		Enabled:      true,
		SafetyMode:   SafetyNormal,
		PanicMode:    false,
		Whitelist:    make([]uint64, 0),
		TrustedRoles: make([]uint64, 0),
	}
	ps.profiles[guildID] = profile
	return profile
}

func (ps *ProfileStore) IsWhitelisted(guildID, userID uint64) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	profile, exists := ps.profiles[guildID]
	if !exists {
		return false
	}

	for _, wid := range profile.Whitelist {
		if wid == userID {
			return true
		}
	}
	return false
}

func (ps *ProfileStore) AddWhitelist(guildID, userID uint64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	profile := ps.profiles[guildID]
	if profile == nil {
		profile = &GuildProfile{
			GuildID:    guildID,
			Enabled:    true,
			SafetyMode: SafetyNormal,
			PanicMode:  false,
			Whitelist:  []uint64{userID},
		}
		ps.profiles[guildID] = profile
		return
	}

	for _, wid := range profile.Whitelist {
		if wid == userID {
			return
		}
	}
	profile.Whitelist = append(profile.Whitelist, userID)
}

func (ps *ProfileStore) RemoveWhitelist(guildID, userID uint64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	profile := ps.profiles[guildID]
	if profile == nil {
		return
	}

	for i, wid := range profile.Whitelist {
		if wid == userID {
			profile.Whitelist = append(profile.Whitelist[:i], profile.Whitelist[i+1:]...)
			return
		}
	}
}

func (ps *ProfileStore) IsEnabled(guildID uint64) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	profile, exists := ps.profiles[guildID]
	if !exists {
		return true
	}
	return profile.Enabled
}

func (ps *ProfileStore) SetEnabled(guildID uint64, enabled bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	profile := ps.profiles[guildID]
	if profile == nil {
		profile = &GuildProfile{
			GuildID:   guildID,
			Enabled:   enabled,
			PanicMode: false,
		}
		ps.profiles[guildID] = profile
		return
	}
	profile.Enabled = enabled
}

func (ps *ProfileStore) SetPanicMode(guildID uint64, enabled bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	profile := ps.profiles[guildID]
	if profile == nil {
		profile = &GuildProfile{
			GuildID:   guildID,
			Enabled:   true,
			PanicMode: enabled,
		}
		ps.profiles[guildID] = profile
		return
	}
	profile.PanicMode = enabled
}

func (ps *ProfileStore) IsPanicMode(guildID uint64) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	profile, exists := ps.profiles[guildID]
	if !exists {
		return false
	}
	return profile.PanicMode
}
