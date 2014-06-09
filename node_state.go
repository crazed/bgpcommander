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

	etcd                  *etcd.Client
	keyPrefix             string
	healthcheckScriptPath string
}

func NewNodeState(hostname string, etcd *etcd.Client, deleteExisting bool, healthcheckScriptPath string) *NodeState {
	state := new(NodeState)
	state.Neighbors = make(map[string]Neighbor)
	state.Routes = make(map[string]Route)
	state.etcd = etcd
	state.keyPrefix = fmt.Sprintf("/bgp/node/%s", hostname)
	state.RoutesKey = fmt.Sprintf("%s/%s", state.keyPrefix, "routes")
	state.NeighborsKey = fmt.Sprintf("%s/%s", state.keyPrefix, "neighbors")
	state.AdminUp = true

	// Make sure we gracefully handle failures
	etcd.CheckRetry = state.HandleEtcdFailure

	log := log.New(os.Stderr, "[bgpcommander] ", log.LstdFlags)
	state.Logger = log

	state.healthcheckScriptPath = healthcheckScriptPath
	if err := os.MkdirAll(state.healthcheckScriptPath, 0755); err != nil {
		state.Logger.Println(err)
		panic("Failed to creath healthceck scrip path!")
	}

	if deleteExisting {
		state.etcd.Delete(state.keyPrefix+"/neighbors", true)
		state.etcd.Delete(state.keyPrefix+"/routes", true)
	}

	return state
}

func (n *NodeState) GetKeyPrefix() string {
	return n.keyPrefix
}

func (n *NodeState) Shutdown() {
	n.Logger.Println("Shutting down!")
	n.RemoveState()
}
