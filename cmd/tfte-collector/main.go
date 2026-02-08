package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/rrb115/tfte/internal/storage"
	"github.com/rrb115/tfte/proto/gen/tfte"
)

// Simple struct to match the JSON input if it differs from Proto mapping,
// or we can use protojson.Unmarshal if the input is already in accurate shape.
// The spec says "ingest exported logs/traces as .ndjson... create adapters".
// We will assume a simple generic JSON envelope or specific fields.
// For Phase 1, let's assume the JSON is fairly close to the Event structure
// or we map it manually.

func main() {
	dbPath := flag.String("db", "./data/tfte.db", "Path to BadgerDB data directory")
	inputFile := flag.String("file", "", "Path to NDJSON file to ingest")
	flag.Parse()

	if *inputFile == "" {
		log.Fatal("Please provide --file <events.ndjson>")
	}

	store, err := storage.NewBadgerStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open storage: %v", err)
	}
	defer store.Close()

	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var batch []*tfte.Event
	batchSize := 100
	count := 0

	for scanner.Scan() {
		line := scanner.Bytes()
		var rawEvent struct {
			Id       string          `json:"id"`
			Type     string          `json:"type"`
			Service  string          `json:"service"`
			Host     string          `json:"host"`
			Ts       int64           `json:"ts"`
			Payload  json.RawMessage `json:"payload"`
			TraceIds []string        `json:"trace_ids"`
		}

		if err := json.Unmarshal(line, &rawEvent); err != nil {
			log.Printf("Skipping invalid JSON line: %v", err)
			continue
		}

		eventTypeVal := tfte.EventType_value[rawEvent.Type]

		var event tfte.Event
		event.Id = rawEvent.Id
		event.Type = tfte.EventType(eventTypeVal)
		event.Service = rawEvent.Service
		event.Host = rawEvent.Host
		event.Ts = rawEvent.Ts
		event.TraceIds = rawEvent.TraceIds

		if len(rawEvent.Payload) > 0 {
			event.Payload = []byte(rawEvent.Payload)
		}

		// Fill defaults if missing
		if event.Id == "" {
			event.Id = uuid.New().String()
		}
		if event.Ts == 0 {
			event.Ts = time.Now().UnixNano()
		}

		batch = append(batch, &event)
		if len(batch) >= batchSize {
			if err := store.IngestEvents(context.Background(), batch); err != nil {
				log.Fatalf("Failed to ingest batch: %v", err)
			}
			count += len(batch)
			batch = nil
			fmt.Printf("Ingested %d events...\r", count)
		}
	}

	if len(batch) > 0 {
		if err := store.IngestEvents(context.Background(), batch); err != nil {
			log.Fatalf("Failed to ingest name batch: %v", err)
		}
		count += len(batch)
	}

	fmt.Printf("\nDone. Total events ingested: %d\n", count)
}
