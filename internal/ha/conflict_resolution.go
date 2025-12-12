package ha

import (
	"go-antinuke-2.0/internal/logging"
)

type ConflictResolver struct {
	cluster *Cluster
}

func NewConflictResolver(cluster *Cluster) *ConflictResolver {
	return &ConflictResolver{
		cluster: cluster,
	}
}

func (cr *ConflictResolver) ResolveSplitBrain() {
	logging.Warn("Split-brain detected, resolving...")

	leaders := cr.findMultipleLeaders()
	if len(leaders) <= 1 {
		return
	}

	chosen := cr.electWinner(leaders)

	for _, node := range leaders {
		if node.ID != chosen.ID {
			node.IsLeader = false
			logging.Info("Demoted node %s from leader", node.ID)
		}
	}

	logging.Info("Split-brain resolved, leader: %s", chosen.ID)
}

func (cr *ConflictResolver) findMultipleLeaders() []*ClusterNode {
	leaders := make([]*ClusterNode, 0)

	for _, node := range cr.cluster.nodes {
		if node.IsLeader {
			leaders = append(leaders, node)
		}
	}

	return leaders
}

func (cr *ConflictResolver) electWinner(leaders []*ClusterNode) *ClusterNode {
	if len(leaders) == 0 {
		return nil
	}

	winner := leaders[0]
	for _, node := range leaders[1:] {
		if node.ID < winner.ID {
			winner = node
		}
	}

	return winner
}

func (cr *ConflictResolver) CheckConflict() bool {
	return len(cr.findMultipleLeaders()) > 1
}
