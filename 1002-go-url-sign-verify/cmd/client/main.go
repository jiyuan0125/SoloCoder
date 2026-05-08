package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/example/url-sign-verify/pkg/signature"
)

type ClientConfig struct {
	AppKey    string
	AppSecret string
	BaseURL   string
	Debug     bool
}

func main() {
	config := &ClientConfig{
		BaseURL: "http://localhost:8080",
	}
	
	flag.StringVar(&config.AppKey, "app-key", "partner-a", "Application key")
	flag.StringVar(&config.AppSecret, "app-secret", "secret-key-for-partner-a", "Application secret")
	flag.StringVar(&config.BaseURL, "base-url", "http://localhost:8080", "Server base URL")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug mode")
	
	cmd := flag.String("cmd", "health", "Command to execute: health, get-users, create-user")
	
	name := flag.String("name", "", "User name (for create-user command)")
	age := flag.Int("age", 0, "User age (for create-user command)")
	email := flag.String("email", "", "User email (for create-user command)")
	
	flag.Parse()
	
	switch *cmd {
	case "health":
		executeHealthCheck(config)
	case "get-users":
		executeGetUsers(config)
	case "create-user":
		executeCreateUser(config, *name, *age, *email)
	default:
		fmt.Printf("Unknown command: %s\n", *cmd)
		fmt.Println("Available commands: health, get-users, create-user")
	}
}

func executeHealthCheck(config *ClientConfig) {
	params := make(map[string]string)
	timestamp := time.Now().Unix()
	params["timestamp"] = strconv.FormatInt(timestamp, 10)
	
	if config.Debug {
		params["debug"] = "true"
	}
	
	signatureKey := "signature"
	params[signatureKey] = signature.GenerateSignature(params, config.AppSecret, signatureKey)
	
	queryString := buildQueryString(params)
	url := fmt.Sprintf("%s/api/health?%s", config.BaseURL, queryString)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	
	req.Header.Set("X-App-Key", config.AppKey)
	
	executeRequest(req, config)
}

func executeGetUsers(config *ClientConfig) {
	params := make(map[string]string)
	timestamp := time.Now().Unix()
	params["timestamp"] = strconv.FormatInt(timestamp, 10)
	
	if config.Debug {
		params["debug"] = "true"
	}
	
	signatureKey := "signature"
	params[signatureKey] = signature.GenerateSignature(params, config.AppSecret, signatureKey)
	
	queryString := buildQueryString(params)
	url := fmt.Sprintf("%s/api/users?%s", config.BaseURL, queryString)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	
	req.Header.Set("X-App-Key", config.AppKey)
	
	executeRequest(req, config)
}

func executeCreateUser(config *ClientConfig, name string, age int, email string) {
	if name == "" {
		fmt.Println("Error: name is required for create-user command")
		return
	}
	
	params := make(map[string]string)
	timestamp := time.Now().Unix()
	params["timestamp"] = strconv.FormatInt(timestamp, 10)
	params["name"] = name
	params["age"] = strconv.Itoa(age)
	params["email"] = email
	
	if config.Debug {
		params["debug"] = "true"
	}
	
	signatureKey := "signature"
	params[signatureKey] = signature.GenerateSignature(params, config.AppSecret, signatureKey)
	
	jsonBody, err := json.Marshal(map[string]interface{}{
		"name":  name,
		"age":   age,
		"email": email,
	})
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	
	queryString := buildQueryString(params)
	url := fmt.Sprintf("%s/api/users?%s", config.BaseURL, queryString)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	
	req.Header.Set("X-App-Key", config.AppKey)
	req.Header.Set("Content-Type", "application/json")
	
	executeRequest(req, config)
}

func buildQueryString(params map[string]string) string {
	values := url.Values{}
	for key, val := range params {
		values.Add(key, val)
	}
	return values.Encode()
}

func executeRequest(req *http.Request, config *ClientConfig) {
	if config.Debug {
		fmt.Printf("Request: %s %s\n", req.Method, req.URL)
		fmt.Printf("Headers: %v\n", req.Header)
	}
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error executing request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}
	
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response:\n%s\n", string(body))
}

func printSignString(params map[string]string, signatureKey string) {
	fmt.Println("Sign string:", signature.BuildSignString(params, signatureKey))
}
