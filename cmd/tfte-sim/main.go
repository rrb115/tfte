package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rrb115/tfte/proto/gen/tfte"
)

const (
	APIBase = "http://localhost:8081/api/events"
)

func main() {
	log.Println("Starting TFTE COMPLEX Simulator...")
	log.Println("Topology: 15+ Microservices")

	var wg sync.WaitGroup
	// Create multiple trace generators for different user behaviors
	wg.Add(4)

	// 1. Browsing Traffic (High Volume, Read Heavy)
	go func() {
		defer wg.Done()
		log.Println("Starting Browsing Traffic (Frontend -> API -> Product -> Redis/DB)")
		// Loop forever
		runSimulationLoop(100*time.Millisecond, simulateBrowsing)
	}()

	// 2. Search Traffic (Medium Volume, Elasticsearch Heavy)
	go func() {
		defer wg.Done()
		log.Println("Starting Search Traffic (Mobile -> API -> Search -> Elastic)")
		runSimulationLoop(300*time.Millisecond, simulateSearch)
	}()

	// 3. Checkout Traffic (Low Volume, Heavy Dependency Chain, Critical)
	go func() {
		defer wg.Done()
		log.Println("Starting Checkout Traffic (Frontend -> API -> Order -> Payment/Inventory)")
		runSimulationLoop(800*time.Millisecond, simulateCheckout)
	}()

	// 4. Background Jobs (Cron -> Analytics)
	go func() {
		defer wg.Done()
		log.Println("Starting Background Jobs (Cron -> Analytics -> BigQuery)")
		runSimulationLoop(2000*time.Millisecond, simulateAnalytics)
	}()

	wg.Wait()
}

func runSimulationLoop(interval time.Duration, simFunc func(string, int64)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		traceID := uuid.New().String()
		now := time.Now().UnixNano()
		simFunc(traceID, now)
	}
}

// --- Scenarios ---

// simulateBrowsing: Simple read path with caching
func simulateBrowsing(traceID string, startTs int64) {
	// Determine failure mode based on time
	// Scenario 1: Redis Slow (t=30s to 50s)
	isRedisSlow := isTimeInWindow(30, 50)

	ts := startTs

	// User -> Frontend
	sendRpcCall("user", "web-frontend", "GET /products/123", 200, 15, traceID, ts)
	ts += 15 * 1e6

	// Frontend -> API Gateway
	sendRpcCall("web-frontend", "api-gateway", "GET /api/v1/products/123", 200, 10, traceID, ts)
	ts += 10 * 1e6

	// API Gateway -> Product Service
	sendRpcCall("api-gateway", "product-service", "GET /products/123", 200, 8, traceID, ts)
	ts += 8 * 1e6

	// Product Service -> Redis (Cache Check)
	// 80% Hit Rate
	if rand.Float64() < 0.8 {
		latency := int64(2)
		if isRedisSlow {
			latency = 500
		} // Latency spike

		sendRpcCall("product-service", "redis-cache", "GET product:123", 200, latency, traceID, ts)
	} else {
		// Cache Miss -> DB
		sendRpcCall("product-service", "redis-cache", "GET product:123", 200, 2, traceID, ts)
		ts += 2 * 1e6

		sendRpcCall("product-service", "postgres-primary", "SELECT * FROM products WHERE id=123", 200, 25, traceID, ts)
	}
}

// simulateSearch: Search path
func simulateSearch(traceID string, startTs int64) {
	ts := startTs

	sendRpcCall("mobile-app", "mobile-api", "GET /search?q=shoes", 200, 40, traceID, ts)
	ts += 40 * 1e6

	sendRpcCall("mobile-api", "product-search-service", "GET /search", 200, 15, traceID, ts)
	ts += 15 * 1e6

	// Heavy weight query
	sendRpcCall("product-search-service", "elasticsearch-cluster", "POST /_search", 200, 150, traceID, ts)
}

// simulateCheckout: Critical path with external dependencies
func simulateCheckout(traceID string, startTs int64) {
	// Scenario 2: Payment Gateway Outage (t=60s to 90s)
	isPaymentDown := isTimeInWindow(60, 90)

	ts := startTs

	// Frontend -> API
	sendRpcCall("web-frontend", "api-gateway", "POST /checkout", 200, 20, traceID, ts)
	ts += 20 * 1e6

	// API -> Order Service
	sendRpcCall("api-gateway", "order-service", "POST /orders", 200, 15, traceID, ts)
	ts += 15 * 1e6

	// Order -> User Service (Validate)
	sendRpcCall("order-service", "user-service", "GET /users/456", 200, 10, traceID, ts)
	ts += 10 * 1e6

	// User -> Auth (Token)
	sendRpcCall("user-service", "auth-service", "POST /verify", 200, 5, traceID, ts)
	ts += 5 * 1e6

	// Order -> Inventory (Reserve)
	sendRpcCall("order-service", "inventory-service", "POST /reserve", 200, 30, traceID, ts)
	ts += 30 * 1e6

	// Inventory -> DB
	sendRpcCall("inventory-service", "postgres-inventory", "UPDATE items SET stock=stock-1", 200, 15, traceID, ts)
	ts += 15 * 1e6

	// Order -> Payment Gateway (External)
	if isPaymentDown {
		// High chance of failure
		if rand.Float64() < 0.9 {
			sendRpcError("order-service", "payment-gateway", "POST /charge", "503", "Service Unavailable", traceID, ts)

			// Cascading error back to user
			sendRpcError("api-gateway", "order-service", "POST /orders", "500", "Payment Failed", traceID, ts+10*1e6)
			return
		}
	}

	sendRpcCall("order-service", "payment-gateway", "POST /charge", 200, 450, traceID, ts) // External calls are slow
	ts += 450 * 1e6

	// Notification (Async)
	sendRpcCall("order-service", "notification-service", "POST /email", 202, 10, traceID, ts)
}

func simulateAnalytics(traceID string, startTs int64) {
	ts := startTs
	sendRpcCall("cron-scheduler", "analytics-aggregator", "POST /run-job", 200, 5, traceID, ts)
	ts += 5 * 1e6
	sendRpcCall("analytics-aggregator", "bigquery-loader", "POST /load", 200, 2000, traceID, ts) // Very slow
}

// Helper: Check if current simulation time (modulo 120s loop) is in window
func isTimeInWindow(startSec, endSec int) bool {
	// Loop every 120s
	t := time.Now().Unix() % 120
	return t >= int64(startSec) && t < int64(endSec)
}

// --- Plumbing ---

func sendRpcCall(src, dst, method string, status int32, latency int64, traceID string, ts int64) {
	// Use JSON payload mapping to tfte.RpcCall struct
	payload := map[string]interface{}{
		"source_service": src,
		"dest_service":   dst,
		"method":         method,
		"status_code":    status,
		"latency_ms":     latency,
	}
	data, _ := json.Marshal(payload)

	evt := &tfte.Event{
		Id:       uuid.New().String(),
		Type:     tfte.EventType_RPC_CALL,
		Service:  src,
		Ts:       ts / 1e6, // Convert to Milliseconds
		Payload:  data,
		TraceIds: []string{traceID},
	}
	sendEvent(evt)
}

func sendRpcError(src, dst, method, code, msg, traceID string, ts int64) {
	payload := map[string]interface{}{
		"source_service": src,
		"dest_service":   dst,
		"method":         method,
		"error_code":     code,
		"error_message":  msg,
	}
	data, _ := json.Marshal(payload)

	evt := &tfte.Event{
		Id:       uuid.New().String(),
		Type:     tfte.EventType_RPC_ERROR,
		Service:  src,
		Ts:       ts / 1e6, // Convert to Milliseconds
		Payload:  data,
		TraceIds: []string{traceID},
	}
	sendEvent(evt)
}

func sendEvent(evt *tfte.Event) {
	payload := []*tfte.Event{evt}
	body, _ := json.Marshal(payload)

	// Create client with short timeout
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(APIBase, "application/json", bytes.NewBuffer(body))
	if err != nil {
		// Silent fail, it's a sim
		return
	}
	defer resp.Body.Close()
}
