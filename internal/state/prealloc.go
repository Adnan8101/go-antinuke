package state

import (
	"go-antinuke-2.0/pkg/memory"
)

type PreallocatedState struct {
	GuildState   *GuildState
	ActorState   *ActorState
	HazardScores *HazardScores
	GuildIDMap   *GuildIDMap
	ActorIDMap   *ActorIDMap
}

var GlobalState *PreallocatedState

func InitAll() error {
	InitGuildState()
	InitActorState()
	InitHazardScores()
	InitGuildIDMap()
	InitActorIDMap()
	InitEventLookup()

	GlobalState = &PreallocatedState{
		GuildState:   GetGuildState(),
		ActorState:   GetActorState(),
		HazardScores: GetHazardScores(),
		GuildIDMap:   GetGuildIDMap(),
		ActorIDMap:   GetActorIDMap(),
	}

	return nil
}

func LockMemory() error {
	return memory.MlockAll()
}

func TouchAll() {
	gs := GetGuildState()
	for i := range gs.counters {
		gs.counters[i].BanCount = 0
		gs.profiles[i].GuildID = 0
	}

	as := GetActorState()
	for i := range as.counters {
		as.counters[i].BanCount = 0
		as.profiles[i].ActorID = 0
	}

	hs := GetHazardScores()
	for i := range hs.entries {
		hs.entries[i].GuildID = 0
	}
}

func ResetAllCounters() {
	gs := GetGuildState()
	for i := uint32(0); i < MaxGuilds; i++ {
		gs.ResetCounters(i)
	}

	as := GetActorState()
	for i := range as.counters {
		as.counters[i] = ActorCounters{}
	}

	hs := GetHazardScores()
	for i := uint32(0); i < MaxHazardEntries; i++ {
		hs.ResetEntry(i)
	}
}
