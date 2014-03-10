package main

import (
	"github.com/coreos/go-etcd/etcd"
	"strings"
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
	CheckPassing     bool
}

func (n *NodeState) handleRouteConfigUpdate(name string, response *etcd.Response) {
	// TODO: update configs, use index
	route := n.Routes[name]
	if route.ConfigIndex != response.EtcdIndex {
		n.Logger.Println("New healthcheck!", response.EtcdIndex)
		n.Logger.Println(response.Node.Value)
		route.ConfigIndex = response.EtcdIndex
		n.Routes[name] = route
	}
}

func (n *NodeState) handleSubscribedRoutes(response *etcd.Response, stop chan bool) []string {
	// the subscribedRoutes key is a space separated list of route names to watch
	routes := strings.Split(response.Node.Value, " ")
	for _, route := range routes {
		// Grab config and healthcheck keys before we start the goroutine
		baseKey := "/bgp/routes/" + route
		healthcheckKey := baseKey + "/healthcheck"
		configKey := baseKey + "/config"

		response, err := n.etcd.Get(healthcheckKey, false, false)
		if err != nil {
			n.Logger.Println("Could not get healtcheck from", healthcheckKey)
		} else {
			n.handleRouteHealthcheckUpdate(route, response)
		}

		response, err = n.etcd.Get(configKey, false, false)
		if err != nil {
			n.Logger.Println("Could not get config from", configKey)
		} else {
			n.handleRouteConfigUpdate(route, response)
		}

		// Watch these keys
		go n.WatchRoute(route, stop)
	}
	return routes
}
