package events

import (
	"encoding/json"
	"fmt"

	"github.com/rrb115/tfte/proto/gen/tfte"
	"google.golang.org/protobuf/proto"
)

// ParseHealthChange parses the payload of a HEALTH_CHANGE event.
// It tries Proto unmarshal first, then JSON if that fails (for backward compat/flexibility).
func ParseHealthChange(payload []byte) (*tfte.HealthChange, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty payload")
	}

	var msg tfte.HealthChange
	if err := proto.Unmarshal(payload, &msg); err == nil {
		return &msg, nil
	}

	// Try JSON
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ParseRpcCall parses RPC_CALL payload
func ParseRpcCall(payload []byte) (*tfte.RpcCall, error) {
	var msg tfte.RpcCall
	if err := proto.Unmarshal(payload, &msg); err == nil {
		return &msg, nil
	}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ParseRpcError parses RPC_ERROR payload
func ParseRpcError(payload []byte) (*tfte.RpcError, error) {
	var msg tfte.RpcError
	if err := proto.Unmarshal(payload, &msg); err == nil {
		return &msg, nil
	}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
