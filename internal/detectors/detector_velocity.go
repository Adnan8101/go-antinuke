package detectors

import (
	"go-antinuke-2.0/internal/state"
)

type VelocityDetector struct {
	lastCounts map[uint32]uint32
}

func NewVelocityDetector() *VelocityDetector {
	return &VelocityDetector{
		lastCounts: make(map[uint32]uint32, state.MaxGuilds),
	}
}

func (d *VelocityDetector) CheckBanVelocity(guildIndex uint32, threshold uint32) (bool, uint32) {
	gs := state.GetGuildState()
	counters := gs.GetCounters(guildIndex)

	currentCount := counters.BanCount
	lastCount, exists := d.lastCounts[guildIndex]

	if !exists {
		d.lastCounts[guildIndex] = currentCount
		return false, 0
	}

	delta := currentCount - lastCount
	d.lastCounts[guildIndex] = currentCount

	triggered := BranchlessGreaterEqual(delta, threshold)

	return triggered != 0, delta
}

func (d *VelocityDetector) CheckChannelVelocity(guildIndex uint32, threshold uint32) (bool, uint32) {
	gs := state.GetGuildState()
	counters := gs.GetCounters(guildIndex)

	currentCount := counters.ChannelDelete
	baseIndex := guildIndex + state.MaxGuilds
	lastCount, exists := d.lastCounts[baseIndex]

	if !exists {
		d.lastCounts[baseIndex] = currentCount
		return false, 0
	}

	delta := currentCount - lastCount
	d.lastCounts[baseIndex] = currentCount

	triggered := BranchlessGreaterEqual(delta, threshold)

	return triggered != 0, delta
}

func (d *VelocityDetector) CheckRoleVelocity(guildIndex uint32, threshold uint32) (bool, uint32) {
	gs := state.GetGuildState()
	counters := gs.GetCounters(guildIndex)

	currentCount := counters.RoleDelete
	baseIndex := guildIndex + (state.MaxGuilds * 2)
	lastCount, exists := d.lastCounts[baseIndex]

	if !exists {
		d.lastCounts[baseIndex] = currentCount
		return false, 0
	}

	delta := currentCount - lastCount
	d.lastCounts[baseIndex] = currentCount

	triggered := BranchlessGreaterEqual(delta, threshold)

	return triggered != 0, delta
}

func (d *VelocityDetector) Reset(guildIndex uint32) {
	delete(d.lastCounts, guildIndex)
	delete(d.lastCounts, guildIndex+state.MaxGuilds)
	delete(d.lastCounts, guildIndex+(state.MaxGuilds*2))
}

func (d *VelocityDetector) ResetAll() {
	d.lastCounts = make(map[uint32]uint32, state.MaxGuilds)
}
