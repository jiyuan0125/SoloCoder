package common

type IssueTokenRequest struct {
	UserID string            `json:"user_id"`
	Claims map[string]any    `json:"claims,omitempty"`
}

type IssueTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	AccessTokenID    string `json:"access_token_id"`
	RefreshTokenID   string `json:"refresh_token_id"`
	AccessTokenExp   int64  `json:"access_token_exp"`
	RefreshTokenExp  int64  `json:"refresh_token_exp"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	AccessTokenID    string `json:"access_token_id"`
	RefreshTokenID   string `json:"refresh_token_id"`
	AccessTokenExp   int64  `json:"access_token_exp"`
	RefreshTokenExp  int64  `json:"refresh_token_exp"`
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type TokenClaimsData struct {
	TokenID   string         `json:"token_id"`
	UserID    string         `json:"user_id"`
	TokenType string         `json:"token_type"`
	Claims    map[string]any `json:"claims,omitempty"`
	Issuer    string         `json:"issuer"`
	IssuedAt  int64          `json:"issued_at"`
	ExpiresAt int64          `json:"expires_at"`
}

type ValidateTokenResponse struct {
	Valid   bool              `json:"valid"`
	Reason  string            `json:"reason,omitempty"`
	Claims  *TokenClaimsData  `json:"claims,omitempty"`
}

type RevokeTokenRequest struct {
	TokenID string `json:"token_id"`
}

type RevokeTokenResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

const (
	EndpointIssue   = "/api/tokens/issue"
	EndpointRefresh = "/api/tokens/refresh"
	EndpointValidate = "/api/tokens/validate"
	EndpointRevoke  = "/api/tokens/revoke"
)
