package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"oauth2-server/pkg/common"
	"oauth2-server/pkg/oauth2"
)

const (
	defaultServerURL = "http://localhost:8080"
	defaultClientID  = "test-client"
	defaultSecret    = "test-secret"
	defaultRedirect  = "http://localhost:8080/callback"
	defaultScope     = "read write"
)

type TokenCache struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "login":
		loginCmd(args)
	case "token":
		tokenCmd(args)
	case "refresh":
		refreshCmd(args)
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  login    - Complete authorization code flow")
	fmt.Println("  token    - Exchange code for tokens manually")
	fmt.Println("  refresh  - Refresh tokens manually")
}

func loginCmd(args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	serverURL := fs.String("server", defaultServerURL, "OAuth2 server URL")
	clientID := fs.String("client-id", defaultClientID, "Client ID")
	clientSecret := fs.String("client-secret", defaultSecret, "Client secret")
	redirectURI := fs.String("redirect-uri", defaultRedirect, "Redirect URI")
	scope := fs.String("scope", defaultScope, "Scope (space separated)")
	fs.Parse(args)

	cache, err := loadTokenCache()
	if err == nil && time.Now().Before(cache.ExpiresAt) {
		fmt.Println("Token still valid, no need to refresh")
		printToken(cache)
		return
	}

	if cache != nil && cache.RefreshToken != "" {
		fmt.Println("Access token expired, attempting refresh...")
		refreshed, err := doRefresh(*serverURL, cache.RefreshToken)
		if err == nil {
			saveTokenCache(refreshed)
			fmt.Println("Token refreshed successfully")
			printToken(refreshed)
			return
		}
		fmt.Println("Refresh failed, starting new flow:", err)
	}

	codeVerifier := oauth2.GenerateCodeVerifier(64)
	codeChallenge := oauth2.GenerateCodeChallenge(codeVerifier)

	code, err := doAuthorize(*serverURL, *clientID, *redirectURI, *scope, codeChallenge)
	if err != nil {
		fmt.Println("Authorize failed:", err)
		os.Exit(1)
	}
	fmt.Printf("Got authorization code: %s\n", code)

	token, err := doToken(*serverURL, *clientID, *clientSecret, code, *redirectURI, codeVerifier)
	if err != nil {
		fmt.Println("Token exchange failed:", err)
		os.Exit(1)
	}

	saveTokenCache(token)
	fmt.Println("Authorization flow completed successfully")
	printToken(token)
}

func tokenCmd(args []string) {
	fs := flag.NewFlagSet("token", flag.ExitOnError)
	serverURL := fs.String("server", defaultServerURL, "OAuth2 server URL")
	clientID := fs.String("client-id", defaultClientID, "Client ID")
	clientSecret := fs.String("client-secret", defaultSecret, "Client secret")
	redirectURI := fs.String("redirect-uri", defaultRedirect, "Redirect URI")
	code := fs.String("code", "", "Authorization code")
	codeVerifier := fs.String("code-verifier", "", "PKCE code verifier")
	fs.Parse(args)

	if *code == "" || *codeVerifier == "" {
		fmt.Println("Error: --code and --code-verifier are required")
		fs.Usage()
		os.Exit(1)
	}

	token, err := doToken(*serverURL, *clientID, *clientSecret, *code, *redirectURI, *codeVerifier)
	if err != nil {
		fmt.Println("Token exchange failed:", err)
		os.Exit(1)
	}

	saveTokenCache(token)
	printToken(token)
}

func refreshCmd(args []string) {
	fs := flag.NewFlagSet("refresh", flag.ExitOnError)
	serverURL := fs.String("server", defaultServerURL, "OAuth2 server URL")
	refreshToken := fs.String("refresh-token", "", "Refresh token (uses cache if empty)")
	fs.Parse(args)

	if *refreshToken == "" {
		cache, err := loadTokenCache()
		if err != nil || cache.RefreshToken == "" {
			fmt.Println("No refresh token available. Please login first or provide --refresh-token")
			os.Exit(1)
		}
		*refreshToken = cache.RefreshToken
	}

	token, err := doRefresh(*serverURL, *refreshToken)
	if err != nil {
		fmt.Println("Refresh failed:", err)
		os.Exit(1)
	}

	saveTokenCache(token)
	fmt.Println("Token refreshed successfully")
	printToken(token)
}

func doAuthorize(serverURL, clientID, redirectURI, scope, codeChallenge string) (string, error) {
	req := common.AuthorizeRequest{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        "code",
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: "S256",
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/auth/authorize", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(data, &errResp)
		return "", errors.New(errResp.Error)
	}

	var authResp common.AuthorizeResponse
	if err := json.Unmarshal(data, &authResp); err != nil {
		return "", err
	}
	return authResp.Code, nil
}

func doToken(serverURL, clientID, clientSecret, code, redirectURI, codeVerifier string) (*TokenCache, error) {
	req := common.TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  redirectURI,
		ClientID:     clientID,
		CodeVerifier: codeVerifier,
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest(http.MethodPost, serverURL+"/auth/token", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)))

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, errors.New(errResp.Error)
	}

	var tokenResp common.TokenResponse
	if err := json.Unmarshal(data, &tokenResp); err != nil {
		return nil, err
	}

	return &TokenCache{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
	}, nil
}

func doRefresh(serverURL, refreshToken string) (*TokenCache, error) {
	req := common.RefreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/auth/refresh", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, errors.New(errResp.Error)
	}

	var tokenResp common.TokenResponse
	if err := json.Unmarshal(data, &tokenResp); err != nil {
		return nil, err
	}

	return &TokenCache{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
	}, nil
}

func cacheFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".oauth2_client_token_cache.json")
}

func loadTokenCache() (*TokenCache, error) {
	data, err := os.ReadFile(cacheFilePath())
	if err != nil {
		return nil, err
	}
	var cache TokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func saveTokenCache(cache *TokenCache) error {
	data, _ := json.MarshalIndent(cache, "", "  ")
	return os.WriteFile(cacheFilePath(), data, 0600)
}

func printToken(cache *TokenCache) {
	fmt.Println("\n=== Token Information ===")
	fmt.Printf("Access Token: %s\n", cache.AccessToken)
	fmt.Printf("Refresh Token: %s\n", cache.RefreshToken)
	fmt.Printf("Token Type: %s\n", cache.TokenType)
	fmt.Printf("Expires At: %s\n", cache.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("Scope: %s\n", cache.Scope)
}
