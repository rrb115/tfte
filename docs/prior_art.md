# Prior Art & Novelty Check

## Summary
TFTE (Temporal Failure Topology Explorer) is a tool for reconstructing, scoring, and visually replaying failure propagation in distributed systems. It differentiates itself by focusing on **deterministic, rule-based causal inference** with **explainable proof bundles**, avoiding the "black box" nature of ML-based approaches, and providing a **time-evolving topology** view rather than just trace-based or log-based views.

## Related Tools & Research

### 1. Distributed Tracing (Jaeger, Zipkin, Tempo)
*   **Summary**: These tools capture and visualize the lifecycle of individual requests (traces) across microservices. They are excellent for identifying latency bottlenecks and errors in specific transaction paths.
*   **Gap**: Tracing tools visualize *instances* of requests. They do not typically aggregate these into a high-level, time-evolving "failure topology" that shows how faults cascade across the system over time. They lack the "amplification score" and specific "causal confidence" metrics for edges based on aggregate windows.
*   **TFTE Differentiator**: TFTE aggregates individual events (including traces) over time windows to infer and score persistent causal dependencies and blast radius, rather than just showing individual Gantt charts.

### 2. ShiViz
*   **Summary**: ShiViz [Link](https://bestchai.bitbucket.io/shiviz/) visualizes distributed system executions as interactive time-space diagrams (Lamport diagrams) generated from logs. It is designed to understand concurrent event ordering and specific execution patterns.
*   **Gap**: ShiViz focuses on happens-before relationships and ordering for debugging concurrency. It does not provide a system-wide "health topology" view, nor does it score "failure amplification" or attempt to rank "root cause candidates" based on failure propagation rules.
*   **TFTE Differentiator**: TFTE provides a higher-level Service Dependency Graph view overlaid with health state and causal weights, rather than a raw message-ordering diagram.

### 3. Root Cause Analysis (RCA) Libraries (PyRCA, DoWhy)
*   **Summary**: Libraries like Salesforce's PyRCA [Link](https://github.com/salesforce/PyRCA) and DoWhy use statistical methods and causal structure learning (PC algorithm, etc.) to infer causal graphs from metrics.
*   **Gap**: These are primarily statistical/ML libraries. They often operate as "black boxes" or probabilistic models that can be hard to audit. They typically lack an interactive, time-replay UI for operators to "watch" the failure unfold.
*   **TFTE Differentiator**: TFTE uses **deterministic, explainable rules** (e.g., "RPC error within 100ms leads to Degradation") rather than probabilistic inference. Every edge in TFTE has a human-readable "proof" attached.

### 4. Simulators & Chaos Engineering (Gremlin, Chaos Mesh)
*   **Summary**: These tools *induce* failures. Some provide dashboards to see the *result*, but their primary focus is injection, not ensuring the *post-hoc reconstruction* and *causal explanation* of an uncontrolled outage.
*   **Gap**: They generate the chaos; TFTE explains the chaos.

## Novelty Statement
**TFTE Novelty Confirmed.**
While components of TFTE exist in isolation (tracing for data, RCA for math, graph viz for topology), no open-source tool currently combines:
1.  **Time-travel replay** of failure propagation on a topology graph.
2.  **Deterministic, rule-based scoring** with distinct "amplification factors" and "provenance proofs".
3.  **Auditable export bundles** that allow a third party to verify the logic without access to the original raw data.

TFTE fills the gap between low-level trace inspection (too granular) and high-level metric dashboards (too aggregate, no causality), providing an **explainable narrative** of an outage.
