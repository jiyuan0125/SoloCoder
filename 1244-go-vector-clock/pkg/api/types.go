package api

import (
	"vectorclock/pkg/vectorclock"
)

type ProduceRequest struct {
	NodeID  string `json:"node_id"`
	Content string `json:"content"`
}

type ProduceResponse struct {
	EventID uint64                 `json:"event_id"`
	Clock   vectorclock.VectorClock `json:"clock"`
}

type ReceiveRequest struct {
	TargetNodeID string                    `json:"target_node_id"`
	FromNodeID   string                    `json:"from_node_id"`
	RemoteClock  vectorclock.VectorClock   `json:"remote_clock"`
	Content      string                    `json:"content"`
}

type ReceiveResponse struct {
	EventID uint64                 `json:"event_id"`
	Clock   vectorclock.VectorClock `json:"clock"`
}

type ClockRequest struct {
	NodeID string `json:"node_id"`
}

type ClockResponse struct {
	NodeID string                   `json:"node_id"`
	Clock  vectorclock.VectorClock   `json:"clock"`
}

type CompareRequest struct {
	ClockA vectorclock.VectorClock `json:"clock_a"`
	ClockB vectorclock.VectorClock `json:"clock_b"`
}

type CompareResponse struct {
	Result string `json:"result"`
}

type EventsRequest struct {
	NodeID string `json:"node_id"`
}

type EventsResponse struct {
	NodeID string               `json:"node_id"`
	Events []*vectorclock.Event `json:"events"`
}

type PruneRequest struct {
	NodeID      string `json:"node_id"`
	RetainCount int    `json:"retain_count"`
}

type PruneResponse struct {
	NodeID       string `json:"node_id"`
	PrunedCount  int    `json:"pruned_count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
