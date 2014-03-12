package main

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"net/http"
	"time"
)

func (n *NodeState) WithdrawRoute(route *Route) {
	fmt.Println("withdraw", route.Config)
}

func (n *NodeState) AnnounceRoute(route *Route) {
	fmt.Println("announce", route.Config)
}

func (n *NodeState) ProcessExaBGPOutput(buf []byte) {
	s := string(buf[:])
	n.Logger.Println("ExaBGP Sent:", s)

	var f interface{}
	err := json.Unmarshal(buf, &f)
	if err != nil {
		n.Logger.Println(err)
		return
	}

	data := f.(map[string]interface{})
	for k, v := range data {
		switch k {
		case "neighbor":
			neighData := v.(map[string]interface{})
			ip := neighData["ip"].(string)
			state := neighData["state"].(string)
			n.SetNeighbor(ip, state)
		}
	}
}

func (n *NodeState) HandleEtcdFailure(cluster *etcd.Cluster, reqs []http.Request, resps []http.Response, err error) error {
	n.Logger.Println("ERROR: Lost connection to etcd! Waiting 5 seconds...")
	time.Sleep(5000 * time.Millisecond)
	return nil
}
