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

	healthcheckScriptPath := "/tmp"
	state := NewNodeState(hostname, client, true, healthcheckScriptPath)

	go readStdin(state)
	state.WatchKeys()

	// goroutines are doing the work, so block here
	select {}
}
