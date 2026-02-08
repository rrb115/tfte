package storage

import (
	"context"

	"github.com/rrb115/tfte/proto/gen/tfte"
)

// EventStore defines the interface for storing and retrieving events.
type EventStore interface {
	// IngestEvents stores a batch of events.
	IngestEvents(ctx context.Context, events []*tfte.Event) error

	// GetEvents retrieves events within a time range, optionally filtering by service.
	GetEvents(ctx context.Context, startTs, endTs int64, serviceFilter string, limit, offset int) ([]*tfte.Event, error)

	// Close closes the storage connection.
	Close() error
}

// GraphStore defines the interface for storing and retrieving graph snapshots.
type GraphStore interface {
	// SaveSnapshot stores a graph snapshot.
	SaveSnapshot(ctx context.Context, snapshot *tfte.GraphSnapshot) error

	// GetSnapshot retrieves the snapshot closest to the given timestamp (<= ts).
	GetSnapshot(ctx context.Context, ts int64) (*tfte.GraphSnapshot, error)
}

// Store combines EventStore and GraphStore.
type Store interface {
	EventStore
	GraphStore
}
