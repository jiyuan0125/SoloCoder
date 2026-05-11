package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"csvjson/pkg/api"
	"csvjson/pkg/csvjson"
)

func getPort() string {
	if envPort := os.Getenv("PORT"); envPort != "" {
		return ":" + envPort
	}
	flagPort := flag.String("port", "8105", "server port")
	flag.Parse()
	return ":" + *flagPort
}

func parseDelimiter(d string) csvjson.Delimiter {
	switch strings.ToLower(d) {
	case "tab", "\t", "tsv":
		return csvjson.DelimiterTab
	default:
		return csvjson.DelimiterComma
	}
}

func handleCSVToJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed, use POST")
		return
	}
	var req api.ConvertRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		req.Content = string(body)
	}
	opts := csvjson.DefaultOptions()
	opts.Delimiter = parseDelimiter(req.Delimiter)
	if req.MaxDigits > 0 {
		opts.MaxDigits = req.MaxDigits
	}
	opts.PrettyJSON = req.Pretty
	records, err := csvjson.CSVToJSON(strings.NewReader(req.Content), opts)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var result []byte
	if opts.PrettyJSON {
		result, err = json.MarshalIndent(records, "", "  ")
	} else {
		result, err = json.Marshal(records)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to marshal JSON")
		return
	}
	writeSuccess(w, string(result))
}

func handleJSONToCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed, use POST")
		return
	}
	var req api.ConvertRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		req.Content = string(body)
	}
	opts := csvjson.DefaultOptions()
	opts.Delimiter = parseDelimiter(req.Delimiter)
	if req.MaxDigits > 0 {
		opts.MaxDigits = req.MaxDigits
	}
	var buf bytes.Buffer
	if err := csvjson.JSONToCSV(strings.NewReader(req.Content), &buf, opts); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, buf.String())
}

func writeSuccess(w http.ResponseWriter, content string) {
	resp := api.ConvertResponse{
		Success: true,
		Content: content,
	}
	b, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func writeError(w http.ResponseWriter, status int, message string) {
	resp := api.ConvertResponse{
		Success: false,
		Error:   message,
	}
	b, _ := json.Marshal(resp)
	w.WriteHeader(status)
	w.Write(b)
}

func main() {
	port := getPort()
	mux := http.NewServeMux()
	mux.HandleFunc(api.PathCSVToJSON, handleCSVToJSON)
	mux.HandleFunc(api.PathJSONToCSV, handleJSONToCSV)
	log.Printf("server listening on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
