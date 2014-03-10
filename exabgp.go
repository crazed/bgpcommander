package main

import (
	"encoding/json"
)

func (n *NodeState) WithdrawRoute(name string) {
	route := n.Routes[name]
	n.Logger.Println("Removing:", route)
	// TODO: write withdraw command to STDOUT
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
