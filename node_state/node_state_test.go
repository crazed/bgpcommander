package node_state

import "testing"
import "github.com/coreos/go-etcd/etcd"

// This expects an etcd server is running on localhost and will fail otherwise
func TestNewNodeState(t *testing.T) {
	t.Parallel()
	const keyPrefixOutput = "/bgp/node/test"
	const routesKeyOutput = "/bgp/node/test/routes"
	const neighborsKeyOutput = "/bgp/node/test/neighbors"

	client := etcd.NewClient(nil)
	state := NewNodeState("test", client, true, "/tmp/scripts")

	// Test that we have all the defaults as expected
	if state.GetKeyPrefix() != keyPrefixOutput {
		t.Errorf("state.GetKeyPrefix() = %v, want %v", state.GetKeyPrefix(), keyPrefixOutput)
	}
	if state.RoutesKey != routesKeyOutput {
		t.Errorf("state.RoutesKey = %v, want %v", state.RoutesKey, routesKeyOutput)
	}
	if state.NeighborsKey != neighborsKeyOutput {
		t.Errorf("state.NeighborsKey = %v, want %v", state.NeighborsKey, neighborsKeyOutput)
	}
	if state.AdminUp != true {
		t.Errorf("state.AdminUp = %v, want %v", state.AdminUp, true)
	}
	if state.GetHealthcheckScriptPath() != "/tmp/scripts" {
		t.Errorf("state.GetHealthcheckScriptPath() = %v, want %v", state.GetHealthcheckScriptPath(), "/tmp/scripts")
	}
}
