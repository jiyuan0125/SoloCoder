package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"combinatorics/api"
	"combinatorics/comb"
)

func main() {
	http.HandleFunc("/", handleRequest)
	port := ":8080"
	log.Printf("Combinatorics server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "only POST method is allowed")
		return
	}

	var req api.CombinatoricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	result, err := processRequest(&req)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccessResponse(w, result)
}

func processRequest(req *api.CombinatoricsRequest) (string, error) {
	switch req.Operation {
	case api.OpCombination:
		val, err := comb.Combination(req.N, req.K)
		if err != nil {
			return "", err
		}
		return val.String(), nil
	case api.OpCombinationMod:
		val, err := comb.CombinationMod(req.N, req.K, req.Mod)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(val, 10), nil
	case api.OpPermutation:
		val, err := comb.Permutation(req.N, req.K)
		if err != nil {
			return "", err
		}
		return val.String(), nil
	case api.OpPermutationMod:
		val, err := comb.PermutationMod(req.N, req.K, req.Mod)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(val, 10), nil
	case api.OpPermutationDup:
		val, err := comb.PermutationWithDuplicates(req.Counts)
		if err != nil {
			return "", err
		}
		return val.String(), nil
	case api.OpPermutationDupMod:
		val, err := comb.PermutationWithDuplicatesMod(req.Counts, req.Mod)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(val, 10), nil
	default:
		return "", fmt.Errorf("unknown operation: %s", req.Operation)
	}
}

func writeSuccessResponse(w http.ResponseWriter, result string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.CombinatoricsResponse{
		Success: true,
		Result:  result,
	})
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.CombinatoricsResponse{
		Success: false,
		Error:   message,
	})
}
