package main

import "testing"
import "os"
import "github.com/coreos/go-etcd/etcd"
import "log"
import "io/ioutil"

func TestRouteUpdates(t *testing.T) {
	t.Parallel()
	scriptsDir, _ := ioutil.TempDir("", "scripts")
	client := etcd.NewClient(nil)
	state := NewNodeState("test", client, true, scriptsDir)
	devNull, _ := os.Open(os.DevNull)
	nullLogger := log.New(devNull, "", log.LstdFlags)
	state.Logger = nullLogger

	// Make sure we can handle signals safely
	signals := make(chan os.Signal, 1)
	go handleSignal(signals, state)

	// Every time there's an event, drop it in a channel
	eventStream := make(chan *Event)
	state.EventHandler = func(event *Event) {
		eventStream <- event
	}

	// Helper function to wait for supplied event name
	eventWait := func(name string) {
		for event := range eventStream {
			if event.Type == name {
				return
			}
		}
	}

	// Create a new route
	const routeHealthCheck = "#!/bin/sh\necho helloworld!\n"
	const routeConfig = "community [ 1111:2 ]"
	const routePrefix = "127.0.0.1/32"
	const routeName = "test-route"
	const routeKey = "/bgp/routes/test-route"
	const subscribedRoutesKey = "/bgp/node/test/subscribedRoutes"

	// Provide a method to clean things up
	cleanUpTest := func() {
		state.EventHandler = nil
		close(eventStream)
		client.Delete(routeKey, true)
		client.Delete(subscribedRoutesKey, false)
		_ = os.RemoveAll(scriptsDir)
		state.Shutdown()
	}
	defer cleanUpTest()

	// Start watching the important keys
	state.WatchKeys()

	var err error
	handleErr := func(key string, value string, err error) {
		t.Errorf("Received an error while setting '%v' => '%v': %v", key, value, err)
	}

	_, err = client.CreateDir(routeKey, 0)
	if err != nil {
		handleErr(routeKey, "dir", err)
	}

	var healthcheckKey = routeKey + "/healthcheck"
	_, err = client.Set(healthcheckKey, routeHealthCheck, 0)
	if err != nil {
		handleErr(healthcheckKey, routeHealthCheck, err)
	}

	var prefixKey = routeKey + "/prefix"
	_, err = client.Set(prefixKey, routePrefix, 0)
	if err != nil {
		handleErr(prefixKey, routePrefix, err)
	}

	var configKey = routeKey + "/config"
	_, err = client.Set(configKey, routeConfig, 0)
	if err != nil {
		handleErr(configKey, routeConfig, err)
	}

	_, err = client.Set(subscribedRoutesKey, routeName, 0)
	if err != nil {
		handleErr(subscribedRoutesKey, routeName, err)
	}

	// Wait until we see route_subscription_update to know we handled everything
	eventWait("route_subscription_update")

	// Test that our new route has been populated internally
	route := state.GetRoute(routeName)
	if route.Name != routeName {
		t.Errorf("route.Name = %v, want %v", route.Name, routeName)
	}
	if route.Prefix != routePrefix {
		t.Errorf("route.Prefix = %v, want %v", route.Prefix, routePrefix)
	}
	if route.Config != routeConfig {
		t.Errorf("route.Config = %v, want %v", route.Config, routeConfig)
	}
	// Test that the route healthcheck has the contents we expect
	contents, _ := ioutil.ReadFile(scriptsDir + "/" + routeName)
	if routeHealthCheck != string(contents) {
		t.Errorf("Healthcheck contents not correct!")
	}
}
