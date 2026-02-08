package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rrb115/tfte/internal/api"
	"github.com/rrb115/tfte/internal/engine"
	"github.com/rrb115/tfte/internal/storage"
)

func main() {
	dbPath := flag.String("db", "./data/tfte.db", "Path to BadgerDB data directory")
	port := flag.Int("port", 9090, "gRPC server port")
	httpPort := flag.Int("http-port", 8081, "HTTP server port")
	flag.Parse()

	// 1. Storage
	store, err := storage.NewBadgerStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open storage: %v", err)
	}
	defer store.Close()

	// 2. Engine
	eng := engine.NewEngine(store)

	// 3. API Server
	srv := api.NewServer(eng, store)

	// 4. Run Servers
	go func() {
		if err := api.RunGRPCServer(*port, srv); err != nil {
			log.Fatalf("Failed to run gRPC server: %v", err)
		}
	}()

	go func() {
		httpGateway := api.NewHTTPGateway(srv)
		addr := fmt.Sprintf(":%d", *httpPort)
		fmt.Printf("HTTP server listening on %s\n", addr)
		if err := http.ListenAndServe(addr, httpGateway); err != nil {
			log.Fatalf("Failed to run HTTP server: %v", err)
		}
	}()

	// Wait for interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down TFTE Core...")
}
