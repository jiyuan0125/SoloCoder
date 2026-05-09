package common

const (
	CurveP256 = "P-256"
	CurveP384 = "P-384"
	CurveP521 = "P-521"
)

var SupportedCurves = []string{CurveP256, CurveP384, CurveP521}

type StartNegotiationRequest struct {
	Curve string `json:"curve"`
}

type StartNegotiationResponse struct {
	SessionID string `json:"session_id"`
	Curve     string `json:"curve"`
	PublicKey string `json:"public_key"`
}

type CompleteNegotiationRequest struct {
	SessionID string `json:"session_id"`
	Curve     string `json:"curve"`
	PublicKey string `json:"public_key"`
}

type CompleteNegotiationResponse struct {
	SessionID    string `json:"session_id"`
	SharedSecret string `json:"shared_secret"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
