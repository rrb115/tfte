package scoring

import (
	"testing"

	"github.com/rrb115/tfte/proto/gen/tfte"
)

func TestScoreEdge_Deterministic(t *testing.T) {
	cfg := DefaultConfig
	failureTs := int64(100000) // 100s in milliseconds

	// Scenario 1: RPC Error close to failure
	events := []*tfte.Event{
		{Id: "1", Ts: 99000, Type: tfte.EventType_RPC_ERROR}, // 1s before
	}

	evidence := ScoreEdge(cfg, failureTs, events)

	// Expected: Base (0.1) + RPC (0.4) + Proximity (~0.3 * (29/30)) = ~0.79
	// delta = 1s. W = 30s. (30-1)/30 = 0.966. 0.966 * 0.3 = 0.29
	// Total approx 0.1 + 0.4 + 0.29 = 0.79

	if evidence.Score < 0.7 || evidence.Score > 0.9 {
		t.Errorf("Expected score around 0.8, got %f", evidence.Score)
	}

	if evidence.RpcBonus != cfg.RpcErrorBonus {
		t.Errorf("Expected RPC bonus %f, got %f", cfg.RpcErrorBonus, evidence.RpcBonus)
	}
}
