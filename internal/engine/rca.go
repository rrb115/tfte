package engine

import (
	"sort"

	"github.com/rrb115/tfte/proto/gen/tfte"
)

// GetRootCauseCandidates identifies nodes that are likely root causes.
// Heuristic: Unhealthy nodes whose failure is NOT well-explained by downstream dependencies.
// i.e. Max(Outgoing Edge Causal Score) is low.
func (e *Engine) GetRootCauseCandidates(snapshot *tfte.GraphSnapshot) []*tfte.Node {
	candidates := make([]*tfte.Node, 0)

	// 1. Map nodes by ID for lookups
	nodeMap := make(map[string]*tfte.Node)
	for _, n := range snapshot.Nodes {
		nodeMap[n.Id] = n
	}

	// 2. Calculate Max Outgoing Score for each node
	// Outgoing edge: Source = Node.
	// If A->B has high score, A's failure is caused by B.
	// So we look for nodes where ALL outgoing edges have low scores (or no outgoing edges).

	maxExplainedScore := make(map[string]float64)

	for _, edge := range snapshot.Edges {
		if edge.CausalConfidence > maxExplainedScore[edge.Source] {
			maxExplainedScore[edge.Source] = edge.CausalConfidence
		}
	}

	// 3. Select Candidates
	for _, n := range snapshot.Nodes {
		if n.HealthStatus == 0 {
			continue
		} // Only unhealthy nodes can be root causes (usually)

		explainedScore := maxExplainedScore[n.Id]

		// Threshold: If explained score < 0.5, we consider it a potential root cause.
		// i.e. We are less than 50% confident that a downstream node caused this.
		if explainedScore < 0.5 {
			candidates = append(candidates, n)
		}
	}

	// 4. Sort by "Severity" or Impact?
	// For now, sort by Downstream Failures? (which we haven't computed yet)
	// Or just by HealthStatus (Dead > Degraded)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].HealthStatus > candidates[j].HealthStatus
	})

	return candidates
}
