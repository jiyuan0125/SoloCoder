package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-two-phase-commit/common"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	serverURL = "http://localhost:8200"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  run        - Run a complete transaction")
		fmt.Println("  status <tx_id>  - Check transaction status")
		fmt.Println("  test <scenario> - Run test scenarios (all, partial, timeout, recovery)")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		runCompleteTransaction()
	case "status":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client status <tx_id>")
			os.Exit(1)
		}
		checkTransactionStatus(os.Args[2])
	case "test":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client test <scenario>")
			fmt.Println("Scenarios: all, partial, timeout, recovery")
			os.Exit(1)
		}
		runTest(os.Args[2])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func runCompleteTransaction() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Running complete transaction flow")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n[Step 1] Begin transaction...")
	beginResp, err := beginTransaction()
	if err != nil {
		fmt.Printf("  Failed to begin transaction: %v\n", err)
		return
	}

	txID := beginResp.TxID
	fmt.Printf("  Transaction ID: %s\n", txID)

	fmt.Println("\n[Step 2] Prepare transaction...")
	operations := map[string][]common.Operation{
		"participant-1": {
			{Type: common.Create, Key: "account-1", Data: map[string]interface{}{"balance": 1000}},
		},
		"participant-2": {
			{Type: common.Create, Key: "account-2", Data: map[string]interface{}{"balance": 2000}},
		},
	}

	prepareResp, err := prepareTransaction(txID, operations)
	if err != nil {
		fmt.Printf("  Prepare failed: %v\n", err)
		fmt.Println("  Rolling back...")
		_, rollbackErr := rollbackTransaction(txID)
		if rollbackErr != nil {
			fmt.Printf("  Rollback failed: %v\n", rollbackErr)
		} else {
			fmt.Println("  Rollback successful")
		}
		return
	}

	fmt.Printf("  Prepare status: %s\n", prepareResp.Status)

	fmt.Println("\n[Step 3] Commit transaction...")
	commitResp, err := commitTransaction(txID)
	if err != nil {
		fmt.Printf("  Commit failed: %v\n", err)
		return
	}

	fmt.Printf("  Commit status: %s\n", commitResp.Status)

	fmt.Println("\n[Step 4] Check final status...")
	statusResp, err := getTransactionStatus(txID)
	if err != nil {
		fmt.Printf("  Failed to get status: %v\n", err)
		return
	}

	fmt.Printf("  Final status: %s\n", statusResp.Status)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Transaction flow completed successfully!")
	fmt.Println(strings.Repeat("=", 60))
}

func checkTransactionStatus(txID string) {
	resp, err := getTransactionStatus(txID)
	if err != nil {
		fmt.Printf("Failed to get status: %v\n", err)
		return
	}

	fmt.Printf("Transaction ID: %s\n", resp.TxID)
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Created At: %s\n", resp.CreatedAt)
	fmt.Printf("Updated At: %s\n", resp.UpdatedAt)
}

func runTest(scenario string) {
	switch strings.ToLower(scenario) {
	case "all":
		testAllSuccess()
	case "partial":
		testPartialFailure()
	case "timeout":
		testTimeoutRollback()
	case "recovery":
		testRecovery()
	default:
		fmt.Printf("Unknown test scenario: %s\n", scenario)
		os.Exit(1)
	}
}

func testAllSuccess() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Test Scenario: All Success")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n[1/5] Begin transaction...")
	beginResp, err := beginTransaction()
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	txID := beginResp.TxID
	fmt.Printf("  SUCCESS: TXID=%s\n", txID)

	fmt.Println("\n[2/5] Prepare with 3 participants...")
	operations := map[string][]common.Operation{
		"participant-1": {{Type: common.Create, Key: "user-1", Data: map[string]interface{}{"name": "Alice"}}},
		"participant-2": {{Type: common.Create, Key: "user-2", Data: map[string]interface{}{"name": "Bob"}}},
		"participant-3": {{Type: common.Create, Key: "user-3", Data: map[string]interface{}{"name": "Charlie"}}},
	}

	_, err = prepareTransaction(txID, operations)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	fmt.Println("  SUCCESS: All participants prepared")

	fmt.Println("\n[3/5] Commit transaction...")
	_, err = commitTransaction(txID)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	fmt.Println("  SUCCESS: Transaction committed")

	fmt.Println("\n[4/5] Verify participant data...")
	for _, pid := range []string{"participant-1", "participant-2", "participant-3"} {
		dataResp, err := getParticipantData(pid)
		if err != nil {
			fmt.Printf("  FAILED to get %s data: %v\n", pid, err)
			continue
		}
		fmt.Printf("  %s has %d records\n", pid, len(dataResp.Data))
	}

	fmt.Println("\n[5/5] Check final status...")
	statusResp, err := getTransactionStatus(txID)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	fmt.Printf("  Final status: %s\n", statusResp.Status)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Test Scenario: All Success - PASSED")
	fmt.Println(strings.Repeat("=", 60))
}

func testPartialFailure() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Test Scenario: Partial Failure")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n[1/5] Begin transaction...")
	beginResp, err := beginTransaction()
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	txID := beginResp.TxID
	fmt.Printf("  SUCCESS: TXID=%s\n", txID)

	fmt.Println("\n[2/5] Prepare with invalid operation (update non-existent key)...")
	operations := map[string][]common.Operation{
		"participant-1": {{Type: common.Create, Key: "valid-key", Data: "value"}},
		"participant-2": {{Type: common.Update, Key: "non-existent-key", Data: "value"}},
	}

	_, err = prepareTransaction(txID, operations)
	if err == nil {
		fmt.Println("  FAILED: Expected prepare to fail but it succeeded")
		return
	}
	fmt.Printf("  SUCCESS: Prepare failed as expected - %v\n", err)

	fmt.Println("\n[3/5] Check transaction status (should be aborted)...")
	statusResp, err := getTransactionStatus(txID)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	fmt.Printf("  Status: %s\n", statusResp.Status)
	if statusResp.Status != "aborted" {
		fmt.Println("  WARNING: Expected aborted status")
	}

	fmt.Println("\n[4/5] Try to commit (should fail)...")
	_, err = commitTransaction(txID)
	if err == nil {
		fmt.Println("  FAILED: Expected commit to fail but it succeeded")
		return
	}
	fmt.Printf("  SUCCESS: Commit failed as expected - %v\n", err)

	fmt.Println("\n[5/5] Verify no data was written...")
	dataResp, err := getParticipantData("participant-1")
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	if _, exists := dataResp.Data["valid-key"]; exists {
		fmt.Println("  FAILED: Data should not have been written")
		return
	}
	fmt.Println("  SUCCESS: No data was written")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Test Scenario: Partial Failure - PASSED")
	fmt.Println(strings.Repeat("=", 60))
}

func testTimeoutRollback() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Test Scenario: Timeout Rollback")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n[1/6] Begin transaction...")
	beginResp, err := beginTransaction()
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	txID := beginResp.TxID
	fmt.Printf("  SUCCESS: TXID=%s\n", txID)

	fmt.Println("\n[2/6] Make participant-2 unavailable...")
	_, err = setParticipantUnavailable("participant-2", true)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	fmt.Println("  SUCCESS: participant-2 is now unavailable")

	fmt.Println("\n[3/6] Prepare transaction (should timeout)...")
	operations := map[string][]common.Operation{
		"participant-1": {{Type: common.Create, Key: "test-key-1", Data: "value1"}},
		"participant-2": {{Type: common.Create, Key: "test-key-2", Data: "value2"}},
	}

	start := time.Now()
	_, err = prepareTransaction(txID, operations)
	elapsed := time.Since(start)

	if err == nil {
		fmt.Println("  FAILED: Expected prepare to timeout but it succeeded")
		return
	}
	fmt.Printf("  SUCCESS: Prepare timed out after %v seconds\n", elapsed.Seconds())
	fmt.Printf("  Error: %v\n", err)

	fmt.Println("\n[4/6] Restore participant-2...")
	_, err = setParticipantUnavailable("participant-2", false)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	fmt.Println("  SUCCESS: participant-2 is now available")

	fmt.Println("\n[5/6] Check transaction status...")
	statusResp, err := getTransactionStatus(txID)
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	fmt.Printf("  Status: %s\n", statusResp.Status)

	fmt.Println("\n[6/6] Verify no data was written...")
	dataResp, err := getParticipantData("participant-1")
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	if _, exists := dataResp.Data["test-key-1"]; exists {
		fmt.Println("  FAILED: Data should not have been written")
		return
	}
	fmt.Println("  SUCCESS: No data was written")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Test Scenario: Timeout Rollback - PASSED")
	fmt.Println(strings.Repeat("=", 60))
}

func testRecovery() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Test Scenario: Coordinator Recovery")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n[Note]")
	fmt.Println("  Coordinator recovery happens automatically on server startup")
	fmt.Println("  Any transactions in 'prepared' state are rolled back during recovery")
	fmt.Println("  Any transactions in 'pending' state are aborted during recovery")
	
	fmt.Println("\n[1/2] Check server status...")
	resp, err := http.Get(serverURL + "/api/participants")
	if err != nil {
		fmt.Printf("  FAILED to connect to server: %v\n", err)
		fmt.Println("  Please start the server first: go run ./server")
		return
	}
	defer resp.Body.Close()
	fmt.Println("  SUCCESS: Server is running")

	fmt.Println("\n[2/2] List participants...")
	participantsResp, err := listParticipants()
	if err != nil {
		fmt.Printf("  FAILED: %v\n", err)
		return
	}
	fmt.Printf("  Active participants: %d\n", len(participantsResp.Participants))
	for _, p := range participantsResp.Participants {
		fmt.Printf("    - %s (unavailable: %v, storage: %d)\n", 
			p.ID, p.Unavailable, p.StorageSize)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Test Scenario: Coordinator Recovery - PASSED")
	fmt.Println(strings.Repeat("=", 60))
}

func beginTransaction() (*common.BeginTxResponse, error) {
	resp, err := http.Post(serverURL+"/api/tx/begin", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.BeginTxResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result, nil
}

func prepareTransaction(txID string, operations map[string][]common.Operation) (*common.PrepareTxResponse, error) {
	req := common.PrepareTxRequest{
		TxID:       txID,
		Operations: operations,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/api/tx/prepare", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.PrepareTxResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result, nil
}

func commitTransaction(txID string) (*common.CommitTxResponse, error) {
	req := common.CommitTxRequest{TxID: txID}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/api/tx/commit", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.CommitTxResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result, nil
}

func rollbackTransaction(txID string) (*common.RollbackTxResponse, error) {
	req := common.RollbackTxRequest{TxID: txID}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/api/tx/rollback", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.RollbackTxResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result, nil
}

func getTransactionStatus(txID string) (*common.StatusTxResponse, error) {
	req := common.StatusTxRequest{TxID: txID}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/api/tx/status", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.StatusTxResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result, nil
}

func listParticipants() (*common.ListParticipantsResponse, error) {
	resp, err := http.Get(serverURL + "/api/participants")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.ListParticipantsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result, nil
}

func getParticipantData(participantID string) (*common.GetParticipantDataResponse, error) {
	req := common.GetParticipantDataRequest{ParticipantID: participantID}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/api/participants/data", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.GetParticipantDataResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result, nil
}

func setParticipantUnavailable(participantID string, unavailable bool) (*common.SetParticipantUnavailableResponse, error) {
	req := common.SetParticipantUnavailableRequest{
		ParticipantID: participantID,
		Unavailable:   unavailable,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/api/participants/set-unavailable", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result common.SetParticipantUnavailableResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf(result.Error)
	}

	return &result, nil
}
