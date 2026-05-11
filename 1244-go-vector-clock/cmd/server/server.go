package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"vectorclock/pkg/api"
	"vectorclock/pkg/vectorclock"
)

var manager = vectorclock.NewNodeManager()

func main() {
	port := flag.String("port", "", "Server port (default: 8080)")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("VC_PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8080"
		}
	}

	http.HandleFunc("/produce", produceHandler)
	http.HandleFunc("/receive", receiveHandler)
	http.HandleFunc("/clock", clockHandler)
	http.HandleFunc("/compare", compareHandler)
	http.HandleFunc("/events", eventsHandler)
	http.HandleFunc("/prune", pruneHandler)

	addr := ":" + *port
	log.Printf("Vector Clock server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, errMsg string) {
	respondJSON(w, status, &api.ErrorResponse{Error: errMsg})
}

func produceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ProduceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NodeID == "" {
		respondError(w, http.StatusBadRequest, "node_id is required")
		return
	}

	event, err := manager.Produce(req.NodeID, req.Content)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, &api.ProduceResponse{
		EventID: event.EventID,
		Clock:   event.Clock,
	})
}

func receiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ReceiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TargetNodeID == "" || req.FromNodeID == "" {
		respondError(w, http.StatusBadRequest, "target_node_id and from_node_id are required")
		return
	}

	event, err := manager.Receive(req.TargetNodeID, req.FromNodeID, req.RemoteClock, req.Content)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, &api.ReceiveResponse{
		EventID: event.EventID,
		Clock:   event.Clock,
	})
}

func clockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		respondError(w, http.StatusBadRequest, "node_id query parameter is required")
		return
	}

	clock, err := manager.GetClock(nodeID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, &api.ClockResponse{
		NodeID: nodeID,
		Clock:  clock,
	})
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result := vectorclock.CompareClocks(req.ClockA, req.ClockB)
	resultStr := causalityToString(result)

	respondJSON(w, http.StatusOK, &api.CompareResponse{
		Result: resultStr,
	})
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		respondError(w, http.StatusBadRequest, "node_id query parameter is required")
		return
	}

	events, err := manager.GetEvents(nodeID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, &api.EventsResponse{
		NodeID: nodeID,
		Events: events,
	})
}

func pruneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.PruneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NodeID == "" {
		respondError(w, http.StatusBadRequest, "node_id is required")
		return
	}

	pruned, err := manager.Prune(req.NodeID, req.RetainCount)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, &api.PruneResponse{
		NodeID:      req.NodeID,
		PrunedCount: pruned,
	})
}

func causalityToString(c vectorclock.CausalityType) string {
	switch c {
	case vectorclock.CausalitySame:
		return "same event"
	case vectorclock.CausalityBefore:
		return "A causally before B"
	case vectorclock.CausalityAfter:
		return "B causally before A"
	case vectorclock.CausalityConcurrent:
		return "concurrent"
	default:
		return fmt.Sprintf("unknown (%d)", c)
	}
}
