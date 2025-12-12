package dispatcher

import (
	"fmt"
	"runtime"

	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/internal/decision"
	"go-antinuke-2.0/internal/notifier"
	"go-antinuke-2.0/internal/state"
	"go-antinuke-2.0/pkg/util"
)

type RESTWorker struct {
	jobQueue    *decision.JobQueue
	httpPool    *HTTPPool
	rateLimiter *RateLimitMonitor
	banExecutor *BanRequestExecutor
	running     bool
	workerID    int
	cpuCore     int
}

func NewRESTWorker(jobQueue *decision.JobQueue, httpPool *HTTPPool, rateLimiter *RateLimitMonitor, workerID, cpuCore int) *RESTWorker {
	return &RESTWorker{
		jobQueue:    jobQueue,
		httpPool:    httpPool,
		rateLimiter: rateLimiter,
		banExecutor: NewBanRequestExecutor(httpPool, rateLimiter),
		workerID:    workerID,
		cpuCore:     cpuCore,
		running:     false,
	}
}

func (rw *RESTWorker) Start() {
	if rw.cpuCore > 0 {
		runtime.LockOSThread()
	}
	rw.running = true
	rw.runLoop()
}

func (rw *RESTWorker) runLoop() {
	for rw.running {
		job, ok := rw.jobQueue.Dequeue()
		if !ok {
			runtime.Gosched()
			continue
		}

		rw.executeJob(job)
	}
}

func (rw *RESTWorker) executeJob(job *decision.Job) {
	switch job.Type {
	case decision.JobTypeBan:
		banTime, err := rw.banExecutor.ExecuteBan(job.GuildID, job.TargetID, job.Reason)
		if err == nil {
			go rw.sendLogAfterBan(job, banTime)
		} else {
			// Ban failed, unmark actor so we can try again or process new events
			rw.handleBanFailure(job.TargetID)
		}
	case decision.JobTypeKick:
		rw.banExecutor.ExecuteKick(job.GuildID, job.TargetID, job.Reason)
	}
}

func (rw *RESTWorker) handleBanFailure(actorID uint64) {
	actorMap := state.GetActorIDMap()
	actorIndex := actorMap.GetIndex(actorID)
	if actorIndex != 0 {
		as := state.GetActorState()
		as.SetBanned(actorIndex, false)
		fmt.Printf("[DISPATCHER] Ban failed for actor %d, cleared banned state\n", actorID)
	}
}

func (rw *RESTWorker) sendLogAfterBan(job *decision.Job, banTimeUS int64) {
	guildIDStr := util.Uint64ToString(job.GuildID)
	actorIDStr := util.Uint64ToString(job.TargetID)

	db := database.GetDB()
	if db == nil {
		return
	}

	config, err := db.GetGuildConfig(guildIDStr)
	if err != nil || config.LogChannelID == "" {
		return
	}

	emoji := "🔨"
	eventName := rw.getEventName(job.EventType)
	if job.PanicMode == 1 {
		emoji = "🚨"
	}

	detectionUS := job.DetectionTime / 1000
	if detectionUS == 0 && job.DetectionTime > 0 {
		detectionUS = 1
	}
	fmt.Printf("[DEBUG] Detection: %d ns -> %d µs | Ban: %d µs\n", job.DetectionTime, detectionUS, banTimeUS)
	notifier.SendEventLogWithBanTime(config.LogChannelID, emoji, eventName, actorIDStr, job.Reason, detectionUS, banTimeUS)
}

func (rw *RESTWorker) getEventName(eventType uint8) string {
	switch eventType {
	case 1:
		return "Mass Ban Attack"
	case 10:
		return "Channel Create Spam"
	case 11:
		return "Channel Delete Attack"
	case 18:
		return "Role Create Spam"
	case 19:
		return "Role Delete Attack"
	case 31:
		return "Webhook Spam"
	case 32:
		return "Permission Escalation"
	default:
		return "Malicious Activity"
	}
}

func (rw *RESTWorker) Stop() {
	rw.running = false
}
