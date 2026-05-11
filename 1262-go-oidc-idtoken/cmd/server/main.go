package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go-oidc-idtoken/pkg/api"
	"go-oidc-idtoken/pkg/idtoken"
)

type Server struct {
	validator *idtoken.Validator
}

func NewServer() *Server {
	nonceCache := idtoken.NewNonceCache()
	validator := idtoken.NewValidator(nil, nonceCache)
	return &Server{
		validator: validator,
	}
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.validator.Validate(req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	errorStrings := make([]string, 0, len(result.Errors))
	for _, err := range result.Errors {
		errorStrings = append(errorStrings, err.Error())
	}

	claims := map[string]interface{}{}
	if result.Claims != nil {
		claims = result.Claims.Raw
	}

	resp := api.VerifyResponse{
		Valid:  result.Valid,
		Claims: claims,
		Errors: errorStrings,
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config := &idtoken.Config{
		Issuer:   req.Issuer,
		ClientID: req.ClientID,
		Secret:   []byte(req.Secret),
	}

	s.validator.UpdateConfig(config)

	resp := api.ConfigResponse{Success: true}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	header, claims, _, _, err := idtoken.Parse(req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.ParseResponse{
		Header: map[string]interface{}{
			"alg": header.Alg,
			"typ": header.Typ,
		},
		Claims: claims.Raw,
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SetConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currentConfig := s.validator.GetConfig()
	if currentConfig == nil {
		currentConfig = &idtoken.Config{}
	}

	key := strings.ToLower(req.Key)
	switch key {
	case "issuer":
		currentConfig.Issuer = req.Value
	case "client_id", "clientid":
		currentConfig.ClientID = req.Value
	case "secret":
		currentConfig.Secret = []byte(req.Value)
	default:
		http.Error(w, "unknown config key: "+req.Key, http.StatusBadRequest)
		return
	}

	s.validator.UpdateConfig(currentConfig)

	resp := api.SetConfigResponse{Success: true}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	var port string

	flag.StringVar(&port, "port", "", "server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
	}

	server := NewServer()

	http.HandleFunc("/oidc/verify", server.handleVerify)
	http.HandleFunc("/oidc/config", server.handleConfig)
	http.HandleFunc("/oidc/parse", server.handleParse)
	http.HandleFunc("/oidc/set-config", server.handleSetConfig)

	addr := ":" + port
	fmt.Printf("Server listening on port %s\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
