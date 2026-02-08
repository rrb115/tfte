package proofs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/rrb115/tfte/internal/engine"
	"github.com/rrb115/tfte/proto/gen/tfte"
)

type MockStore struct {
	events []*tfte.Event
}

func (m *MockStore) IngestEvents(ctx context.Context, events []*tfte.Event) error {
	m.events = append(m.events, events...)
	return nil
}

func (m *MockStore) GetEvents(ctx context.Context, startTs, endTs int64, serviceFilter string, limit, offset int) ([]*tfte.Event, error) {
	var result []*tfte.Event
	skipped := 0
	for _, e := range m.events {
		if e.Ts >= startTs && e.Ts <= endTs {
			if skipped < offset {
				skipped++
				continue
			}
			result = append(result, e)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockStore) Close() error                                                         { return nil }
func (m *MockStore) SaveSnapshot(ctx context.Context, snapshot *tfte.GraphSnapshot) error { return nil }
func (m *MockStore) GetSnapshot(ctx context.Context, ts int64) (*tfte.GraphSnapshot, error) {
	return nil, nil
}

func TestGenerateProofBundle(t *testing.T) {
	// 1. Setup
	store := &MockStore{}

	// Create some events
	// A -> B check
	rpcPayload, _ := json.Marshal(map[string]string{"dest_service": "B"})
	hcPayload, _ := json.Marshal(map[string]string{"new_status": "DOWN"})

	events := []*tfte.Event{
		{Id: "1", Ts: 100, Type: tfte.EventType_RPC_CALL, Service: "A", Payload: rpcPayload},
		{Id: "2", Ts: 200, Type: tfte.EventType_HEALTH_CHANGE, Service: "B", Payload: hcPayload},
	}
	store.IngestEvents(context.Background(), events)

	eng := engine.NewEngine(store)

	// 2. Generate Bundle at T=300
	ctx := context.Background()
	bundleBytes, hash, err := GenerateProofBundle(ctx, eng, "2", 300)
	if err != nil {
		t.Fatalf("GenerateProofBundle failed: %v", err)
	}

	if len(bundleBytes) == 0 {
		t.Fatal("Bundle bytes is empty")
	}
	if hash == "" {
		t.Fatal("Hash is empty")
	}

	// 3. Verify Tar Gz
	gr, err := gzip.NewReader(bytes.NewReader(bundleBytes))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	foundFiles := make(map[string]bool)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Tar error: %v", err)
		}

		foundFiles[header.Name] = true

		// Read content
		content, _ := io.ReadAll(tr)

		if header.Name == "events.json" {
			var evts []*tfte.Event
			if err := json.Unmarshal(content, &evts); err != nil {
				t.Errorf("Failed to unmarshal events.json: %v", err)
			}
			if len(evts) != 2 {
				t.Errorf("Expected 2 events in bundle, got %d", len(evts))
			}
		}
	}

	expectedFiles := []string{"manifest.json", "snapshot.json", "evidence.json", "events.json"}
	for _, f := range expectedFiles {
		if !foundFiles[f] {
			t.Errorf("Missing file in bundle: %s", f)
		}
	}
}
