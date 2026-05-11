package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"oauth2-server/pkg/common"
	"oauth2-server/pkg/oauth2"
)

const defaultPort = "8080"

func main() {
	port := flag.String("port", "", "Server port")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("OAUTH2_SERVER_PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = defaultPort
		}
	}

	store := oauth2.NewStore()
	store.RegisterClient(&oauth2.Client{
		ID:           "test-client",
		Secret:       "test-secret",
		RedirectURIs: []string{"http://localhost:8080/callback"},
	})

	service := oauth2.NewService(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/authorize", authorizeHandler(service))
	mux.HandleFunc("/auth/token", tokenHandler(service))
	mux.HandleFunc("/auth/refresh", refreshHandler(service))

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("OAuth2 server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func authorizeHandler(service *oauth2.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "invalid_request", "method not allowed")
			return
		}

		var req common.AuthorizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
			return
		}

		code, err := service.Authorize(
			req.ClientID,
			req.RedirectURI,
			req.ResponseType,
			req.Scope,
			req.CodeChallenge,
			req.CodeChallengeMethod,
		)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), "")
			return
		}

		writeJSON(w, http.StatusOK, &common.AuthorizeResponse{Code: code})
	}
}

func tokenHandler(service *oauth2.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "invalid_request", "method not allowed")
			return
		}

		clientID, clientSecret, err := oauth2.ParseBasicAuth(r.Header.Get("Authorization"))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_client", "invalid Basic auth")
			return
		}

		var req common.TokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
			return
		}

		if req.GrantType != "authorization_code" {
			writeError(w, http.StatusBadRequest, "unsupported_grant_type", "")
			return
		}

		token, err := service.Exchange(clientID, clientSecret, req.Code, req.RedirectURI, req.CodeVerifier)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), "")
			return
		}

		writeJSON(w, http.StatusOK, tokenToResponse(token))
	}
}

func refreshHandler(service *oauth2.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "invalid_request", "method not allowed")
			return
		}

		var req common.RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
			return
		}

		if req.GrantType != "refresh_token" {
			writeError(w, http.StatusBadRequest, "unsupported_grant_type", "")
			return
		}

		token, err := service.Refresh(req.RefreshToken)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), "")
			return
		}

		writeJSON(w, http.StatusOK, tokenToResponse(token))
	}
}

func tokenToResponse(t *oauth2.Token) *common.TokenResponse {
	return &common.TokenResponse{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		ExpiresIn:    int(time.Until(t.AccessExpiry).Seconds()),
		RefreshToken: t.RefreshToken,
		Scope:        t.Scope,
	}
}

func writeError(w http.ResponseWriter, status int, err, desc string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&common.ErrorResponse{Error: err, ErrorDescription: desc})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
