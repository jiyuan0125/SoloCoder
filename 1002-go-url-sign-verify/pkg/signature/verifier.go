package signature

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type VerificationResult struct {
	Success       bool
	Error         string
	SignString    string
	Signature     string
	ClientSign    string
	DebugEnabled  bool
}

func NewVerificationResult() *VerificationResult {
	return &VerificationResult{
		Success:      false,
		DebugEnabled: false,
	}
}

type Verifier struct {
	config *VerifierConfig
}

func NewVerifier(config *VerifierConfig) *Verifier {
	if config == nil {
		config = DefaultVerifierConfig()
	}
	return &Verifier{config: config}
}

func (v *Verifier) isDebugMode(params map[string]string, headers map[string]string) bool {
	if v.config.IsProduction {
		return false
	}
	
	debugKey := v.config.DebugKey
	if val, ok := params[debugKey]; ok {
		return val == "true" || val == "1"
	}
	
	if val, ok := headers[debugKey]; ok {
		return val == "true" || val == "1"
	}
	
	return false
}

func (v *Verifier) Verify(appKey string, path string, params map[string]string, headers map[string]string) *VerificationResult {
	result := NewVerificationResult()
	
	debugEnabled := v.isDebugMode(params, headers)
	result.DebugEnabled = debugEnabled
	
	if appKey == "" {
		headerAppKey := headers[strings.ToLower(v.config.AppKeyHeaderName)]
		if headerAppKey == "" {
			result.Error = "missing app key in header"
			return result
		}
		appKey = headerAppKey
	}
	
	app, err := v.config.GetApp(appKey)
	if err != nil {
		result.Error = "invalid app key"
		return result
	}
	
	if !app.IsPathAllowed(path) {
		result.Error = "path not in whitelist"
		return result
	}
	
	clientSign, hasSign := params[v.config.SignatureKey]
	if !hasSign {
		result.Error = "missing signature"
		return result
	}
	result.ClientSign = clientSign
	
	timestampStr, hasTimestamp := params[v.config.TimestampKey]
	if !hasTimestamp {
		result.Error = "missing timestamp"
		return result
	}
	
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		result.Error = "invalid timestamp format"
		return result
	}
	
	now := time.Now().Unix()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	if diff > int64(v.config.TimestampTolerance.Seconds()) {
		result.Error = "timestamp expired"
		return result
	}
	
	signString := BuildSignString(params, v.config.SignatureKey)
	result.SignString = signString
	
	expectedSign := CalculateHMACSHA256(signString, app.AppSecret)
	result.Signature = expectedSign
	
	if !hmacEqual(expectedSign, clientSign) {
		result.Error = "signature mismatch"
		return result
	}
	
	result.Success = true
	return result
}

func hmacEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

func (v *Verifier) ValidatePath(appKey string, path string) error {
	if appKey == "" {
		return errors.New("missing app key")
	}
	
	app, err := v.config.GetApp(appKey)
	if err != nil {
		return err
	}
	
	if !app.IsPathAllowed(path) {
		return errors.New("path not in whitelist")
	}
	
	return nil
}
