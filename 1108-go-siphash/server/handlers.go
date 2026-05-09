package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/siphash-service/common"
	"github.com/siphash-service/siphash"
)

func handleHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.HashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var hash uint64
	variant := req.Variant
	if variant == "" {
		variant = common.VariantSipHash24
	}

	switch variant {
	case common.VariantSipHash24:
		hash = siphash.Sum64WithRounds(globalKey, req.Data, 2, 4)
	case common.VariantSipHash13:
		hash = siphash.Sum64WithRounds(globalKey, req.Data, 1, 3)
	default:
		writeError(w, http.StatusBadRequest, "Invalid variant")
		return
	}

	resp := common.HashResponse{
		Hash:    hash,
		Hex:     fmt.Sprintf("%016x", hash),
		Variant: variant,
	}

	writeJSON(w, http.StatusOK, resp)
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Key) == 0 {
		writeError(w, http.StatusBadRequest, "Key cannot be empty")
		return
	}

	globalMap.Put(req.Key, req.Value)

	writeJSON(w, http.StatusOK, common.PutResponse{OK: true})
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	value, found := globalMap.Get(req.Key)

	writeJSON(w, http.StatusOK, common.GetResponse{
		Value: value,
		Found: found,
	})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	deleted := globalMap.Delete(req.Key)

	writeJSON(w, http.StatusOK, common.DeleteResponse{OK: deleted})
}

func handleBatchPut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.BatchPutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	success := 0
	for _, item := range req.Items {
		if len(item.Key) > 0 {
			globalMap.Put(item.Key, item.Value)
			success++
		}
	}

	writeJSON(w, http.StatusOK, common.BatchPutResponse{Success: success})
}

func handleBatchGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.BatchGetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	items := make([]common.BatchGetItem, 0, len(req.Keys))
	for _, key := range req.Keys {
		value, found := globalMap.Get(key)
		items = append(items, common.BatchGetItem{
			Key:   key,
			Value: value,
			Found: found,
		})
	}

	writeJSON(w, http.StatusOK, common.BatchGetResponse{Items: items})
}

func handleBatchDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.BatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	deleted := 0
	for _, key := range req.Keys {
		if globalMap.Delete(key) {
			deleted++
		}
	}

	writeJSON(w, http.StatusOK, common.BatchDeleteResponse{Deleted: deleted})
}

func handleRotateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.RotateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	newKey, err := siphash.NewKeyFromBytes(req.NewKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	globalMap.RotateKey(newKey)
	globalKey = newKey

	writeJSON(w, http.StatusOK, common.RotateKeyResponse{OK: true})
}

func handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	keys := globalMap.Keys()
	writeJSON(w, http.StatusOK, common.ListResponse{
		Keys:  keys,
		Count: len(keys),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, common.ErrorResponse{Error: message})
}
