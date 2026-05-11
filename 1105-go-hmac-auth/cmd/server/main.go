package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sort"

	"hmac-auth/pkg/hmacauth"
	"hmac-auth/pkg/types"
)

var keyStore *hmacauth.KeyStore

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func respondError(w http.ResponseWriter, status int, errMsg string) {
	respondJSON(w, status, types.ErrorResponse{Error: errMsg})
}

func signHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req types.SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Method == "" || req.Path == "" {
		respondError(w, http.StatusBadRequest, "method and path are required")
		return
	}

	if req.Timestamp == 0 {
		respondError(w, http.StatusBadRequest, "timestamp is required")
		return
	}

	key, version := keyStore.GetCurrentKey()
	signature := hmacauth.Sign(key, req.Method, req.Path, req.Timestamp, []byte(req.Body))

	respondJSON(w, http.StatusOK, types.SignResponse{
		Signature: signature,
		Version:   version,
	})
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req types.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Method == "" || req.Path == "" || req.Signature == "" {
		respondError(w, http.StatusBadRequest, "method, path and signature are required")
		return
	}

	if req.Timestamp == 0 {
		respondError(w, http.StatusBadRequest, "timestamp is required")
		return
	}

	valid := false
	message := "signature invalid"

	if req.Version != nil {
		keys := keyStore.GetAllKeys()
		key, exists := keys[*req.Version]
		if exists {
			valid = hmacauth.Verify(key, req.Method, req.Path, req.Timestamp, []byte(req.Body), req.Signature)
			if valid {
				message = "signature valid"
			}
		}
	} else {
		valid = hmacauth.VerifyWithKeyStore(keyStore, req.Method, req.Path, req.Timestamp, []byte(req.Body), req.Signature)
		if valid {
			message = "signature valid"
		}
	}

	respondJSON(w, http.StatusOK, types.VerifyResponse{
		Valid:   valid,
		Message: message,
	})
}

func rotateKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	_, oldVer := keyStore.GetCurrentKey()
	if err := keyStore.RotateKey(); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to rotate key")
		return
	}
	_, newVer := keyStore.GetCurrentKey()

	respondJSON(w, http.StatusOK, types.RotateKeyResponse{
		Success:         true,
		NewVersion:      newVer,
		PreviousVersion: oldVer,
	})
}

func keyListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	keys := keyStore.GetAllKeys()
	versions := make([]int, 0, len(keys))
	for v := range keys {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	currentVer := 0
	if len(versions) > 0 {
		currentVer = versions[len(versions)-1]
	}

	respondJSON(w, http.StatusOK, types.KeyListResponse{
		CurrentVersion: currentVer,
		Versions:       versions,
	})
}

func main() {
	addr := flag.String("addr", ":8203", "server address")
	flag.Parse()

	var err error
	keyStore, err = hmacauth.NewKeyStore()
	if err != nil {
		log.Fatalf("failed to create key store: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/sign", signHandler)
	mux.HandleFunc("/verify", verifyHandler)
	mux.HandleFunc("/keys/rotate", rotateKeyHandler)
	mux.HandleFunc("/keys", keyListHandler)

	log.Printf("server starting on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
