package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"tls-simulator/common"
	"tls-simulator/tls"
)

type Server struct {
	sessionCache map[string]struct{}
	mu           sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		sessionCache: make(map[string]struct{}),
	}
}

func (s *Server) AddSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionCache[sessionID] = struct{}{}
}

func (s *Server) getSessionCache() map[string]struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cache := make(map[string]struct{})
	for k, v := range s.sessionCache {
		cache[k] = v
	}
	return cache
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, errMsg string) {
	writeJSON(w, status, common.APIError{
		Error:   http.StatusText(status),
		Code:    status,
		Message: errMsg,
	})
}

func (s *Server) handleCipherSuites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	cipherSuites := tls.GetAllCipherSuites()
	response := common.CipherSuitesResponse{
		CipherSuites: make([]common.CipherSuiteInfo, 0, len(cipherSuites)),
	}
	for _, cs := range cipherSuites {
		response.CipherSuites = append(response.CipherSuites, common.CipherSuiteInfo{
			ID:          fmt.Sprintf("0x%04X", cs.ID),
			Name:        cs.Name,
			KeyExchange: cs.KeyExchange,
			Encryption:  cs.Encryption,
			MAC:         cs.MAC,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleHandshake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.HandshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}
	defer r.Body.Close()
	config := tls.DefaultSimulationConfig()
	if len(req.ClientCipherSuites) > 0 {
		clientSuites, err := tls.ParseCipherSuiteIDs(req.ClientCipherSuites)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid client cipher suites: %v", err))
			return
		}
		config.ClientCipherSuites = clientSuites
	}
	if len(req.ServerCipherSuites) > 0 {
		serverSuites, err := tls.ParseCipherSuiteIDs(req.ServerCipherSuites)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid server cipher suites: %v", err))
			return
		}
		config.ServerCipherSuites = serverSuites
	}
	extensions := []tls.Extension{}
	if req.SNI != "" {
		extensions = append(extensions, &tls.SNIExtension{HostName: req.SNI})
	} else {
		extensions = append(extensions, &tls.SNIExtension{HostName: "example.com"})
	}
	extensions = append(extensions, &tls.SupportedGroupsExtension{
		Groups: []tls.NamedGroup{tls.GroupX25519, tls.GroupSecP256R1, tls.GroupSecP384R1},
	})
	extensions = append(extensions, &tls.SignatureAlgorithmsExtension{
		Algorithms: []tls.SignatureAlgorithm{
			tls.SignatureRSA_PSS_RSAE_SHA256,
			tls.SignatureRSA_PKCS1_SHA256,
			tls.SignatureECDSA_P256_SHA256,
		},
	})
	if len(req.ALPNProtocols) > 0 {
		extensions = append(extensions, &tls.ALPNExtension{Protocols: req.ALPNProtocols})
	} else {
		extensions = append(extensions, &tls.ALPNExtension{Protocols: []string{"h2", "http/1.1"}})
	}
	config.ClientExtensions = extensions
	if req.ResumeSessionID != "" {
		sessionID, err := hex.DecodeString(req.ResumeSessionID)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid resume session ID: %v", err))
			return
		}
		config.ResumeSessionID = tls.SessionID(sessionID)
		s.AddSession(string(sessionID))
	}
	config.SessionCache = s.getSessionCache()
	result, _ := tls.RunHandshakeSimulation(config)
	response := common.HandshakeResponse{
		Success:          result.Success,
		IsSessionResume:  result.IsSessionResume,
		Error:            result.Error,
		ClientRandom:     result.ClientRandom,
		ServerRandom:     result.ServerRandom,
		SessionID:        result.SessionID,
		ClientExtensions: result.ClientExtensions,
		ServerExtensions: result.ServerExtensions,
		Steps:            result.StepsText,
		FullMessageDump:  result.FullMessageDump,
	}
	if result.NegotiatedSuite != nil {
		response.NegotiatedSuite = &common.CipherSuiteInfo{
			ID:          result.NegotiatedSuite.ID,
			Name:        result.NegotiatedSuite.Name,
			KeyExchange: result.NegotiatedSuite.KeyExchange,
			Encryption:  result.NegotiatedSuite.Encryption,
			MAC:         result.NegotiatedSuite.MAC,
		}
	}
	if result.Success && result.SessionID != "" {
		sessionBytes, err := hex.DecodeString(result.SessionID)
		if err == nil {
			s.AddSession(string(sessionBytes))
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	port := flag.String("port", "8700", "HTTP server port")
	flag.Parse()
	server := NewServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/cipher-suites", server.handleCipherSuites)
	mux.HandleFunc("/handshake", server.handleHandshake)
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("TLS Handshake Simulator Server starting on %s", addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health          - Health check")
	log.Printf("  GET  /cipher-suites   - List supported cipher suites")
	log.Printf("  POST /handshake       - Simulate TLS handshake")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
