package signature

import (
	"errors"
	"strings"
	"time"
)

type AppConfig struct {
	AppKey      string
	AppSecret   string
	PathWhitelist []string
}

type VerifierConfig struct {
	Apps             map[string]*AppConfig
	TimestampKey     string
	SignatureKey     string
	AppKeyHeaderName string
	TimestampTolerance time.Duration
	DebugKey         string
	IsProduction     bool
}

func DefaultVerifierConfig() *VerifierConfig {
	return &VerifierConfig{
		Apps:             make(map[string]*AppConfig),
		TimestampKey:     "timestamp",
		SignatureKey:     "signature",
		AppKeyHeaderName: "X-App-Key",
		TimestampTolerance: 5 * time.Minute,
		DebugKey:         "debug",
		IsProduction:     false,
	}
}

func (vc *VerifierConfig) AddApp(appKey, appSecret string, pathWhitelist []string) {
	vc.Apps[appKey] = &AppConfig{
		AppKey:        appKey,
		AppSecret:     appSecret,
		PathWhitelist: pathWhitelist,
	}
}

func (vc *VerifierConfig) GetApp(appKey string) (*AppConfig, error) {
	app, exists := vc.Apps[appKey]
	if !exists {
		return nil, errors.New("app key not found")
	}
	return app, nil
}

func (ac *AppConfig) IsPathAllowed(path string) bool {
	if len(ac.PathWhitelist) == 0 {
		return true
	}
	for _, allowedPath := range ac.PathWhitelist {
		if path == allowedPath || strings.HasPrefix(path, allowedPath+"/") {
			return true
		}
	}
	return false
}
