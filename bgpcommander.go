package main

import (
	"bufio"
	"flag"
	"github.com/coreos/go-etcd/etcd"
	"os"
	"strings"
)

func readStdin(state *NodeState) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		state.ProcessExaBGPOutput(scanner.Bytes())
	}
}

type machines []string

func (m *machines) String() string {
	return strings.Join(*m, ", ")
}

func (m *machines) Set(value string) error {
	for _, machine := range strings.Split(value, ",") {
		*m = append(*m, strings.TrimSpace(machine))
	}
	return nil
}

func main() {
	cluster := machines{"http://127.0.0.1:4001"}
	healthcheckScriptPath := "/tmp"
	hostname, _ := os.Hostname()

	flag.Var(&cluster, "c", "Comma separated list of etcd cluster members")
	flag.StringVar(&healthcheckScriptPath, "p", healthcheckScriptPath, "Path to store healthcheck scripts")
	flag.StringVar(&hostname, "n", hostname, "Override node's name")
	flag.Parse()

	client := etcd.NewClient(cluster)

	state := NewNodeState(hostname, client, true, healthcheckScriptPath)
	state.Logger.Println("Using etcd servers:", cluster.String())
	state.Logger.Println("Using node name:", hostname)
	state.Logger.Println("Using healthcheck script path:", healthcheckScriptPath)

	go readStdin(state)
	state.WatchKeys()

	// goroutines are doing the work, so block here
	select {}
}
