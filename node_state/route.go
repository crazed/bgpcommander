package node_state

import (
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

type Route struct {
	Name             string
	AdminUp          bool
	PrefixIndex      uint64
	Prefix           string
	Healthcheck      string
	HealthcheckIndex uint64
	ConfigIndex      uint64
	Config           string
	CheckPassing     bool
	StopCheck        chan bool
	CheckRunning     bool
}

func (n *NodeState) GetRoute(name string) *Route {
	route, found := n.Routes[name]
	if !found {
		route = *new(Route)
		route.Name = name
		route.StopCheck = make(chan bool)
		route.CheckRunning = false
		route.HealthcheckIndex = 0
	}
	n.Routes[name] = route
	return &route
}

func (n *NodeState) UpdateRoute(route *Route) {
	n.Routes[route.Name] = *route
}

func (n *NodeState) handleRoutePrefixUpdate(route *Route, response *etcd.Response) {
	if route.PrefixIndex != response.Node.ModifiedIndex {
		route.Prefix = response.Node.Value
		route.PrefixIndex = response.Node.ModifiedIndex
		n.newEvent("info", "route_prefix_update", "new prefix for "+route.Name+" is "+route.Prefix)
		n.UpdateRoute(route)
		// Announce the route again if we change prefixes
		if route.CheckPassing && n.AdminUp {
			n.AnnounceRoute(route)
		}
	}
}

func (n *NodeState) handleRouteConfigUpdate(route *Route, response *etcd.Response) {
	if route.ConfigIndex != response.Node.ModifiedIndex {
		route.Config = response.Node.Value
		route.ConfigIndex = response.Node.ModifiedIndex
		n.newEvent("info", "route_config_update", "new config for "+route.Name+" is '"+route.Config+"'")
		n.UpdateRoute(route)
		// Announce the route again if we're good to handle any config changes
		if route.CheckPassing && n.AdminUp {
			n.AnnounceRoute(route)
		}
	}
}

func (n *NodeState) handleSubscribedRoutes(response *etcd.Response, stop chan bool) []string {
	// the subscribedRoutes key is a space separated list of route names to watch
	routes := strings.Split(response.Node.Value, " ")
	for _, routeName := range routes {
		route := n.GetRoute(routeName)
		// Grab config, healthcheck, and prefix keys before we start the goroutine
		baseKey := "/bgp/routes/" + route.Name
		healthcheckKey := baseKey + "/healthcheck"
		configKey := baseKey + "/config"
		prefixKey := baseKey + "/prefix"

		response, err := n.etcd.Get(healthcheckKey, false, false)
		if err != nil {
			n.newEvent("error", "route_healthcheck_error", "Could not get healthcheck from '"+healthcheckKey+"'")
		} else {
			n.handleRouteHealthcheckUpdate(route, response)
		}

		response, err = n.etcd.Get(configKey, false, false)
		if err != nil {
			n.newEvent("error", "route_config_error", "Could not get config from '"+configKey+"'")
		} else {
			n.handleRouteConfigUpdate(route, response)
		}

		response, err = n.etcd.Get(prefixKey, false, false)
		if err != nil {
			n.newEvent("error", "route_prefix_error", "Could not get config from '"+prefixKey+"'")
		} else {
			n.handleRoutePrefixUpdate(route, response)
		}

		// Watch these keys
		go n.WatchRoute(route, stop)
	}
	n.newEvent("info", "route_subscription_update", "new subscribed routes are '"+response.Node.Value+"'")
	return routes
}

func (n *NodeState) WithdrawAllRoutes() {
	for key := range n.Routes {
		route := n.GetRoute(key)
		n.WithdrawRoute(route)
	}
}
