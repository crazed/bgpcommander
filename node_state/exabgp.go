package node_state

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"net/http"
	"time"
)

func (n *NodeState) WithdrawRoute(route *Route) {
	n.SetRelativeKey("/routes/"+route.Name+"/state", "withdraw")
	fmt.Printf("withdraw route %s %s\n", route.Prefix, route.Config)
}

func (n *NodeState) AnnounceRoute(route *Route) {
	n.SetRelativeKey("/routes/"+route.Name+"/state", "announce")
	fmt.Printf("announce route %s %s\n", route.Prefix, route.Config)
}

func (n *NodeState) ProcessExaBGPOutput(buf []byte) {
	s := string(buf[:])
	n.newEvent("info", "exabgp_data", s)

	var f interface{}
	err := json.Unmarshal(buf, &f)
	if err != nil {
		n.newEvent("error", "exabgp_json_error", "Failed to parse JSON from Exabgp: "+err.Error())
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
	n.newEvent("error", "etcd_connection_failure", "Lost connection to etcd! Waiting 5 seconds...")
	time.Sleep(5000 * time.Millisecond)
	return nil
}
