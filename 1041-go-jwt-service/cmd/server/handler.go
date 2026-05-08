package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"jwt-service/pkg/common"
	"jwt-service/pkg/jwtcore"
)

type HTTPHandler struct {
	service   *jwtcore.Service
	blacklist *jwtcore.Blacklist
}

func NewHTTPHandler(service *jwtcore.Service, blacklist *jwtcore.Blacklist) *HTTPHandler {
	return &HTTPHandler{
		service:   service,
		blacklist: blacklist,
	}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(common.EndpointIssue, h.handleIssue)
	mux.HandleFunc(common.EndpointRefresh, h.handleRefresh)
	mux.HandleFunc(common.EndpointValidate, h.handleValidate)
	mux.HandleFunc(common.EndpointRevoke, h.handleRevoke)
}

func (h *HTTPHandler) handleIssue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST method is allowed")
		return
	}

	var req common.IssueTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.UserID == "" {
		h.writeError(w, http.StatusBadRequest, "missing_user_id", "user_id is required")
		return
	}

	pair, err := h.service.IssueTokenPair(req.UserID, req.Claims)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "failed to issue tokens")
		return
	}

	resp := common.IssueTokenResponse{
		AccessToken:     pair.AccessToken,
		RefreshToken:    pair.RefreshToken,
		AccessTokenID:   pair.AccessClaims.TokenID,
		RefreshTokenID:  pair.RefreshClaims.TokenID,
		AccessTokenExp:  pair.AccessClaims.ExpiresAt.Unix(),
		RefreshTokenExp: pair.RefreshClaims.ExpiresAt.Unix(),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST method is allowed")
		return
	}

	var req common.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.RefreshToken == "" {
		h.writeError(w, http.StatusBadRequest, "missing_refresh_token", "refresh_token is required")
		return
	}

	pair, err := h.service.RefreshTokenPair(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, jwtcore.ErrTokenRevoked):
			h.writeError(w, http.StatusUnauthorized, "token_compromised", err.Error())
		case errors.Is(err, jwtcore.ErrTokenExpired):
			h.writeError(w, http.StatusUnauthorized, "token_expired", "refresh token has expired")
		case errors.Is(err, jwtcore.ErrInvalidTokenType):
			h.writeError(w, http.StatusBadRequest, "invalid_token_type", "expected a refresh token")
		case errors.Is(err, jwtcore.ErrInvalidToken):
			h.writeError(w, http.StatusUnauthorized, "invalid_token", "refresh token is invalid")
		default:
			h.writeError(w, http.StatusInternalServerError, "internal_error", "failed to refresh tokens")
		}
		return
	}

	resp := common.RefreshTokenResponse{
		AccessToken:     pair.AccessToken,
		RefreshToken:    pair.RefreshToken,
		AccessTokenID:   pair.AccessClaims.TokenID,
		RefreshTokenID:  pair.RefreshClaims.TokenID,
		AccessTokenExp:  pair.AccessClaims.ExpiresAt.Unix(),
		RefreshTokenExp: pair.RefreshClaims.ExpiresAt.Unix(),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST method is allowed")
		return
	}

	var req common.ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Token == "" {
		h.writeError(w, http.StatusBadRequest, "missing_token", "token is required")
		return
	}

	claims, err := h.service.ValidateToken(req.Token, jwtcore.AccessToken)
	if err != nil {
		var reason string
		switch {
		case errors.Is(err, jwtcore.ErrTokenRevoked):
			reason = "token_revoked"
		case errors.Is(err, jwtcore.ErrTokenExpired):
			reason = "token_expired"
		case errors.Is(err, jwtcore.ErrInvalidTokenType):
			claims, err = h.service.ValidateToken(req.Token, jwtcore.RefreshToken)
			if err != nil {
				h.writeJSON(w, http.StatusOK, common.ValidateTokenResponse{
					Valid:  false,
					Reason: "invalid_token",
				})
				return
			}
		default:
			h.writeJSON(w, http.StatusOK, common.ValidateTokenResponse{
				Valid:  false,
				Reason: "invalid_token",
			})
			return
		}

		if reason != "" {
			h.writeJSON(w, http.StatusOK, common.ValidateTokenResponse{
				Valid:  false,
				Reason: reason,
			})
			return
		}
	}

	if claims == nil {
		h.writeJSON(w, http.StatusOK, common.ValidateTokenResponse{
			Valid:  false,
			Reason: "invalid_token",
		})
		return
	}

	if h.blacklist.IsUserRevoked(claims.UserID) {
		h.writeJSON(w, http.StatusOK, common.ValidateTokenResponse{
			Valid:  false,
			Reason: "user_revoked",
		})
		return
	}

	resp := common.ValidateTokenResponse{
		Valid: true,
		Claims: &common.TokenClaimsData{
			TokenID:   claims.TokenID,
			UserID:    claims.UserID,
			TokenType: string(claims.TokenType),
			Claims:    claims.Claims,
			Issuer:    claims.Issuer,
			IssuedAt:  claims.IssuedAt.Unix(),
			ExpiresAt: claims.ExpiresAt.Unix(),
		},
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *HTTPHandler) handleRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST method is allowed")
		return
	}

	var req common.RevokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.TokenID == "" {
		h.writeError(w, http.StatusBadRequest, "missing_token_id", "token_id is required")
		return
	}

	h.blacklist.Revoke(req.TokenID, time.Now().Add(30*24*time.Hour))

	h.writeJSON(w, http.StatusOK, common.RevokeTokenResponse{
		Success: true,
		Message: "token revoked",
	})
}

func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, common.ErrorResponse{
		Error:   code,
		Message: message,
	})
}
