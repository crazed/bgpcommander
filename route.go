package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
)

type Route struct {
	Name             string
	AdminUp          bool
	Prefix           string
	Communities      []string
	LocalPref        int
	MED              int
	Healthcheck      string
	HealthcheckIndex uint64
	ConfigIndex      uint64
}

func (n *NodeState) WatchRouteUpdates(watch chan *etcd.Response) {
	for response := range watch {
		fmt.Println(response.Node.Key, response.Node.Value)
	}
}
