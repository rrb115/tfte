# TFTE (Temporal Failure Topology Explorer)

Deterministic, time-travel replay of failures across a microservice graph. TFTE ingests events (health changes, RPC calls/errors), reconstructs a service topology at any timestamp, scores causal edges, and surfaces root-cause candidates with explainable evidence.

## Components
- **Storage**: BadgerDB via `internal/storage` for events and snapshots.
- **Engine**: `internal/engine` replays events, computes node health/amplification, edge causal confidence, and RCA candidates.
- **API**: gRPC + HTTP gateway in `cmd/tfte-core` exposing `/api/events`, `/api/graph`, `/api/evidence`, `/api/root-cause`, `/api/proof`.
- **UI**: React/Vite app in `ui/` rendering the graph and timeline.

## Quickstart (local)
1) Backend
```
go run ./cmd/tfte-core -db ./data/tfte.db -port 9090 -http-port 8081
```

2) Seed events (pick one):
- Run the simulator (sends RPC_CALL/RPC_ERROR traffic to the API):
```
go run ./cmd/tfte-sim
```
- Or ingest NDJSON of events:
```
go run ./cmd/tfte-collector --file ./events.ndjson --db ./data/tfte.db
```

3) UI
```
cd ui
npm install
npm run dev   # visit http://localhost:5173
```

## Event shapes (JSON payloads)
All timestamps are **milliseconds** since epoch.
- `HEALTH_CHANGE`: `{ "new_status": "DOWN" | "DEGRADED" | "HEALTHY" }`
- `RPC_CALL`: `{ "dest_service": "service-b", "method": "GET /foo", "status_code": 200, "latency_ms": 12 }`
- `RPC_ERROR`: `{ "dest_service": "service-b", "method": "POST /bar", "error_code": "503", "error_message": "Service Unavailable" }`

POST events to `/api/events` as an array:
```json
[
  {
    "id": "evt-1",
    "type": "HEALTH_CHANGE",
    "service": "auth",
    "ts": 1730000000000,
    "payload": { "new_status": "DOWN" }
  },
  {
    "id": "evt-2",
    "type": "RPC_ERROR",
    "service": "api-gateway",
    "ts": 1730000000500,
    "payload": { "dest_service": "auth", "method": "POST /login", "error_code": "503", "error_message": "backend down" }
  }
]
```

## Key endpoints (HTTP gateway)
- `GET /api/graph?timestamp=<ms>` → `GraphSnapshot` (nodes, edges, causal scores).
- `POST /api/events` → ingest events batch.
- `GET /api/evidence?source=<svc>&target=<svc>&timestamp=<ms>` → edge evidence breakdown.
- `GET /api/root-cause?timestamp=<ms>` → candidate root causes.
- `GET /api/proof?root_event_id=<id>&timestamp=<ms>` → downloadable proof bundle.

## Notes
- Scoring window and computations are in **milliseconds** to match stored events.
- BadgerDB files default to `./data/tfte.db`; override with `-db` flag in `cmd/tfte-core`.
