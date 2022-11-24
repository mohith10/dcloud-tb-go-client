package tbclient

const path = "%s/topologies"
const pathOne = path + "/%s"

func (c *Client) getTopologyService() *service[Topology, topologyCollection] {
	return &service[Topology, topologyCollection]{
		readService[Topology, topologyCollection]{
			client:       c,
			resourcePath: "/topologies",
		},
	}
}

func (c *Client) GetAllTopologies() ([]Topology, error) {
	topologies, err := c.getTopologyService().getAll()
	return topologies, err
}

func (c *Client) GetTopology(uid string) (*Topology, error) {
	return c.getTopologyService().getOne(uid)
}

func (c *Client) CreateTopology(topology Topology) (*Topology, error) {
	return c.getTopologyService().create(topology)
}

func (c *Client) UpdateTopology(topology Topology) (*Topology, error) {
	return c.getTopologyService().update(topology.Uid, topology)
}

func (c *Client) DeleteTopology(topologyUid string) error {
	return c.getTopologyService().delete(topologyUid)
}
