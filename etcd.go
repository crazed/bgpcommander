package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

func (n *NodeState) WatchAdminKey() {
	adminState, err := n.GetRelativeKey("adminState", false, false)
	if err != nil {
		fmt.Println("No adminState exists, setting to up")
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
			fmt.Println("Changing adminstate!")
			if adminState.Node.Value == "up" {
				n.AdminUp = true
			} else {
				n.AdminUp = false
			}
		}
	}(updates)

	_, err = n.etcd.Watch(n.keyPrefix+"/adminState", 0, false, updates, nil)
}

func (n *NodeState) WatchRoute(name string, stop chan bool) {
	updates := make(chan *etcd.Response)
	go func(updates chan *etcd.Response) {
		for route := range updates {
			fmt.Println(route.Node.Key, route.Node.Value)
		}
	}(updates)
	key := "/bgp/routes/" + name
	fmt.Println("Starting to watch:", key)
	n.etcd.Watch(key, 0, true, updates, stop)
	fmt.Println("Stop watching:", key)
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

	response, _ := n.etcd.Get(n.keyPrefix+"/subscribedRoutes", false, false)
	lastRoutes := n.handleSubscribedRoutes(response, stop)

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
		fmt.Println("Creating base node:", n.keyPrefix)
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
