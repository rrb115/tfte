package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rrb115/tfte/internal/scoring"
	"github.com/rrb115/tfte/internal/storage"
	"github.com/rrb115/tfte/proto/gen/tfte"
)

type Engine struct {
	store storage.Store
}

func NewEngine(store storage.Store) *Engine {
	return &Engine{store: store}
}

// GetEvents proxies to the storage layer
func (e *Engine) GetEvents(ctx context.Context, req *tfte.GetEventsRequest) (*tfte.GetEventsResponse, error) {
	events, err := e.store.GetEvents(ctx, req.StartTs, req.EndTs, req.ServiceFilter, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	return &tfte.GetEventsResponse{Events: events}, nil
}

// GetGraphAt reconstructs the graph state at a specific timestamp.
// For Phase 1/2, we will do a simple replay of valid events up to that timestamp.
// In a real production system, we would load the nearest previous snapshot and replay forward.

// GetEdgeEvidence returns the stored evidence for a specific edge.
func (e *Engine) GetEdgeEvidence(ctx context.Context, req *tfte.GetEdgeEvidenceRequest) (*tfte.EdgeEvidence, error) {
	// For Phase 3, we reconstruct on the fly or need to cache it.
	// Since GetGraphAt reconstructs fresh, we should probably do the same here or cache the last result.
	// For simplicity/correctness, let's reconstruct the graph at the requested timestamp and find the evidence.

	// We need to fetch the graph at the "associated_timestamp" provided in request
	_, evidenceMap, err := e.reconstructGraphWithEvidence(ctx, req.AssociatedTimestamp)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s|%s", req.SourceService, req.TargetService)
	if ev, ok := evidenceMap[key]; ok {
		return ev, nil
	}

	return nil, fmt.Errorf("evidence not found for edge %s->%s at ts %d", req.SourceService, req.TargetService, req.AssociatedTimestamp)
}

// GetGraphWithEvidence returns both the snapshot and the evidence map.
func (e *Engine) GetGraphWithEvidence(ctx context.Context, ts int64) (*tfte.GraphSnapshot, map[string]*tfte.EdgeEvidence, error) {
	return e.reconstructGraphWithEvidence(ctx, ts)
}

// GetGraphAt reconstructs the graph state at a specific timestamp.
func (e *Engine) GetGraphAt(ctx context.Context, ts int64) (*tfte.GraphSnapshot, error) {
	snapshot, _, err := e.reconstructGraphWithEvidence(ctx, ts)
	return snapshot, err
}

func (e *Engine) reconstructGraphWithEvidence(ctx context.Context, ts int64) (*tfte.GraphSnapshot, map[string]*tfte.EdgeEvidence, error) {
	// Window size: 1 hour in MILLISECONDS
	windowSize := int64(3600 * 1000)
	startTs := ts - windowSize
	if startTs < 0 {
		startTs = 0
	}

	// Fetch enough events to reconstruct state
	events, err := e.store.GetEvents(ctx, startTs, ts, "", 10000, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	snapshot, evidence := ReconstructGraph(ts, events)
	return snapshot, evidence, nil
}

// ReconstructGraph builds the graph snapshot AND evidence map.
func ReconstructGraph(atTs int64, events []*tfte.Event) (*tfte.GraphSnapshot, map[string]*tfte.EdgeEvidence) {
	nodes := make(map[string]*tfte.Node)
	edges := make(map[string]*tfte.Edge)
	evidenceMap := make(map[string]*tfte.EdgeEvidence)

	// Temporary storage for edge interactions to compute scores later
	// Key: "src|dst" -> list of events
	edgeInteractions := make(map[string][]*tfte.Event)

	// Sort events by timestamp to ensure determinism
	sort.Slice(events, func(i, j int) bool {
		if events[i].Ts == events[j].Ts {
			return events[i].Id < events[j].Id
		}
		return events[i].Ts < events[j].Ts
	})

	for _, evt := range events {
		if evt.Ts > atTs {
			break
		}

		ensureNode(nodes, evt.Service)

		switch evt.Type {
		case tfte.EventType_HEALTH_CHANGE:
			var hc struct {
				NewStatus string `json:"new_status"`
			}
			if err := json.Unmarshal(evt.Payload, &hc); err == nil {
				node := nodes[evt.Service]
				if hc.NewStatus == "DOWN" {
					node.HealthStatus = 2
				} else if hc.NewStatus == "DEGRADED" {
					node.HealthStatus = 1
				} else {
					node.HealthStatus = 0
				}
			}

		case tfte.EventType_RPC_CALL, tfte.EventType_RPC_ERROR:
			var rpc struct {
				DestService string `json:"dest_service"`
			}
			if err := json.Unmarshal(evt.Payload, &rpc); err == nil && rpc.DestService != "" {
				ensureNode(nodes, rpc.DestService)
				key := fmt.Sprintf("%s|%s", evt.Service, rpc.DestService)

				// Accumulate interaction
				edgeInteractions[key] = append(edgeInteractions[key], evt)

				if _, exists := edges[key]; !exists {
					edges[key] = &tfte.Edge{
						Source:           evt.Service,
						Target:           rpc.DestService,
						IsActive:         true,
						CausalConfidence: 0.1, // Will be updated by scoring
					}
				}
			}
		}
	}

	// Compute scores
	cfg := scoring.DefaultConfig
	for key, interactionEvents := range edgeInteractions {
		edge := edges[key]
		targetNode := nodes[edge.Target]

		// If target is NOT Healthy (Status > 0), we calculate score based on failure time (atTs).
		// If target IS Healthy (Status == 0), the causal confidence of a "failure propagation" is low.
		// However, for visualization, we might still want to show the "stress" score.
		// Let's reduce the base score if target is healthy.

		failureTs := atTs

		// Adjust config if target is healthy
		currentCfg := cfg
		if targetNode.HealthStatus == 0 {
			currentCfg.BaseScore = 0.01 // Very low confidence if target is healthy
		}

		ev := scoring.ScoreEdge(currentCfg, failureTs, interactionEvents)
		ev.SourceService = edge.Source
		ev.TargetService = edge.Target

		evidenceMap[key] = ev
		edge.CausalConfidence = ev.Score

		// Calculate Amplification Score for Source Node
		// Rule: If Target is DOWN/DEGRADED, and Source sent errors, increase Source's amplification score.
		if targetNode.HealthStatus > 0 {
			errorCount := 0
			for _, evt := range interactionEvents {
				if evt.Type == tfte.EventType_RPC_ERROR {
					errorCount++
				}
			}
			if errorCount > 0 {
				// Simple increment for now.
				// In real system, normalize by time window or request rate.
				sourceNode := nodes[edge.Source]
				if sourceNode != nil {
					sourceNode.AmplificationScore += float64(errorCount)
				}
			}
		}
	}

	snapshot := &tfte.GraphSnapshot{
		Timestamp: atTs,
		Nodes:     make([]*tfte.Node, 0, len(nodes)),
		Edges:     make([]*tfte.Edge, 0, len(edges)),
	}

	for _, n := range nodes {
		snapshot.Nodes = append(snapshot.Nodes, n)
	}
	for _, e := range edges {
		snapshot.Edges = append(snapshot.Edges, e)
	}

	// Sort
	sort.Slice(snapshot.Nodes, func(i, j int) bool { return snapshot.Nodes[i].Id < snapshot.Nodes[j].Id })
	sort.Slice(snapshot.Edges, func(i, j int) bool {
		if snapshot.Edges[i].Source == snapshot.Edges[j].Source {
			return snapshot.Edges[i].Target < snapshot.Edges[j].Target
		}
		return snapshot.Edges[i].Source < snapshot.Edges[j].Source
	})

	return snapshot, evidenceMap
}

func ensureNode(nodes map[string]*tfte.Node, serviceID string) {
	if serviceID == "" {
		return
	}
	if _, exists := nodes[serviceID]; !exists {
		nodes[serviceID] = &tfte.Node{Id: serviceID, ServiceName: serviceID, HealthStatus: 0}
	}
}
