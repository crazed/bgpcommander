package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"os"
)

type NodeState struct {
	Neighbors    map[string]Neighbor
	Routes       map[string]Route
	RoutesKey    string
	NeighborsKey string
	AdminUp      bool
	Logger       *log.Logger

	etcd      *etcd.Client
	keyPrefix string
}

func NewNodeState(hostname string, etcd *etcd.Client, deleteExisting bool) *NodeState {
	state := new(NodeState)
	state.Neighbors = make(map[string]Neighbor)
	state.Routes = make(map[string]Route)
	state.etcd = etcd
	state.keyPrefix = fmt.Sprintf("/bgp/node/%s", hostname)
	state.RoutesKey = fmt.Sprintf("%s/%s", state.keyPrefix, "routes")
	state.NeighborsKey = fmt.Sprintf("%s/%s", state.keyPrefix, "neighbors")
	state.AdminUp = true

	log := log.New(os.Stderr, "[bgpcommander] ", log.LstdFlags)
	state.Logger = log

	if deleteExisting {
		state.etcd.Delete(state.keyPrefix+"/neighbors", true)
	}

	return state
}
