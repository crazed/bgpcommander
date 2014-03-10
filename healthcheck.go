package main

import (
	"github.com/coreos/go-etcd/etcd"
	"io/ioutil"
	"os/exec"
	"path"
	"time"
)

func (n *NodeState) runCheck(route *Route) {
	for {
		select {
		case <-route.StopCheck:
			route.CheckRunning = false
			route.CheckPassing = false
			n.UpdateRoute(route)
			return
		default:
			route.CheckRunning = true
			script := n.GetFilePathForRouteCheck(route.Name)
			out, err := exec.Command(script).Output()
			if err != nil {
				n.Logger.Println("ERROR: script failed:", script, string(out))
				if route.CheckPassing != false {
					route.CheckPassing = false
					n.WithdrawRoute(route)
				}
			} else {
				if route.CheckPassing != true {
					route.CheckPassing = true
					n.AnnounceRoute(route)
				}
			}
			n.UpdateRoute(route)
			// TODO: make sleep configurable
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func (n *NodeState) handleRouteHealthcheckUpdate(route *Route, response *etcd.Response) {
	if route.HealthcheckIndex != response.Node.ModifiedIndex {
		n.Logger.Println("New healthcheck!", response.Node.ModifiedIndex)
		n.createHealthcheckScript(route.Name, response.Node.Value)

		if route.CheckRunning {
			route.StopCheck <- true
		}
		if n.AdminUp {
			go n.runCheck(route)
		}

		route.HealthcheckIndex = response.Node.ModifiedIndex
		n.UpdateRoute(route)
	}
}

func (n *NodeState) GetFilePathForRouteCheck(name string) string {
	return path.Join(n.healthcheckScriptPath, name)
}

func (n *NodeState) createHealthcheckScript(name string, scriptContent string) {
	filePath := n.GetFilePathForRouteCheck(name)
	err := ioutil.WriteFile(filePath, []byte(scriptContent), 0755)
	if err != nil {
		n.Logger.Println("ERROR: failed to write healthcheck", filePath, err)
		route := n.Routes[name]
		route.CheckPassing = false
		n.Routes[name] = route
	}
}

func (n *NodeState) StartAllChecks() {
	for key := range n.Routes {
		route := n.GetRoute(key)
		if !route.CheckRunning {
			go n.runCheck(route)
		}
	}
}

func (n *NodeState) StopAllChecks() {
	for key := range n.Routes {
		route := n.GetRoute(key)
		if route.CheckRunning {
			route.StopCheck <- true
		}
	}
}
