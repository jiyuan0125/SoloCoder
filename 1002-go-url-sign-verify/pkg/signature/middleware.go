package signature

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type Middleware struct {
	verifier *Verifier
	config   *VerifierConfig
}

func NewMiddleware(config *VerifierConfig) *Middleware {
	verifier := NewVerifier(config)
	return &Middleware{
		verifier: verifier,
		config:   config,
	}
}

func (m *Middleware) extractParams(r *http.Request) (map[string]string, error) {
	params := make(map[string]string)
	
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	
	for key, values := range r.Form {
		if _, exists := params[key]; !exists && len(values) > 0 {
			params[key] = values[0]
		}
	}
	
	return params, nil
}

func (m *Middleware) extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

func (m *Middleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appKey := r.Header.Get(m.config.AppKeyHeaderName)
		
		params, err := m.extractParams(r)
		if err != nil {
			http.Error(w, "failed to parse request parameters", http.StatusBadRequest)
			return
		}
		
		headers := m.extractHeaders(r)
		path := r.URL.Path
		
		result := m.verifier.Verify(appKey, path, params, headers)
		
		if result.DebugEnabled && !result.Success {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			debugResponse := map[string]interface{}{
				"success":     false,
				"error":       result.Error,
				"sign_string": result.SignString,
				"signature":   result.Signature,
				"client_sign": result.ClientSign,
				"debug":       true,
			}
			json.NewEncoder(w).Encode(debugResponse)
			return
		}
		
		if !result.Success {
			http.Error(w, result.Error, http.StatusUnauthorized)
			return
		}
		
		if result.DebugEnabled {
			r = r.WithContext(setDebugInfo(r.Context(), result))
		}
		
		next.ServeHTTP(w, r)
	}
}

type contextKey string

const debugInfoKey contextKey = "debug_info"

type DebugInfo struct {
	SignString string
	Signature  string
	ClientSign string
}

func setDebugInfo(ctx context.Context, result *VerificationResult) context.Context {
	debugInfo := &DebugInfo{
		SignString: result.SignString,
		Signature:  result.Signature,
		ClientSign: result.ClientSign,
	}
	return context.WithValue(ctx, debugInfoKey, debugInfo)
}

func GetDebugInfo(ctx context.Context) *DebugInfo {
	if info, ok := ctx.Value(debugInfoKey).(*DebugInfo); ok {
		return info
	}
	return nil
}

func URLValuesToMap(values url.Values) map[string]string {
	result := make(map[string]string)
	for key, vals := range values {
		if len(vals) > 0 {
			result[key] = vals[0]
		}
	}
	return result
}
