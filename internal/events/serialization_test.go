package events_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rrb115/tfte/proto/gen/tfte"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestEventSerialization(t *testing.T) {
	evt := &tfte.Event{
		Id:       uuid.New().String(),
		Type:     tfte.EventType_RPC_ERROR,
		Service:  "frontend",
		Host:     "fe-1",
		Ts:       time.Now().UnixNano(),
		TraceIds: []string{"trace-123"},
	}

	// Test Proto Marshal/Unmarshal
	data, err := proto.Marshal(evt)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var evt2 tfte.Event
	if err := proto.Unmarshal(data, &evt2); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if evt.Id != evt2.Id {
		t.Errorf("ID mismatch: got %v, want %v", evt2.Id, evt.Id)
	}
	if evt.Type != evt2.Type {
		t.Errorf("Type mismatch: got %v, want %v", evt2.Type, evt.Type)
	}
}

func TestJSONSerialization(t *testing.T) {
	evt := &tfte.Event{
		Id:      "test-id-1",
		Type:    tfte.EventType_HEALTH_CHANGE,
		Service: "db-shard-1",
		Ts:      1670000000000000000,
	}

	// Test JSON Marshal/Unmarshal (for collector ingestion)
	data, err := protojson.Marshal(evt)
	if err != nil {
		t.Fatalf("Failed to marshal json: %v", err)
	}

	t.Logf("JSON: %s", string(data))

	var evt2 tfte.Event
	if err := protojson.Unmarshal(data, &evt2); err != nil {
		t.Fatalf("Failed to unmarshal json: %v", err)
	}

	if evt.Id != evt2.Id {
		t.Errorf("ID mismatch: got %v, want %v", evt2.Id, evt.Id)
	}
}
