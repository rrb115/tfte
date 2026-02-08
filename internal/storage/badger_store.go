package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/rrb115/tfte/proto/gen/tfte"
	"google.golang.org/protobuf/proto"
)

type BadgerStore struct {
	db *badger.DB
}

// NewBadgerStore creates a new BadgerDB-backed store.
func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path)
	// Optimize for lower memory usage if needed, or stick to defaults for now.
	opts.Logger = nil // Disable default logger for cleaner output

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &BadgerStore{db: db}, nil
}

func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// Event Key Format: "event:<timestamp_ns>:<event_id>"
// This allows range scans by time.

func eventKey(ts int64, id string) []byte {
	return []byte(fmt.Sprintf("event:%020d:%s", ts, id))
}

func (s *BadgerStore) IngestEvents(ctx context.Context, events []*tfte.Event) error {
	return s.db.Update(func(txn *badger.Txn) error {
		for _, e := range events {
			data, err := proto.Marshal(e)
			if err != nil {
				return fmt.Errorf("failed to marshal event %s: %w", e.Id, err)
			}
			key := eventKey(e.Ts, e.Id)
			if err := txn.Set(key, data); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetEvents retrieves events within a time range, optionally filtering by service.
func (s *BadgerStore) GetEvents(ctx context.Context, startTs, endTs int64, serviceFilter string, limit, offset int) ([]*tfte.Event, error) {
	var events []*tfte.Event

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()

		// Keys are "event:<ts>:<uuid>"
		// Start seeking from startTs
		startKey := eventKey(startTs, "")
		// End before endTs (exclusive) - use a prefix for the end time to ensure all events up to endTs are included
		endKeyPrefix := []byte(fmt.Sprintf("event:%020d", endTs))

		skipped := 0
		for it.Seek(startKey); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()

			// Check if we passed the end time (exclusive)
			if bytes.Compare(k, endKeyPrefix) >= 0 {
				break
			}

			// Simple prefix check to ensure it is an event (should always be true based on seek)
			if !bytes.HasPrefix(k, []byte("event:")) {
				continue // Should not happen with correct seek, but good for robustness
			}

			err := item.Value(func(val []byte) error {
				var evt tfte.Event
				if err := proto.Unmarshal(val, &evt); err != nil {
					return err
				}

				// Service filter
				if serviceFilter != "" && evt.Service != serviceFilter {
					return nil // Skip this event if it doesn't match the filter
				}

				// Pagination: Offset
				if skipped < offset {
					skipped++
					return nil // Skip this event if we are still in the offset range
				}

				// Limit
				if len(events) < limit {
					events = append(events, &evt)
				}
				return nil
			})
			if err != nil {
				return err
			}

			if limit > 0 && len(events) >= limit {
				break // Stop if we've collected enough events
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return events, nil
}

// GraphSnapshot Key Format: "snapshot:<timestamp_ns>"

func snapshotKey(ts int64) []byte {
	return []byte(fmt.Sprintf("snapshot:%020d", ts))
}

func (s *BadgerStore) SaveSnapshot(ctx context.Context, snapshot *tfte.GraphSnapshot) error {
	data, err := proto.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(snapshotKey(snapshot.Timestamp), data)
	})
}

func (s *BadgerStore) GetSnapshot(ctx context.Context, ts int64) (*tfte.GraphSnapshot, error) {
	var snapshot tfte.GraphSnapshot

	err := s.db.View(func(txn *badger.Txn) error {
		// seek to the key, if not found, maybe look backwards?
		// For now, exact match or closest previous?
		// "GetSnapshot retrieves the snapshot closest to the given timestamp (<= ts)"

		opts := badger.DefaultIteratorOptions
		opts.Reverse = true // We want to look backwards
		it := txn.NewIterator(opts)
		defer it.Close()

		it.Seek(snapshotKey(ts))
		if !it.Valid() {
			return badger.ErrKeyNotFound
		}

		item := it.Item()
		k := string(item.Key())

		// Verify it is a snapshot key
		if len(k) < 9 || k[:9] != "snapshot:" {
			return badger.ErrKeyNotFound
		}

		return item.Value(func(val []byte) error {
			return proto.Unmarshal(val, &snapshot)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil // Return nil if no snapshot found
		}
		return nil, err
	}

	return &snapshot, nil
}
