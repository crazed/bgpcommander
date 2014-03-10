package main

import (
	"encoding/json"
	"fmt"
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
