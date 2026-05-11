package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/solocoder/http2-analyzer/api"
	"github.com/solocoder/http2-analyzer/frameparser"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/health", handleHealth)

	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.ParseResponse{Error: "method not allowed, use POST"})
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ParseResponse{Error: fmt.Sprintf("failed to read request body: %v", err)})
		return
	}
	defer r.Body.Close()

	var req api.ParseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ParseResponse{Error: fmt.Sprintf("invalid JSON: %v", err)})
		return
	}

	var data []byte
	switch req.DataKind {
	case "hex", "":
		data, err = hex.DecodeString(req.Data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ParseResponse{Error: fmt.Sprintf("invalid hex string: %v", err)})
			return
		}
	case "file":
		data, err = ioutil.ReadFile(req.Data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ParseResponse{Error: fmt.Sprintf("failed to read file: %v", err)})
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ParseResponse{Error: fmt.Sprintf("invalid data_kind: %s (use 'hex' or 'file')", req.DataKind)})
		return
	}

	parser := frameparser.NewParser()
	if req.Settings != nil {
		if req.Settings.MaxFrameSize > 0 {
			parser.SetMaxFrameSize(req.Settings.MaxFrameSize)
		}
		if req.Settings.HeaderTableSize > 0 {
			parser.SetHeaderTableSize(req.Settings.HeaderTableSize)
		}
	}

	internalFrames, err := parser.Parse(data)
	if err != nil {
		resp := api.ParseResponse{Error: err.Error()}
		if len(internalFrames) > 0 {
			resp.Frames = convertFrames(internalFrames)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.ParseResponse{Frames: convertFrames(internalFrames)}
	json.NewEncoder(w).Encode(resp)
}

func convertFrames(internal []frameparser.ParsedFrameInternal) []api.ParsedFrame {
	result := make([]api.ParsedFrame, len(internal))
	for i, f := range internal {
		result[i] = api.ParsedFrame{
			Offset:   f.Offset,
			Type:     f.Type,
			TypeCode: f.TypeCode,
			Flags:    f.Flags,
			FlagsCode: f.FlagsCode,
			StreamID: f.StreamID,
			Length:   f.Length,
			Payload:  f.Payload,
			RawHex:   f.RawHex,
		}
	}
	return result
}
