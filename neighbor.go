package main

type Neighbor struct {
	State     string
	IPAddress string
}

func (n *NodeState) SetNeighbor(ip string, state string) {
	neighbor, found := n.Neighbors[ip]
	if !found {
		neighbor = *new(Neighbor)
	}
	neighbor.IPAddress = ip
	neighbor.State = state
	_, err := n.SetRelativeKey("neighbors/"+ip+"/state", state)
	if err != nil {
		n.Logger.Println("Failed to update etcd for neighbor!", err)
	}
	n.Neighbors[ip] = neighbor
}
