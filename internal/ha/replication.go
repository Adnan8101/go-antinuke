package ha

type ReplicationManager struct {
	cluster *Cluster
}

func NewReplicationManager(cluster *Cluster) *ReplicationManager {
	return &ReplicationManager{
		cluster: cluster,
	}
}

func (rm *ReplicationManager) ReplicateState(state interface{}) error {
	return nil
}

func (rm *ReplicationManager) SyncWithPeer(nodeID string) error {
	return nil
}

func (rm *ReplicationManager) GetReplicationLag() int {
	return 0
}
