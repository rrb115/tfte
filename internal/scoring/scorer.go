package scoring

import (
	"math"

	"github.com/rrb115/tfte/proto/gen/tfte"
)

// Config holds tunable parameters for scoring.
type Config struct {
	// WindowSize is in milliseconds to match event timestamps stored in Badger.
	WindowSize           int64
	BaseScore            float64
	TraceBonus           float64
	RpcErrorBonus        float64
	TimeProximityBonus   float64
	AmplificationPenalty float64
}

var DefaultConfig = Config{
	WindowSize:           30000, // 30s in milliseconds
	BaseScore:            0.1,
	TraceBonus:           0.7,
	RpcErrorBonus:        0.4,
	TimeProximityBonus:   0.3,
	AmplificationPenalty: 0.2, // applied as negative
}

// ScoreEdge calculates the causal confidence score for a dependency A -> B
// given that B failed at failureTs, and we have interaction events from A -> B.
func ScoreEdge(cfg Config, failureTs int64, interactions []*tfte.Event) *tfte.EdgeEvidence {
	evidence := &tfte.EdgeEvidence{
		BaseScore:            cfg.BaseScore,
		ContributingEventIds: []string{},
	}

	total := cfg.BaseScore

	// Track max bonuses found
	hasTraceMatch := false
	hasRpcError := false
	minDelta := cfg.WindowSize // Start with max window

	for _, evt := range interactions {
		// Only consider events before the failure timestamp and inside the scoring window (all in ms)
		if evt.Ts > failureTs {
			continue
		}
		delta := failureTs - evt.Ts
		if delta > cfg.WindowSize {
			continue
		}

		evidence.ContributingEventIds = append(evidence.ContributingEventIds, evt.Id)

		if delta < minDelta {
			minDelta = delta
		}

		// Check for Trace ID match (strongest signal)
		// in Phase 3 we assume if specific trace IDs are passed in params we check them,
		// but here we are looking at aggregated edge events.
		// If ANY event has a trace ID that is also implicated in B's failure (not passed here yet),
		// we would set this. For now, let's assume if the RPC call itself has a trace_id it's a "trace capable" link
		// but "Trace Bonus" usually means "Same Request Trace found on both sides".
		// Let's implement simpler logic: if RPC_ERROR, it's a strong signal.

		if evt.Type == tfte.EventType_RPC_ERROR {
			hasRpcError = true
		}

		// If event actually shares a trace ID with the failure event (which we don't have access to in this sig yet),
		// we would set hasTraceMatch.
		// TODO: Pass target failure event to check trace correlation.
	}

	if hasTraceMatch {
		evidence.TraceBonus = cfg.TraceBonus
		total += cfg.TraceBonus
	}

	if hasRpcError {
		evidence.RpcBonus = cfg.RpcErrorBonus
		total += cfg.RpcErrorBonus
	}

	// Time Proximity Bonus: closer to failure is better
	// formula: max(0, (W - delta)/W * max_bonus)
	if len(evidence.ContributingEventIds) > 0 {
		proximityFactor := float64(cfg.WindowSize-minDelta) / float64(cfg.WindowSize)
		if proximityFactor < 0 {
			proximityFactor = 0
		}
		evidence.TimeProximityBonus = proximityFactor * cfg.TimeProximityBonus
		total += evidence.TimeProximityBonus
	}

	// Normalize/Clamp to [0, 1]
	if total > 1.0 {
		total = 1.0
	}
	if total < 0.0 {
		total = 0.0
	}

	// Deterministic rounding to 6 decimal places
	evidence.Score = math.Round(total*1000000) / 1000000

	return evidence
}
