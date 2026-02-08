package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rrb115/tfte/proto/gen/tfte"
)

// HTTPGateway wraps the GRPC server methods to provide a simple REST API.
type HTTPGateway struct {
	server *Server
}

func NewHTTPGateway(server *Server) *HTTPGateway {
	return &HTTPGateway{server: server}
}

func (h *HTTPGateway) enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (h *HTTPGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	switch r.URL.Path {
	case "/api/events":
		if r.Method == "POST" {
			h.handleIngest(w, r)
		} else {
			h.handleGetEvents(w, r)
		}
	case "/api/graph":
		h.handleGetGraph(w, r)
	case "/api/evidence":
		h.handleGetEvidence(w, r)
	case "/api/root-cause":
		h.handleGetRootCause(w, r)
	case "/api/proof":
		h.handleGetProof(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *HTTPGateway) handleIngest(w http.ResponseWriter, r *http.Request) {
	var events []*tfte.Event
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.server.IngestEvents(r.Context(), &tfte.IngestEventsRequest{Events: events})
	h.writeJSON(w, resp, err)
}

func (h *HTTPGateway) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	startTs, _ := strconv.ParseInt(q.Get("start_ts"), 10, 64)
	endTs, _ := strconv.ParseInt(q.Get("end_ts"), 10, 64)
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	if limit == 0 {
		limit = 100
	}

	resp, err := h.server.GetEvents(r.Context(), &tfte.GetEventsRequest{
		StartTs:       startTs,
		EndTs:         endTs,
		ServiceFilter: q.Get("service_filter"),
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
	h.writeJSON(w, resp, err)
}

func (h *HTTPGateway) handleGetGraph(w http.ResponseWriter, r *http.Request) {
	ts, _ := strconv.ParseInt(r.URL.Query().Get("timestamp"), 10, 64)
	resp, err := h.server.GetGraphSnapshot(r.Context(), &tfte.GetGraphRequest{Timestamp: ts})
	h.writeJSON(w, resp, err)
}

func (h *HTTPGateway) handleGetEvidence(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ts, _ := strconv.ParseInt(q.Get("timestamp"), 10, 64)

	resp, err := h.server.GetEdgeEvidence(r.Context(), &tfte.GetEdgeEvidenceRequest{
		SourceService:       q.Get("source"),
		TargetService:       q.Get("target"),
		AssociatedTimestamp: ts,
	})
	h.writeJSON(w, resp, err)
}

func (h *HTTPGateway) handleGetRootCause(w http.ResponseWriter, r *http.Request) {
	ts, _ := strconv.ParseInt(r.URL.Query().Get("timestamp"), 10, 64)
	resp, err := h.server.GetRootCause(r.Context(), &tfte.GetRootCauseRequest{Timestamp: ts})
	h.writeJSON(w, resp, err)
}

func (h *HTTPGateway) handleGetProof(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ts, _ := strconv.ParseInt(q.Get("timestamp"), 10, 64)
	resp, err := h.server.GetProofBundle(r.Context(), &tfte.GetProofBundleRequest{
		RootEventId: q.Get("root_event_id"),
		Timestamp:   ts,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return bytes directly? Or JSON?
	// JSON response wrapper is safer for now.
	h.writeJSON(w, resp, nil)
}

func (h *HTTPGateway) writeJSON(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
