package decision

import (
	"fmt"
	"runtime"

	"go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/correlator"
	"go-antinuke-2.0/internal/forensics"
	"go-antinuke-2.0/internal/logging"
	"go-antinuke-2.0/internal/state"
)

type DecisionEngine struct {
	alertQueue  *correlator.AlertQueue
	jobQueue    *JobQueue
	forensicLog *forensics.ForensicLogger
	running     bool
	cpuCore     int
}

func NewDecisionEngine(alertQueue *correlator.AlertQueue, jobQueue *JobQueue, cpuCore int) *DecisionEngine {
	// Try to initialize forensic logger, but don't fail if it can't be created
	forensicLog, err := forensics.NewForensicLogger("logs/decisions.log")
	if err != nil {
		logging.Warn("Failed to create forensic logger in decision engine: %v", err)
	}

	return &DecisionEngine{
		alertQueue:  alertQueue,
		jobQueue:    jobQueue,
		forensicLog: forensicLog,
		running:     false,
		cpuCore:     cpuCore,
	}
}

func (de *DecisionEngine) Start() {
	runtime.LockOSThread()
	de.running = true
	de.runLoop()
}

func (de *DecisionEngine) runLoop() {
	for de.running {
		alert, ok := de.alertQueue.Dequeue()
		if !ok {
			runtime.Gosched()
			continue
		}

		fmt.Printf("[DECISION] Alert received! Actor=%d, Flags=%d, PanicMode=%d\n", alert.ActorID, alert.Flags, alert.PanicMode)
		incident := de.processAlert(alert)
		if incident != nil {
			fmt.Printf("[DECISION] Executing ban for actor %d\n", incident.ActorID)
			de.executeDecision(incident)
		} else {
			fmt.Printf("[DECISION] No incident created\n")
		}
	}
}

func (de *DecisionEngine) processAlert(alert *correlator.Alert) *IncidentPacket {
	profile := config.GetProfileStore().Get(alert.GuildID)
	if profile == nil {
		return nil
	}

	safetyMode := profile.SafetyMode
	severity := EvaluateSeverity(alert.Flags, alert.Severity)

	incident := &IncidentPacket{
		GuildID:    alert.GuildID,
		ActorID:    alert.ActorID,
		TargetID:   alert.TargetID,
		EventType:  alert.EventType,
		Severity:   severity,
		Confidence: 95,
		Timestamp:  alert.Timestamp,
		Flags:      alert.Flags,
		SafetyMode: uint8(safetyMode),
		PanicMode:  alert.PanicMode,
	}

	return incident
}

func (de *DecisionEngine) executeDecision(incident *IncidentPacket) {
	safetyMode := config.SafetyMode(incident.SafetyMode)

	shouldBan := config.ShouldAutoBan(safetyMode) && incident.Severity >= uint8(SeverityHigh)
	shouldLockdown := config.ShouldAutoLockdown(safetyMode) && incident.Severity >= uint8(SeverityCritical)
	shouldQuarantine := config.ShouldQuarantine(safetyMode) && incident.Severity >= uint8(SeverityMedium)

	// In panic mode, immediately kick the user to stop ongoing actions
	// Kick is faster than ban and removes them from server instantly
	if incident.PanicMode == 1 && shouldBan {
		fmt.Printf("[DECISION] PANIC MODE - Issuing immediate kick for actor %d\n", incident.ActorID)
		kickJob := NewKickJob(incident.GuildID, incident.ActorID, "PANIC MODE - Immediate Removal")
		de.jobQueue.Enqueue(kickJob)
	}

	// Log to forensics
	if de.forensicLog != nil {
		eventType := "unknown"
		switch incident.EventType {
		case 0:
			eventType = "ban"
		case 1:
			eventType = "channel_delete"
		case 2:
			eventType = "role_delete"
		}

		de.forensicLog.Log(&forensics.ForensicLog{
			EventType: eventType,
			GuildID:   incident.GuildID,
			ActorID:   incident.ActorID,
			TargetID:  incident.TargetID,
			Severity:  incident.Severity,
			Data: map[string]interface{}{
				"flags":      incident.Flags,
				"confidence": incident.Confidence,
				"ban":        shouldBan,
				"lockdown":   shouldLockdown,
				"quarantine": shouldQuarantine,
			},
		})
	}

	if shouldBan {
		as := state.GetActorState()
		actorMap := state.GetActorIDMap()

		actorIndex := actorMap.GetIndex(incident.ActorID)
		if actorIndex == 0 {
			actorIndex = actorMap.Register(incident.ActorID)
		}

		// In panic mode, ALWAYS queue ban jobs even if already marked as banned
		// This ensures redundant ban attempts for critical threats
		if incident.PanicMode == 0 {
			// Normal mode: check if already banned to avoid duplicate jobs
			if as.IsBanned(actorIndex) {
				fmt.Printf("[DECISION] Actor %d already banned, skipping (normal mode)\n", incident.ActorID)
				return
			}
		} else {
			// Panic mode: always attempt ban, even if already queued
			fmt.Printf("[DECISION] PANIC MODE - Forcing ban attempt for actor %d\n", incident.ActorID)
		}

		as.SetBanned(actorIndex, true)

		reason := de.getBanReason(incident)
		job := NewBanJob(incident.GuildID, incident.ActorID, reason, incident.EventType, incident.PanicMode, incident.Timestamp)
		de.jobQueue.Enqueue(job)
	}

	if shouldLockdown {
		reason := "Critical Threat Detected - Emergency Server Lockdown Activated"
		if incident.PanicMode == 1 {
			reason = "Panic Mode - Critical Threat - Emergency Server Lockdown"
		}
		job := NewLockdownJob(incident.GuildID, reason)
		de.jobQueue.Enqueue(job)
	}

	if shouldQuarantine && !shouldBan {
		reason := "Suspicious Activity Detected - Member Quarantined"
		if incident.PanicMode == 1 {
			reason = "Panic Mode - Suspicious Activity - Member Quarantined"
		}
		job := NewQuarantineJob(incident.GuildID, incident.ActorID, reason)
		de.jobQueue.Enqueue(job)
	}
}

func (de *DecisionEngine) getBanReason(incident *IncidentPacket) string {
	eventName := ""
	switch incident.EventType {
	case 1:
		eventName = "Mass Ban Detection"
	case 10:
		eventName = "Channel Create Spam"
	case 11:
		eventName = "Channel Delete Attack"
	case 18:
		eventName = "Role Create Spam"
	case 19:
		eventName = "Role Delete Attack"
	case 31:
		eventName = "Webhook Spam"
	case 32:
		eventName = "Permission Escalation"
	default:
		eventName = "Malicious Activity"
	}

	if incident.PanicMode == 1 {
		return fmt.Sprintf("Panic Mode - %s - Instant Ban Enforced", eventName)
	}

	if incident.Severity >= uint8(SeverityCritical) {
		return fmt.Sprintf("Critical Threat - %s - Auto-Ban Protection", eventName)
	}

	if incident.Severity >= uint8(SeverityHigh) {
		return fmt.Sprintf("High Severity - %s - Anti-Nuke Protection", eventName)
	}

	return fmt.Sprintf("%s - Automated Protection", eventName)
}

func (de *DecisionEngine) Stop() {
	de.running = false
	if de.forensicLog != nil {
		de.forensicLog.Close()
	}
}

func (de *DecisionEngine) IsRunning() bool {
	return de.running
}

type JobQueue struct {
	jobs []Job
	mask uint32
	head uint32
	tail uint32
}

func NewJobQueue(size uint32) *JobQueue {
	if size&(size-1) != 0 {
		size = nextPowerOf2Job(size)
	}
	return &JobQueue{
		jobs: make([]Job, size),
		mask: size - 1,
	}
}

func (jq *JobQueue) Enqueue(job *Job) bool {
	nextHead := (jq.head + 1) & jq.mask
	if nextHead == jq.tail {
		return false
	}
	jq.jobs[jq.head] = *job
	jq.head = nextHead
	return true
}

func (jq *JobQueue) Dequeue() (*Job, bool) {
	if jq.tail == jq.head {
		return nil, false
	}
	job := &jq.jobs[jq.tail]
	jq.tail = (jq.tail + 1) & jq.mask
	return job, true
}

func nextPowerOf2Job(n uint32) uint32 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

type Job struct {
	Type          uint8
	EventType     uint8
	PanicMode     uint8
	_             uint8
	GuildID       uint64
	TargetID      uint64
	Reason        string
	DetectionTime int64
	Data          uint64
}

const (
	JobTypeBan = iota
	JobTypeKick
	JobTypeQuarantine
	JobTypeLockdown
	JobTypeRoleRemove
)

func NewBanJob(guildID, userID uint64, reason string, eventType, panicMode uint8, detectionTime int64) *Job {
	return &Job{
		Type:          JobTypeBan,
		EventType:     eventType,
		PanicMode:     panicMode,
		GuildID:       guildID,
		TargetID:      userID,
		Reason:        reason,
		DetectionTime: detectionTime,
	}
}

func NewQuarantineJob(guildID, userID uint64, reason string) *Job {
	return &Job{
		Type:     JobTypeQuarantine,
		GuildID:  guildID,
		TargetID: userID,
		Reason:   reason,
	}
}

func NewKickJob(guildID, userID uint64, reason string) *Job {
	return &Job{
		Type:     JobTypeKick,
		GuildID:  guildID,
		TargetID: userID,
		Reason:   reason,
	}
}

func NewLockdownJob(guildID uint64, reason string) *Job {
	return &Job{
		Type:    JobTypeLockdown,
		GuildID: guildID,
		Reason:  reason,
	}
}
