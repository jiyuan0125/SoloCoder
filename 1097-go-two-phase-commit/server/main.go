package main

import (
	"encoding/json"
	"go-two-phase-commit/common"
	"go-two-phase-commit/tpc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	coordinator *tpc.InMemoryCoordinator
)

func main() {
	coordinator = tpc.NewInMemoryCoordinator()

	participant1 := tpc.NewInMemoryParticipant("participant-1")
	participant2 := tpc.NewInMemoryParticipant("participant-2")
	participant3 := tpc.NewInMemoryParticipant("participant-3")

	coordinator.AddParticipant(participant1)
	coordinator.AddParticipant(participant2)
	coordinator.AddParticipant(participant3)

	log.Println("Coordinator recovery...")
	coordinator.Recover()

	http.HandleFunc("/api/tx/begin", handleBeginTx)
	http.HandleFunc("/api/tx/prepare", handlePrepareTx)
	http.HandleFunc("/api/tx/commit", handleCommitTx)
	http.HandleFunc("/api/tx/rollback", handleRollbackTx)
	http.HandleFunc("/api/tx/status", handleTxStatus)

	http.HandleFunc("/api/participants", handleListParticipants)
	http.HandleFunc("/api/participants/data", handleGetParticipantData)
	http.HandleFunc("/api/participants/add", handleAddParticipant)
	http.HandleFunc("/api/participants/remove", handleRemoveParticipant)
	http.HandleFunc("/api/participants/set-unavailable", handleSetParticipantUnavailable)

	server := &http.Server{
		Addr: ":8080",
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down server...")
		os.Exit(0)
	}()

	log.Println("Server starting on port 8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleBeginTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tx, err := coordinator.Begin()
	response := common.BeginTxResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
		response.TxID = tx.ID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePrepareTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.PrepareTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	operations := make(map[string][]tpc.Operation)
	for participantID, ops := range req.Operations {
		tpcOps := make([]tpc.Operation, len(ops))
		for i, op := range ops {
			tpcOps[i] = tpc.Operation{
				Type: tpc.OperationType(op.Type),
				Key:  op.Key,
				Data: op.Data,
			}
		}
		operations[participantID] = tpcOps
	}

	_, err := coordinator.Prepare(req.TxID, operations)
	response := common.PrepareTxResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
		response.Status = common.Prepared.String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCommitTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.CommitTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := coordinator.Commit(req.TxID)
	response := common.CommitTxResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
		response.Status = common.TransactionStatus(tx.Status).String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleRollbackTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RollbackTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := coordinator.Rollback(req.TxID)
	response := common.RollbackTxResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
		response.Status = common.TransactionStatus(tx.Status).String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTxStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.StatusTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := coordinator.GetStatus(req.TxID)
	response := common.StatusTxResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
		response.TxID = tx.ID
		response.Status = common.TransactionStatus(tx.Status).String()
		response.CreatedAt = tx.CreatedAt
		response.UpdatedAt = tx.UpdatedAt
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListParticipants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	participants := coordinator.GetParticipants()
	participantStatuses := make([]common.ParticipantStatus, len(participants))
	for i, p := range participants {
		status := p.GetStatus()
		participantStatuses[i] = common.ParticipantStatus{
			ID:           status.ID,
			PreparedTx:   status.PreparedTx,
			Unavailable:  status.Unavailable,
			StorageSize:  status.StorageSize,
		}
	}

	response := common.ListParticipantsResponse{
		Success:      true,
		Participants: participantStatuses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetParticipantData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.GetParticipantDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	participants := coordinator.GetParticipants()
	var targetParticipant tpc.ITransactionParticipant
	for _, p := range participants {
		if p.ID() == req.ParticipantID {
			targetParticipant = p
			break
		}
	}

	response := common.GetParticipantDataResponse{}
	if targetParticipant == nil {
		response.Success = false
		response.Error = "participant not found"
	} else {
		response.Success = true
		response.Data = targetParticipant.GetData()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAddParticipant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AddParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	participant := tpc.NewInMemoryParticipant(req.ParticipantID)
	err := coordinator.AddParticipant(participant)
	response := common.AddParticipantResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleRemoveParticipant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RemoveParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := coordinator.RemoveParticipant(req.ParticipantID)
	response := common.RemoveParticipantResponse{}
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleSetParticipantUnavailable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SetParticipantUnavailableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	participants := coordinator.GetParticipants()
	var targetParticipant tpc.ITransactionParticipant
	for _, p := range participants {
		if p.ID() == req.ParticipantID {
			targetParticipant = p
			break
		}
	}

	response := common.SetParticipantUnavailableResponse{}
	if targetParticipant == nil {
		response.Success = false
		response.Error = "participant not found"
	} else {
		targetParticipant.SetUnavailable(req.Unavailable)
		response.Success = true
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
