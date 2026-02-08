package engine

import (
	"encoding/json"
	"testing"

	"github.com/rrb115/tfte/proto/gen/tfte"
)

func TestReconstructGraph(t *testing.T) {
	// 1. Setup events
	// A -> B (RPC)
	// B goes DOWN

	events := []*tfte.Event{
		{
			Id: "1", Ts: 100, Type: tfte.EventType_RPC_CALL, Service: "A",
			Payload: jsonPayload(t, map[string]string{"dest_service": "B"}),
		},
		{
			Id: "2", Ts: 200, Type: tfte.EventType_HEALTH_CHANGE, Service: "B",
			Payload: jsonPayload(t, map[string]string{"new_status": "DOWN"}),
		},
	}

	// 2. Reconstruct at T=300
	snapshot, _ := ReconstructGraph(300, events)

	// 3. Verify
	if len(snapshot.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(snapshot.Nodes))
	}

	// Verify Node B is DOWN (2)
	var nodeB *tfte.Node
	for _, n := range snapshot.Nodes {
		if n.Id == "B" {
			nodeB = n
			break
		}
	}
	if nodeB == nil {
		t.Fatal("Node B not found")
	}
	if nodeB.HealthStatus != 2 {
		t.Errorf("Expected Node B status 2 (DOWN), got %d", nodeB.HealthStatus)
	}

	// Verify Edge A->B
	if len(snapshot.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(snapshot.Edges))
	}
	edge := snapshot.Edges[0]
	if edge.Source != "A" || edge.Target != "B" {
		t.Errorf("Expected A->B edge, got %s->%s", edge.Source, edge.Target)
	}
}

func jsonPayload(t *testing.T, data interface{}) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
