package state

import (
	"sync/atomic"
)

const (
	MaxActors = 16384
	ActorMask = MaxActors - 1
)

type ActorCounters struct {
	BanCount       uint32
	KickCount      uint32
	ChannelDelete  uint32
	RoleDelete     uint32
	WebhookCreate  uint32
	PermChange     uint32
	TotalActions   uint32
	ThreatLevel    uint32
	LastActionTime int64
	FirstSeenTime  int64
	FlagsSet       uint32
	Banned         uint32
	_              [16]byte
}

type ActorProfile struct {
	ActorID     uint64
	GuildID     uint64
	Whitelisted uint32
	TrustScore  uint32
	_           [48]byte
}

type ActorState struct {
	counters [MaxActors]ActorCounters
	profiles [MaxActors]ActorProfile
}

var globalActorState *ActorState

func InitActorState() {
	globalActorState = &ActorState{}
}

func GetActorState() *ActorState {
	return globalActorState
}

func (a *ActorState) GetCounters(actorIndex uint32) *ActorCounters {
	return &a.counters[actorIndex&ActorMask]
}

func (a *ActorState) GetProfile(actorIndex uint32) *ActorProfile {
	return &a.profiles[actorIndex&ActorMask]
}

func (a *ActorState) IncrementBans(actorIndex uint32) uint32 {
	atomic.AddUint32(&a.counters[actorIndex&ActorMask].TotalActions, 1)
	return atomic.AddUint32(&a.counters[actorIndex&ActorMask].BanCount, 1)
}

func (a *ActorState) GetChannelDeleteCount(actorIndex uint32) uint32 {
	return atomic.LoadUint32(&a.counters[actorIndex&ActorMask].ChannelDelete)
}

func (a *ActorState) IncrementChannelDeletes(actorIndex uint32) uint32 {
	atomic.AddUint32(&a.counters[actorIndex&ActorMask].TotalActions, 1)
	return atomic.AddUint32(&a.counters[actorIndex&ActorMask].ChannelDelete, 1)
}

func (a *ActorState) IncrementRoleDeletes(actorIndex uint32) uint32 {
	atomic.AddUint32(&a.counters[actorIndex&ActorMask].TotalActions, 1)
	return atomic.AddUint32(&a.counters[actorIndex&ActorMask].RoleDelete, 1)
}

func (a *ActorState) IncrementWebhooks(actorIndex uint32) uint32 {
	atomic.AddUint32(&a.counters[actorIndex&ActorMask].TotalActions, 1)
	return atomic.AddUint32(&a.counters[actorIndex&ActorMask].WebhookCreate, 1)
}

func (a *ActorState) UpdateThreatLevel(actorIndex, level uint32) {
	atomic.StoreUint32(&a.counters[actorIndex&ActorMask].ThreatLevel, level)
}

func (a *ActorState) GetThreatLevel(actorIndex uint32) uint32 {
	return atomic.LoadUint32(&a.counters[actorIndex&ActorMask].ThreatLevel)
}

func (a *ActorState) SetBanned(actorIndex uint32, banned bool) {
	val := uint32(0)
	if banned {
		val = 1
	}
	atomic.StoreUint32(&a.counters[actorIndex&ActorMask].Banned, val)
}

func (a *ActorState) TrySetBanned(actorIndex uint32) bool {
	return atomic.CompareAndSwapUint32(&a.counters[actorIndex&ActorMask].Banned, 0, 1)
}

func (a *ActorState) IsBanned(actorIndex uint32) bool {
	return atomic.LoadUint32(&a.counters[actorIndex&ActorMask].Banned) == 1
}

func (a *ActorState) IsTriggered(actorIndex uint32) bool {
	return (atomic.LoadUint32(&a.counters[actorIndex&ActorMask].FlagsSet) & 0x80000000) != 0
}

func (a *ActorState) SetTriggered(actorIndex uint32, triggered bool) {
	if triggered {
		for {
			old := atomic.LoadUint32(&a.counters[actorIndex&ActorMask].FlagsSet)
			new := old | 0x80000000
			if atomic.CompareAndSwapUint32(&a.counters[actorIndex&ActorMask].FlagsSet, old, new) {
				break
			}
		}
	} else {
		for {
			old := atomic.LoadUint32(&a.counters[actorIndex&ActorMask].FlagsSet)
			new := old & ^uint32(0x80000000)
			if atomic.CompareAndSwapUint32(&a.counters[actorIndex&ActorMask].FlagsSet, old, new) {
				break
			}
		}
	}
}

func (a *ActorState) IsWhitelisted(actorIndex uint32) bool {
	return atomic.LoadUint32(&a.profiles[actorIndex&ActorMask].Whitelisted) == 1
}

func (a *ActorState) SetWhitelisted(actorIndex uint32, whitelisted bool) {
	val := uint32(0)
	if whitelisted {
		val = 1
	}
	atomic.StoreUint32(&a.profiles[actorIndex&ActorMask].Whitelisted, val)
}

// ClearGuildActorStates clears all actor state when bot is re-added to guild
// This ensures previously banned users can be detected again
func ClearGuildActorStates(guildID uint64) {
	as := GetActorState()
	actorMap := GetActorIDMap()

	// Iterate through all registered actors
	for i := uint32(0); i < MaxActors; i++ {
		profile := as.GetProfile(i)
		if profile.GuildID == guildID {
			// Clear counters
			counters := as.GetCounters(i)
			atomic.StoreUint32(&counters.BanCount, 0)
			atomic.StoreUint32(&counters.KickCount, 0)
			atomic.StoreUint32(&counters.ChannelDelete, 0)
			atomic.StoreUint32(&counters.RoleDelete, 0)
			atomic.StoreUint32(&counters.WebhookCreate, 0)
			atomic.StoreUint32(&counters.PermChange, 0)
			atomic.StoreUint32(&counters.TotalActions, 0)
			atomic.StoreUint32(&counters.ThreatLevel, 0)
			atomic.StoreUint32(&counters.FlagsSet, 0)
			atomic.StoreUint32(&counters.Banned, 0)
			atomic.StoreInt64(&counters.LastActionTime, 0)
			atomic.StoreInt64(&counters.FirstSeenTime, 0)
		}
	}

	// Also clear actor ID mapping for this guild
	actorMap.ClearGuild(guildID)
}

// ClearActorState clears state for a specific actor
// This is used when a user is unbanned or rejoins
func ClearActorState(actorID uint64) {
	as := GetActorState()
	actorMap := GetActorIDMap()

	actorIndex := actorMap.GetIndex(actorID)
	if actorIndex == 0 {
		return
	}

	counters := as.GetCounters(actorIndex)
	atomic.StoreUint32(&counters.BanCount, 0)
	atomic.StoreUint32(&counters.KickCount, 0)
	atomic.StoreUint32(&counters.ChannelDelete, 0)
	atomic.StoreUint32(&counters.RoleDelete, 0)
	atomic.StoreUint32(&counters.WebhookCreate, 0)
	atomic.StoreUint32(&counters.PermChange, 0)
	atomic.StoreUint32(&counters.TotalActions, 0)
	atomic.StoreUint32(&counters.ThreatLevel, 0)
	atomic.StoreUint32(&counters.FlagsSet, 0)
	atomic.StoreUint32(&counters.Banned, 0)
	atomic.StoreInt64(&counters.LastActionTime, 0)
	atomic.StoreInt64(&counters.FirstSeenTime, 0)
}
