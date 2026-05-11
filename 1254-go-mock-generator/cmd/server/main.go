package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gomock-generator/pkg/api"
	"gomock-generator/pkg/mockgen"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("GOMOCK_SERVER_PORT")
	}
	if port == "" {
		port = "8501"
	}

	http.HandleFunc("/generate", handleGenerate)
	http.HandleFunc("/parse", handleParse)

	log.Printf("server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	var req api.GenerateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "parse request: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.GenerateResponse{}
	code, err := mockgen.GenerateMock(req.Source, req.Interface)
	if err != nil {
		resp.Err = err.Error()
	} else {
		resp.Code = code
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.ParseResponse{}
	ifaces, err := mockgen.ListInterfaces(string(body))
	if err != nil {
		resp.Err = err.Error()
	} else {
		resp.Interfaces = ifaces
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
