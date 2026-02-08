package proofs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/rrb115/tfte/internal/engine"
	"github.com/rrb115/tfte/proto/gen/tfte"
)

// ProofBundle contains all data needed to reproduce a decision.
type ProofBundle struct {
	Manifest *Manifest
	Snapshot *tfte.GraphSnapshot
	Evidence map[string]*tfte.EdgeEvidence
	Events   []*tfte.Event
	// TODO: Add config used?
	// For Phase 4, we bundle graph + evidence + raw events events.
}

type Manifest struct {
	RootEventID string            `json:"root_event_id"`
	Timestamp   int64             `json:"timestamp"`
	Files       map[string]string `json:"files"` // filename -> sha256
}

// GenerateProofBundle creates a verifiable tarball.
func GenerateProofBundle(ctx context.Context, eng *engine.Engine, rootEventID string, ts int64) ([]byte, string, error) {
	// 1. Get Graph & Evidence
	snapshot, evidence, err := eng.GetGraphWithEvidence(ctx, ts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get graph: %w", err)
	}

	// 2. Fetch raw events
	// Engine uses 1 hour window.
	windowSize := int64(3600 * 1e9) // 1h in ns
	startTs := ts - windowSize
	if startTs < 0 {
		startTs = 0
	}

	eventsResp, err := eng.GetEvents(ctx, &tfte.GetEventsRequest{
		StartTs: startTs,
		EndTs:   ts,
		Limit:   10000,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch events: %w", err)
	}

	// Create buffers
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	manifest := &Manifest{
		RootEventID: rootEventID,
		Timestamp:   ts,
		Files:       make(map[string]string),
	}

	// Helper to add file
	addFile := func(name string, data interface{}) error {
		jsonBytes, err := json.MarshalIndent(data, "", "  ") // Indent for readability
		if err != nil {
			return err
		}

		// Hash
		hash := sha256.Sum256(jsonBytes)
		manifest.Files[name] = hex.EncodeToString(hash[:])

		// Add to tar
		header := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(jsonBytes)),
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tw.Write(jsonBytes); err != nil {
			return err
		}
		return nil
	}

	// Add Snapshot
	if err := addFile("snapshot.json", snapshot); err != nil {
		return nil, "", err
	}

	// Add Evidence
	if err := addFile("evidence.json", evidence); err != nil {
		return nil, "", err
	}

	// Add Events
	if err := addFile("events.json", eventsResp.Events); err != nil {
		return nil, "", err
	}

	// Add Manifest (last)
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	header := &tar.Header{
		Name: "manifest.json",
		Mode: 0600,
		Size: int64(len(manifestBytes)),
	}
	tw.WriteHeader(header)
	tw.Write(manifestBytes)

	// Close
	tw.Close()
	gw.Close()

	finalBytes := buf.Bytes()
	bundleHash := sha256.Sum256(finalBytes)

	return finalBytes, hex.EncodeToString(bundleHash[:]), nil
}
