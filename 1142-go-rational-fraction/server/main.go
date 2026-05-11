package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"fractional-service/api"
	"fractional-service/fraction"
)

func main() {
	http.HandleFunc("/", handleOperation)

	fmt.Println("Starting fractional service on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleOperation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	result, err := executeOperation(&req)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.OperationResponse{
		Success: true,
		Result:  result,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func executeOperation(req *api.OperationRequest) (*api.FractionResult, error) {
	var frac *fraction.Fraction
	var err error

	switch req.Operation {
	case api.OperationAdd, api.OperationSub, api.OperationMul, api.OperationDiv:
		frac, err = executeBinaryOperation(req)
	case api.OperationDecimal:
		return executeDecimalOperation(req)
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}

	if err != nil {
		return nil, err
	}

	return buildFractionResult(frac), nil
}

func executeBinaryOperation(req *api.OperationRequest) (*fraction.Fraction, error) {
	f1, err := fraction.Parse(req.Operand1)
	if err != nil {
		return nil, fmt.Errorf("invalid operand1: %v", err)
	}

	f2, err := fraction.Parse(req.Operand2)
	if err != nil {
		return nil, fmt.Errorf("invalid operand2: %v", err)
	}

	switch req.Operation {
	case api.OperationAdd:
		return f1.Add(f2)
	case api.OperationSub:
		return f1.Sub(f2)
	case api.OperationMul:
		return f1.Mul(f2)
	case api.OperationDiv:
		return f1.Div(f2)
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}
}

func executeDecimalOperation(req *api.OperationRequest) (*api.FractionResult, error) {
	f, err := fraction.Parse(req.Operand1)
	if err != nil {
		return nil, fmt.Errorf("invalid operand1: %v", err)
	}

	precision := req.Precision
	if precision <= 0 {
		precision = 10
	}

	decimal, err := f.ToDecimal(precision)
	if err != nil {
		return nil, err
	}

	result := buildFractionResult(f)
	result.Decimal = decimal
	return result, nil
}

func buildFractionResult(f *fraction.Fraction) *api.FractionResult {
	result := &api.FractionResult{
		Fraction: f.String(),
	}

	if f.IsInteger() {
		result.Integer = strconv.FormatInt(f.Numerator, 10)
	}

	if mixed, ok := f.MixedString(); ok {
		result.Mixed = mixed
	}

	return result
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.OperationResponse{
		Success: false,
		Error:   message,
	})
}

func parsePathParams(path string) (api.Operation, string, string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 1 {
		return "", "", "", fmt.Errorf("invalid path")
	}

	op := api.Operation(parts[0])
	operand1 := ""
	operand2 := ""

	if len(parts) >= 2 {
		operand1 = parts[1]
	}

	if len(parts) >= 3 {
		operand2 = parts[2]
	}

	return op, operand1, operand2, nil
}
