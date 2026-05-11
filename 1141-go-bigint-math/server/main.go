package main

import (
	"encoding/json"
	"log"
	"net/http"
	"bigcalc/bigmath"
	"bigcalc/common"
)

func main() {
	http.HandleFunc("/calc", handleCalculate)
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, common.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	resp := executeOperation(req)
	writeResponse(w, resp)
}

func executeOperation(req common.Request) common.Response {
	switch req.Operation {
	case common.OpAdd:
		result, err := bigmath.Add(req.A, req.B)
		if err != nil {
			return common.Response{Success: false, Error: err.Error()}
		}
		return common.Response{Success: true, Result: result}

	case common.OpSubtract:
		result, err := bigmath.Subtract(req.A, req.B)
		if err != nil {
			return common.Response{Success: false, Error: err.Error()}
		}
		return common.Response{Success: true, Result: result}

	case common.OpMultiply:
		result, err := bigmath.Multiply(req.A, req.B)
		if err != nil {
			return common.Response{Success: false, Error: err.Error()}
		}
		return common.Response{Success: true, Result: result}

	case common.OpDivide:
		q, r, err := bigmath.Divide(req.A, req.B)
		if err != nil {
			return common.Response{Success: false, Error: err.Error()}
		}
		return common.Response{
			Success:   true,
			Quotient:  q,
			Remainder: r,
		}

	default:
		return common.Response{
			Success: false,
			Error:   "Unknown operation",
		}
	}
}

func writeResponse(w http.ResponseWriter, resp common.Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
