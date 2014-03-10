package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

func (n *NodeState) WatchAdminKey() {
	adminState, err := n.GetRelativeKey("adminState", false, false)
	if err != nil {
		n.Logger.Println("No adminState exists, setting to up")
		n.SetRelativeKey("adminState", "up")
	} else {
		if adminState.Node.Value == "up" {
			n.AdminUp = true
		} else {
			n.AdminUp = false
		}
	}

	updates := make(chan *etcd.Response)
	go func(updates chan *etcd.Response) {
		for adminState := range updates {
			n.Logger.Println("Changing adminstate!")
			if adminState.Node.Value == "up" {
				n.AdminUp = true
			} else {
				n.AdminUp = false
			}
		}
	}(updates)

	_, err = n.etcd.Watch(n.keyPrefix+"/adminState", 0, false, updates, nil)
}

func (n *NodeState) handleRouteHealthcheckUpdate(name string, response *etcd.Response) {
	// TODO: update healthchecks, use index
	route := n.Routes[name]
	if route.HealthcheckIndex != response.EtcdIndex {
		n.Logger.Println("New healthcheck!", response.EtcdIndex)
		n.Logger.Println(response.Node.Value)
		route.HealthcheckIndex = response.EtcdIndex
		n.Routes[name] = route
	}
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

func (n *NodeState) WatchRoute(name string, stop chan bool) {
	updates := make(chan *etcd.Response)
	key := "/bgp/routes/" + name
	go func(updates chan *etcd.Response) {
		for route := range updates {
			switch route.Node.Key {
			case key + "/healthcheck":
				n.handleRouteHealthcheckUpdate(name, route)
			case key + "/config":
				n.handleRouteConfigUpdate(name, route)
			default:
				n.Logger.Println("Unknown key:", route.Node.Key)
			}
		}
	}(updates)
	n.Logger.Println("Starting to watch:", key)
	n.etcd.Watch(key, 0, true, updates, stop)
	n.Logger.Println("Stop watching:", key)
}

func (n *NodeState) handleSubscribedRoutes(response *etcd.Response, stop chan bool) []string {
	// the subscribedRoutes key is a space separated list of route names to watch
	routes := strings.Split(response.Node.Value, " ")
	for _, route := range routes {
		go n.WatchRoute(route, stop)
	}
	return routes
}

func (n *NodeState) WatchSubscribedRoutes() {
	updates := make(chan *etcd.Response)
	stop := make(chan bool)

	var lastRoutes []string
	response, err := n.etcd.Get(n.keyPrefix+"/subscribedRoutes", false, false)
	if err != nil {
		n.Logger.Println("Could not get subscribedRoutes:", err)
	} else {
		lastRoutes = n.handleSubscribedRoutes(response, stop)
	}

	go func(updates chan *etcd.Response, stop chan bool, lastRoutes []string) {
		for subscribedRoutes := range updates {
			// Start off by killing all previous goroutines that were launched
			// to watch route keys
			for _ = range lastRoutes {
				stop <- true
			}
			newRoutes := n.handleSubscribedRoutes(subscribedRoutes, stop)
			// Find all routes that we are no longer subscribed to, and remove
			// them from our internal structure
			for _, a := range lastRoutes {
				lastInNew := false
				for _, b := range newRoutes {
					if a == b {
						lastInNew = true
						break
					}
				}
				if !lastInNew {
					delete(n.Routes, a)
					n.WithdrawRoute(a)
				}
			}
			lastRoutes = newRoutes
		}
	}(updates, stop, lastRoutes)

	n.etcd.Watch(n.keyPrefix+"/subscribedRoutes", 0, false, updates, nil)
}

func (n *NodeState) WatchKeys() {
	// Here we grab the subscribedRoutes key for this node, then start watching for
	// updates to each route. On top of that we start watching for node level changes.
	_, err := n.etcd.Get(n.keyPrefix, false, false)
	if err != nil {
		n.Logger.Println("Creating base node:", n.keyPrefix)
		n.etcd.CreateDir(n.keyPrefix, 0)
	}

	go n.WatchAdminKey()
	go n.WatchSubscribedRoutes()
}

func (n *NodeState) SetRelativeKey(key string, value string) (*etcd.Response, error) {
	fullKey := fmt.Sprintf("%s/%s", n.keyPrefix, key)
	return n.etcd.Set(fullKey, value, 0)
}

func (n *NodeState) GetRelativeKey(key string, sort, recursive bool) (*etcd.Response, error) {
	fullKey := fmt.Sprintf("%s/%s", n.keyPrefix, key)
	return n.etcd.Get(fullKey, sort, recursive)
}
