package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/uuid-generator/common"
	"github.com/uuid-generator/uuid"
)

func main() {
	addr := flag.String("addr", ":8080", "server address")
	flag.Parse()

	http.HandleFunc("/generate", handleGenerate)
	http.HandleFunc("/parse", handleParse)

	log.Printf("UUID server starting on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}

	uuids, err := generateUUIDs(req)
	resp := common.GenerateResponse{UUIDs: uuids}
	if err != nil {
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp := common.ParseResponse{}
	u, err := uuid.Parse(req.UUID)
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Version = u.Version()
		resp.Variant = u.Variant()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func generateUUIDs(req common.GenerateRequest) ([]string, error) {
	uuids := make([]string, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		var u uuid.UUID
		var err error

		switch req.Version {
		case 1:
			u, err = uuid.NewV1()
		case 3:
			ns, err := getNamespace(req.Namespace)
			if err != nil {
				return nil, err
			}
			if req.Name == "" {
				return nil, uuid.ErrInvalidUUID
			}
			u = uuid.NewV3(ns, req.Name)
		case 4:
			u, err = uuid.NewV4()
		case 5:
			ns, err := getNamespace(req.Namespace)
			if err != nil {
				return nil, err
			}
			if req.Name == "" {
				return nil, uuid.ErrInvalidUUID
			}
			u = uuid.NewV5(ns, req.Name)
		default:
			return nil, uuid.ErrInvalidVersion
		}

		if err != nil {
			return nil, err
		}

		uuids = append(uuids, u.String())
	}

	return uuids, nil
}

func getNamespace(ns string) (uuid.UUID, error) {
	switch ns {
	case "dns", "DNS":
		return uuid.NamespaceDNS, nil
	case "url", "URL":
		return uuid.NamespaceURL, nil
	case "oid", "OID":
		return uuid.NamespaceOID, nil
	case "x500", "X500":
		return uuid.NamespaceX500, nil
	default:
		return uuid.Parse(ns)
	}
}
