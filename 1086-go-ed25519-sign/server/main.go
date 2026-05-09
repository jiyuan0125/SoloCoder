package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/example/ed25519-sign/common"
	"github.com/example/ed25519-sign/cryptolib"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, &common.ErrorResponse{Error: err.Error()})
}

func generateKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	pub, priv, err := cryptolib.GenerateKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, &common.GenerateKeyResponse{
		PublicKey:  pub,
		PrivateKey: priv,
	})
}

func signHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req common.SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	sig, err := cryptolib.Sign(req.PrivateKey, req.Message)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, cryptolib.ErrInvalidBase64) || errors.Is(err, cryptolib.ErrInvalidKeyLength) {
			status = http.StatusBadRequest
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, &common.SignResponse{Signature: sig})
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req common.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	err := cryptolib.Verify(req.PublicKey, req.Message, req.Signature)
	if err != nil {
		if errors.Is(err, cryptolib.ErrInvalidSignature) {
			writeJSON(w, http.StatusOK, &common.VerifyResponse{Valid: false})
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, &common.VerifyResponse{Valid: true})
}

func convertPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req common.ConvertPublicKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	xPub, err := cryptolib.PublicKeyToX25519(req.PublicKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, &common.ConvertPublicKeyResponse{X25519PublicKey: xPub})
}

func convertPrivateKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req common.ConvertPrivateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	xPriv, err := cryptolib.PrivateKeyToX25519(req.PrivateKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, &common.ConvertPrivateKeyResponse{X25519PrivateKey: xPriv})
}

func inspectKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req common.InspectKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	info, err := cryptolib.InspectKey(req.Key)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, &common.InspectKeyResponse{
		Type:        string(info.Type),
		EncodedLen:  info.EncodedLen,
		DecodedLen:  info.DecodedLen,
		Fingerprint: info.Fingerprint,
	})
}

func main() {
	http.HandleFunc("/generate-key", generateKeyHandler)
	http.HandleFunc("/sign", signHandler)
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/convert-public-key", convertPublicKeyHandler)
	http.HandleFunc("/convert-private-key", convertPrivateKeyHandler)
	http.HandleFunc("/inspect-key", inspectKeyHandler)

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
