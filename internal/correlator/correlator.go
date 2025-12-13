package correlator

import (
	"fmt"
	"runtime"

	"go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/detectors"
	"go-antinuke-2.0/internal/ingest"
	"go-antinuke-2.0/internal/state"
	"go-antinuke-2.0/internal/sys"
	"go-antinuke-2.0/pkg/util"
)

type Correlator struct {
	ringBuffer         *ingest.RingBuffer
	alertQueue         *AlertQueue
	banDetector        *detectors.BanDetector
	channelDetector    *detectors.ChannelDeleteDetector
	roleDetector       *detectors.RoleDeleteDetector
	permDetector       *detectors.PermissionDetector
	velocityDetector   *detectors.VelocityDetector
	multiActorDetector *detectors.MultiActorDetector
	flagDetector       *detectors.FlagDetector
	running            bool
	cpuCore            int
}

func NewCorrelator(ringBuffer *ingest.RingBuffer, alertQueue *AlertQueue, cpuCore int) *Correlator {
	return &Correlator{
		ringBuffer:         ringBuffer,
		alertQueue:         alertQueue,
		banDetector:        detectors.NewBanDetector(),
		channelDetector:    detectors.NewChannelDeleteDetector(),
		roleDetector:       detectors.NewRoleDeleteDetector(),
		permDetector:       detectors.NewPermissionDetector(),
		velocityDetector:   detectors.NewVelocityDetector(),
		multiActorDetector: detectors.NewMultiActorDetector(),
		flagDetector:       detectors.NewFlagDetector(),
		running:            false,
		cpuCore:            cpuCore,
	}
}

func (c *Correlator) Start() {
	if err := sys.PinToCore(c.cpuCore); err != nil {
		// logging.Warn("Failed to pin correlator to core %d: %v", c.cpuCore, err)
	}
	runtime.LockOSThread()
	c.running = true
	c.runLoop()
}

func (c *Correlator) runLoop() {
	guildMap := state.GetGuildIDMap()
	actorMap := state.GetActorIDMap()
	profileStore := config.GetProfileStore()

	// Batch processing for better throughput
	const batchSize = 64
	eventBatch := make([]*ingest.Event, 0, batchSize)

	for c.running {
		// Try to fill batch
		for len(eventBatch) < batchSize {
			event, ok := c.ringBuffer.Dequeue()
			if !ok {
				break
			}
			eventBatch = append(eventBatch, event)
		}

		// If no events, continue (spin-wait for lowest latency)
		if len(eventBatch) == 0 {
			continue
		}

		// Process batch
		for _, event := range eventBatch {
			c.processEvent(event, guildMap, actorMap, profileStore)
		}

		// Clear batch for reuse
		eventBatch = eventBatch[:0]
	}
}

func (c *Correlator) processEvent(event *ingest.Event, guildMap *state.GuildIDMap, actorMap *state.ActorIDMap, profileStore *config.ProfileStore) {
	timestamp := util.NowMono()

	guildIndex := guildMap.GetIndex(event.GuildID)
	if guildIndex == 0 {
		guildIndex = guildMap.Register(event.GuildID)
	}

	actorIndex := uint32(0)
	if event.ActorID != 0 {
		actorIndex = actorMap.GetIndex(event.ActorID)
		if actorIndex == 0 {
			actorIndex = actorMap.Register(event.ActorID)
		}
	} else {
		return
	}

	profile := profileStore.Get(event.GuildID)
	if profile == nil || !profile.Enabled {
		// fmt.Printf("[CORRELATOR] Skipping event - profile disabled for guild %d\n", event.GuildID)
		return
	}

	// CRITICAL: Never punish the bot itself or server owner
	if event.ActorID == state.GetBotID() {
		// fmt.Printf("[CORRELATOR] Skipping event - actor is the bot itself (%d)\n", event.ActorID)
		return
	}

	if profile.OwnerID != 0 && event.ActorID == profile.OwnerID {
		// fmt.Printf("[CORRELATOR] Skipping event - actor is server owner (%d)\n", event.ActorID)
		return
	}

	if actorIndex != 0 && profileStore.IsWhitelisted(event.GuildID, event.ActorID) {
		// fmt.Printf("[CORRELATOR] Skipping event - actor %d is whitelisted\n", event.ActorID)
		return
	}

	as := state.GetActorState()

	// In panic mode, check if actor is already triggered OR banned to skip processing
	// This prevents race conditions where multiple events slip through before ban executes
	if profile.PanicMode {
		if as.IsBanned(actorIndex) {
			fmt.Printf("[CORRELATOR] PANIC MODE - Actor %d already banned, skipping event\n", event.ActorID)
			return
		}
		if as.IsTriggered(actorIndex) {
			fmt.Printf("[CORRELATOR] PANIC MODE - Actor %d already triggered (ban pending), skipping event\n", event.ActorID)
			return
		}
	}

	// In normal mode, check if already triggered
	alreadyTriggered := as.IsTriggered(actorIndex)

	// PANIC MODE: Ultra-fast path - skip all unnecessary operations
	if profile.PanicMode {
		// CRITICAL: Mark actor as triggered+banned IMMEDIATELY to block race conditions
		as.SetTriggered(actorIndex, true)
		as.SetBanned(actorIndex, true)

		// Fast-path: Directly set flag based on event type without calling detectors
		var flag uint32
		switch event.EventType {
		case ingest.EventTypeBan:
			flag = detectors.FlagBanTriggered
		case ingest.EventTypeChannelCreate, ingest.EventTypeChannelDelete:
			flag = detectors.FlagChannelTriggered
		case ingest.EventTypeRoleCreate, ingest.EventTypeRoleDelete:
			flag = detectors.FlagRoleTriggered
		default:
			flag = detectors.FlagBanTriggered // Default to ban for any malicious event
		}

		// Queue alert immediately with zero-cost timestamp
		alert := c.alertQueue.Get()
		alert.GuildID = event.GuildID
		alert.ActorID = event.ActorID
		alert.EventType = event.EventType
		alert.Flags = flag
		alert.Timestamp = 0 // Skip timing in panic mode for max speed
		alert.Severity = detectors.GetSeverityFromFlags(flag)
		alert.PanicMode = 1
		c.alertQueue.Enqueue(alert)
		return // Skip normal detection path
	}

	// NORMAL MODE: Full detection with thresholds
	thresholds := config.GetGuildThresholds(event.GuildID, profile.MemberCount)

	if alreadyTriggered {
		return
	}

	detectionStart := util.NowMono()
	flags := uint32(0)

	switch event.EventType {
	case ingest.EventTypeBan:
		triggered, _ := c.banDetector.Detect(guildIndex, actorIndex, timestamp, thresholds.BanThreshold)
		if triggered {
			flags = c.flagDetector.SetFlag(flags, detectors.FlagBanTriggered)
		}

	case ingest.EventTypeChannelCreate:
		triggered, _ := c.channelDetector.Detect(guildIndex, actorIndex, timestamp, thresholds.ChannelThreshold)
		fmt.Printf("[CORRELATOR] Channel create detected - triggered=%v, threshold=%d\n", triggered, thresholds.ChannelThreshold)
		if triggered {
			flags = c.flagDetector.SetFlag(flags, detectors.FlagChannelTriggered)
			fmt.Printf("[CORRELATOR] FLAGS SET! Creating alert for actor %d\n", event.ActorID)
		}

	case ingest.EventTypeChannelDelete:
		triggered, _ := c.channelDetector.Detect(guildIndex, actorIndex, timestamp, thresholds.ChannelThreshold)
		if triggered {
			flags = c.flagDetector.SetFlag(flags, detectors.FlagChannelTriggered)
		}

	case ingest.EventTypeRoleCreate:
		triggered, _ := c.roleDetector.Detect(guildIndex, actorIndex, timestamp, thresholds.RoleThreshold)
		if triggered {
			flags = c.flagDetector.SetFlag(flags, detectors.FlagRoleTriggered)
		}

	case ingest.EventTypeRoleDelete:
		triggered, _ := c.roleDetector.Detect(guildIndex, actorIndex, timestamp, thresholds.RoleThreshold)
		if triggered {
			flags = c.flagDetector.SetFlag(flags, detectors.FlagRoleTriggered)
		}
	}

	if flags != 0 {
		// Normal mode: set triggered flag and queue alert
		as.SetTriggered(actorIndex, true)

		detectionTime := util.NowMono() - detectionStart

		alert := c.alertQueue.Get()
		alert.GuildID = event.GuildID
		alert.ActorID = event.ActorID
		alert.EventType = event.EventType
		alert.Flags = flags
		alert.Timestamp = detectionTime
		alert.Severity = detectors.GetSeverityFromFlags(flags)
		alert.PanicMode = 0
		c.alertQueue.Enqueue(alert)
	}
}

func (c *Correlator) Stop() {
	c.running = false
}

func (c *Correlator) IsRunning() bool {
	return c.running
}
