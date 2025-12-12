package state

type GuildSizeBucket uint8

const (
	BucketTiny GuildSizeBucket = iota
	BucketSmall
	BucketMedium
	BucketLarge
	BucketHuge
)

type ThresholdSet struct {
	BanThreshold  uint32
	ChanThreshold uint32
	RoleThreshold uint32
	WebThreshold  uint32
	VelocityLimit uint32
}

var DefaultThresholds = [5]ThresholdSet{
	{BanThreshold: 3, ChanThreshold: 2, RoleThreshold: 2, WebThreshold: 5, VelocityLimit: 10},
	{BanThreshold: 5, ChanThreshold: 3, RoleThreshold: 3, WebThreshold: 8, VelocityLimit: 15},
	{BanThreshold: 7, ChanThreshold: 5, RoleThreshold: 5, WebThreshold: 10, VelocityLimit: 20},
	{BanThreshold: 10, ChanThreshold: 7, RoleThreshold: 7, WebThreshold: 15, VelocityLimit: 30},
	{BanThreshold: 15, ChanThreshold: 10, RoleThreshold: 10, WebThreshold: 20, VelocityLimit: 40},
}

func GetGuildBucket(size uint32) GuildSizeBucket {
	switch {
	case size < 100:
		return BucketTiny
	case size < 1000:
		return BucketSmall
	case size < 5000:
		return BucketMedium
	case size < 20000:
		return BucketLarge
	default:
		return BucketHuge
	}
}

func GetThresholds(bucket GuildSizeBucket) ThresholdSet {
	return DefaultThresholds[bucket]
}

type EventTypeLookup [256]uint8

const (
	EventTypeUnknown = iota
	EventTypeBan
	EventTypeKick
	EventTypeChannelDelete
	EventTypeRoleDelete
	EventTypeWebhookCreate
	EventTypePermissionChange
	EventTypeMemberRemove
)

var GlobalEventLookup EventTypeLookup

func InitEventLookup() {
	GlobalEventLookup[0] = EventTypeUnknown
}

func GetEventType(typeID uint8) uint8 {
	return GlobalEventLookup[typeID]
}

type GuildIDMap struct {
	ids    [MaxGuilds]uint64
	lookup map[uint64]uint32
}

var globalGuildIDMap *GuildIDMap

func InitGuildIDMap() {
	globalGuildIDMap = &GuildIDMap{
		lookup: make(map[uint64]uint32, MaxGuilds),
	}
}

func GetGuildIDMap() *GuildIDMap {
	return globalGuildIDMap
}

func (g *GuildIDMap) Register(guildID uint64) uint32 {
	if idx, exists := g.lookup[guildID]; exists {
		return idx
	}

	// Reserve index 0 as "not found" sentinel.
	// Indices start at 1 to avoid the entire pipeline collapsing into guildIndex=0.
	idx := uint32(len(g.lookup) + 1)
	if idx >= MaxGuilds {
		return 0
	}

	g.ids[idx] = guildID
	g.lookup[guildID] = idx
	return idx
}

func (g *GuildIDMap) GetIndex(guildID uint64) uint32 {
	if idx, exists := g.lookup[guildID]; exists {
		return idx
	}
	return 0
}

type ActorIDMap struct {
	ids    [MaxActors]uint64
	lookup map[uint64]uint32
}

var globalActorIDMap *ActorIDMap

func InitActorIDMap() {
	globalActorIDMap = &ActorIDMap{
		lookup: make(map[uint64]uint32, MaxActors),
	}
}

func GetActorIDMap() *ActorIDMap {
	return globalActorIDMap
}

func (a *ActorIDMap) Register(actorID uint64) uint32 {
	if idx, exists := a.lookup[actorID]; exists {
		return idx
	}

	// Reserve index 0 as "not found" sentinel.
	// Indices start at 1 to avoid actorIndex=0 being treated as both valid and invalid.
	idx := uint32(len(a.lookup) + 1)
	if idx >= MaxActors {
		return 0
	}

	a.ids[idx] = actorID
	a.lookup[actorID] = idx
	return idx
}

func (a *ActorIDMap) GetIndex(actorID uint64) uint32 {
	if idx, exists := a.lookup[actorID]; exists {
		return idx
	}
	return 0
}

// Bot ID storage
var globalBotID uint64

func SetBotID(botID uint64) {
	globalBotID = botID
}

func GetBotID() uint64 {
	return globalBotID
}

// ClearGuild removes all actor mappings for a specific guild
// Called when bot is removed and re-added to reset state
func (a *ActorIDMap) ClearGuild(guildID uint64) {
	// Note: In current implementation, we don't track guild per actor in the map
	// The GuildID is stored in ActorProfile. This is a placeholder for future enhancement
	// For now, we rely on ActorState counter clearing
}
