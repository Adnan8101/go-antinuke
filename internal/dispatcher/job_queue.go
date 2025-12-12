package dispatcher

type JobPriority uint8

const (
	PriorityLow JobPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

type PrioritizedJob struct {
	Priority  JobPriority
	GuildID   uint64
	TargetID  uint64
	Action    string
	Timestamp int64
}

type PriorityQueue struct {
	critical []*PrioritizedJob
	high     []*PrioritizedJob
	normal   []*PrioritizedJob
	low      []*PrioritizedJob
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		critical: make([]*PrioritizedJob, 0, 256),
		high:     make([]*PrioritizedJob, 0, 512),
		normal:   make([]*PrioritizedJob, 0, 1024),
		low:      make([]*PrioritizedJob, 0, 2048),
	}
}

func (pq *PriorityQueue) Enqueue(job *PrioritizedJob) {
	switch job.Priority {
	case PriorityCritical:
		pq.critical = append(pq.critical, job)
	case PriorityHigh:
		pq.high = append(pq.high, job)
	case PriorityNormal:
		pq.normal = append(pq.normal, job)
	case PriorityLow:
		pq.low = append(pq.low, job)
	}
}

func (pq *PriorityQueue) Dequeue() (*PrioritizedJob, bool) {
	if len(pq.critical) > 0 {
		job := pq.critical[0]
		pq.critical = pq.critical[1:]
		return job, true
	}

	if len(pq.high) > 0 {
		job := pq.high[0]
		pq.high = pq.high[1:]
		return job, true
	}

	if len(pq.normal) > 0 {
		job := pq.normal[0]
		pq.normal = pq.normal[1:]
		return job, true
	}

	if len(pq.low) > 0 {
		job := pq.low[0]
		pq.low = pq.low[1:]
		return job, true
	}

	return nil, false
}

func (pq *PriorityQueue) Size() int {
	return len(pq.critical) + len(pq.high) + len(pq.normal) + len(pq.low)
}
