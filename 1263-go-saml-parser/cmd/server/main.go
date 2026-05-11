package main

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"saml-parser/internal/samlparser"
	"saml-parser/pkg/model"
)

var (
	publicKey *rsa.PublicKey
)

func main() {
	portFlag := flag.String("port", "8080", "HTTP server port")
	publicKeyPath := flag.String("public-key", "", "Path to RSA public key file (PEM format)")
	flag.Parse()

	port := getPort(*portFlag)

	if *publicKeyPath != "" {
		keyData, err := os.ReadFile(*publicKeyPath)
		if err != nil {
			fmt.Printf("Warning: Failed to read public key: %v\n", err)
		} else {
			pemStr := strings.TrimSpace(string(keyData))
			pemStr = strings.ReplaceAll(pemStr, "-----BEGIN PUBLIC KEY-----", "")
			pemStr = strings.ReplaceAll(pemStr, "-----END PUBLIC KEY-----", "")
			pemStr = strings.ReplaceAll(pemStr, "\n", "")
			pemStr = strings.ReplaceAll(pemStr, "\r", "")
			
			pk, err := samlparser.ParseRSAPublicKeyFromPEM(pemStr)
			if err != nil {
				fmt.Printf("Warning: Failed to parse public key: %v\n", err)
			} else {
				publicKey = pk
				fmt.Println("Public key loaded successfully")
			}
		}
	}

	http.HandleFunc("/saml/parse", handleParse)
	http.HandleFunc("/saml/validate", handleValidate)

	fmt.Printf("SAML Parser Server listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func getPort(defaultPort string) string {
	if envPort := os.Getenv("SAML_PORT"); envPort != "" {
		return envPort
	}
	return defaultPort
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req model.ParseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, "Invalid JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, _ := samlparser.Parse(req.XML)
	writeJSONResponse(w, resp)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req model.ValidateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONError(w, "Invalid JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}

	options := samlparser.ValidationOptions{
		Audience:  req.Audience,
		Recipient: req.Recipient,
		PublicKey: publicKey,
	}

	resp, _ := samlparser.Validate(req.XML, options)
	writeJSONResponse(w, resp)
}

func writeJSONResponse(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
