package ha

import (
	"sync/atomic"
	"time"
)

type LeaderElection struct {
	cluster      *Cluster
	isLeader     uint32
	electionTerm uint64
	votes        map[string]bool
}

func NewLeaderElection(cluster *Cluster) *LeaderElection {
	return &LeaderElection{
		cluster: cluster,
		votes:   make(map[string]bool),
	}
}

func (le *LeaderElection) StartElection() {
	atomic.AddUint64(&le.electionTerm, 1)

	le.votes = make(map[string]bool)
	le.votes[le.cluster.localNodeID] = true

	time.Sleep(100 * time.Millisecond)

	if len(le.votes) > len(le.cluster.nodes)/2 {
		le.BecomeLeader()
	}
}

func (le *LeaderElection) BecomeLeader() {
	atomic.StoreUint32(&le.isLeader, 1)
	le.cluster.SetLeader(le.cluster.localNodeID)
}

func (le *LeaderElection) StepDown() {
	atomic.StoreUint32(&le.isLeader, 0)
}

func (le *LeaderElection) IsLeader() bool {
	return atomic.LoadUint32(&le.isLeader) == 1
}

func (le *LeaderElection) GetTerm() uint64 {
	return atomic.LoadUint64(&le.electionTerm)
}
