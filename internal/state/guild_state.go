package state

import (
	"sync/atomic"
)

const (
	MaxGuilds = 8192
	GuildMask = MaxGuilds - 1
)

type GuildCounters struct {
	BanCount       uint32
	KickCount      uint32
	ChannelDelete  uint32
	RoleDelete     uint32
	WebhookCreate  uint32
	PermChange     uint32
	MemberRemove   uint32
	Reserved       uint32
	LastBanTime    int64
	LastChanTime   int64
	LastRoleTime   int64
	LastWebTime    int64
	VelocityScore  uint32
	TriggerFlags   uint32
	LockdownActive uint32
	_              [20]byte
}

type GuildProfile struct {
	GuildID       uint64
	Size          uint32
	BanThreshold  uint32
	ChanThreshold uint32
	RoleThreshold uint32
	WebThreshold  uint32
	VelocityLimit uint32
	Enabled       uint32
	SafetyMode    uint32
	_             [32]byte
}

type GuildState struct {
	counters [MaxGuilds]GuildCounters
	profiles [MaxGuilds]GuildProfile
}

var globalGuildState *GuildState

func InitGuildState() {
	globalGuildState = &GuildState{}
}

func GetGuildState() *GuildState {
	return globalGuildState
}

func (g *GuildState) GetCounters(guildIndex uint32) *GuildCounters {
	return &g.counters[guildIndex&GuildMask]
}

func (g *GuildState) GetProfile(guildIndex uint32) *GuildProfile {
	return &g.profiles[guildIndex&GuildMask]
}

func (g *GuildState) IncrementBans(guildIndex uint32) uint32 {
	return atomic.AddUint32(&g.counters[guildIndex&GuildMask].BanCount, 1)
}

func (g *GuildState) GetChannelDeleteCount(guildIndex uint32) uint32 {
	return atomic.LoadUint32(&g.counters[guildIndex&GuildMask].ChannelDelete)
}

func (g *GuildState) IncrementChannelDeletes(guildIndex uint32) uint32 {
	return atomic.AddUint32(&g.counters[guildIndex&GuildMask].ChannelDelete, 1)
}

func (g *GuildState) IncrementRoleDeletes(guildIndex uint32) uint32 {
	return atomic.AddUint32(&g.counters[guildIndex&GuildMask].RoleDelete, 1)
}

func (g *GuildState) IncrementWebhooks(guildIndex uint32) uint32 {
	return atomic.AddUint32(&g.counters[guildIndex&GuildMask].WebhookCreate, 1)
}

func (g *GuildState) IncrementKicks(guildIndex uint32) uint32 {
	return atomic.AddUint32(&g.counters[guildIndex&GuildMask].KickCount, 1)
}

func (g *GuildState) IncrementPermChanges(guildIndex uint32) uint32 {
	return atomic.AddUint32(&g.counters[guildIndex&GuildMask].PermChange, 1)
}

func (g *GuildState) ResetCounters(guildIndex uint32) {
	c := &g.counters[guildIndex&GuildMask]
	atomic.StoreUint32(&c.BanCount, 0)
	atomic.StoreUint32(&c.KickCount, 0)
	atomic.StoreUint32(&c.ChannelDelete, 0)
	atomic.StoreUint32(&c.RoleDelete, 0)
	atomic.StoreUint32(&c.WebhookCreate, 0)
	atomic.StoreUint32(&c.PermChange, 0)
	atomic.StoreUint32(&c.MemberRemove, 0)
	atomic.StoreUint32(&c.VelocityScore, 0)
	atomic.StoreUint32(&c.TriggerFlags, 0)
}

func (g *GuildState) SetLockdown(guildIndex uint32, active bool) {
	val := uint32(0)
	if active {
		val = 1
	}
	atomic.StoreUint32(&g.counters[guildIndex&GuildMask].LockdownActive, val)
}

func (g *GuildState) IsLockdown(guildIndex uint32) bool {
	return atomic.LoadUint32(&g.counters[guildIndex&GuildMask].LockdownActive) == 1
}
