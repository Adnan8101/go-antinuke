package ha

type ClusterNode struct {
	ID       string
	Address  string
	IsLeader bool
	IsAlive  bool
}

type Cluster struct {
	nodes       map[string]*ClusterNode
	localNodeID string
}

func NewCluster(localNodeID string) *Cluster {
	return &Cluster{
		nodes:       make(map[string]*ClusterNode),
		localNodeID: localNodeID,
	}
}

func (c *Cluster) AddNode(id, address string) {
	c.nodes[id] = &ClusterNode{
		ID:       id,
		Address:  address,
		IsLeader: false,
		IsAlive:  true,
	}
}

func (c *Cluster) RemoveNode(id string) {
	delete(c.nodes, id)
}

func (c *Cluster) GetLeader() *ClusterNode {
	for _, node := range c.nodes {
		if node.IsLeader {
			return node
		}
	}
	return nil
}

func (c *Cluster) IsLeader() bool {
	node := c.nodes[c.localNodeID]
	return node != nil && node.IsLeader
}

func (c *Cluster) SetLeader(nodeID string) {
	for id, node := range c.nodes {
		node.IsLeader = (id == nodeID)
	}
}

func (c *Cluster) GetNodes() []*ClusterNode {
	nodes := make([]*ClusterNode, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}
