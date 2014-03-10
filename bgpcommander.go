package main

import (
	"bufio"
	"github.com/coreos/go-etcd/etcd"
	"os"
)

func readStdin(state *NodeState) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		state.ProcessExaBGPOutput(scanner.Bytes())
	}
}

func main() {
	machines := []string{"http://127.0.0.1:4001"}
	client := etcd.NewClient(machines)
	hostname, _ := os.Hostname()
	state := NewNodeState(hostname, client, true)

	routeUpdates := make(chan *etcd.Response)
	stopWatches := make(chan bool)

	state.WatchKeys()

	go state.WatchRouteUpdates(routeUpdates)
	go readStdin(state)

	state.Logger.Println("Watching:", state.RoutesKey)
	_, err := client.Watch(state.RoutesKey, 0, true, routeUpdates, stopWatches)

	if err != nil {
		state.Logger.Println(err)
	}
}
