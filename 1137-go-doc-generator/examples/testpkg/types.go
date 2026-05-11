package testpkg

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultPort is the default server port.
	DefaultPort = 8080

	// DefaultTimeout is the default timeout duration.
	DefaultTimeout = 30 * time.Second

	// MaxRetries is the maximum number of retry attempts.
	MaxRetries = 3
)

var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput is returned when input is invalid.
	ErrInvalidInput = errors.New("invalid input")
)

type ConfigType string

const (
	ConfigTypeDevelopment ConfigType = "development"
	ConfigTypeProduction  ConfigType = "production"
	ConfigTypeStaging     ConfigType = "staging"
)

// Config represents the configuration for an application.
//
// Config stores all necessary settings for running the application,
// including database connections, server settings, and feature flags.
//
// Example:
//
//	cfg := &Config{
//	    Name: "my-app",
//	    Port: 8080,
//	    Debug: true,
//	}
type Config struct {
	// Name is the name of the application.
	Name string `json:"name" yaml:"name"`

	// Port is the port the server listens on.
	Port int `json:"port" yaml:"port"`

	// Debug enables debug mode.
	Debug bool `json:"debug" yaml:"debug"`

	// Timeout is the request timeout duration.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// Database is the database configuration.
	Database DatabaseConfig `json:"database" yaml:"database"`

	// Features contains feature flags.
	Features map[string]bool `json:"features" yaml:"features"`

	// Tags are additional tags for the configuration.
	Tags []string `json:"tags" yaml:"tags"`

	// internalField is an internal field (unexported).
	internalField string
}

// DatabaseConfig represents database connection settings.
type DatabaseConfig struct {
	// Host is the database host.
	Host string `json:"host"`

	// Port is the database port.
	Port int `json:"port"`

	// User is the database user.
	User string `json:"user"`

	// Password is the database password.
	Password string `json:"password"`

	// Name is the database name.
	Name string `json:"name"`
}

// Logger is an alias for a logging function.
type Logger = func(format string, args ...interface{})

// Service represents a generic service interface.
type Service interface {
	// Start starts the service.
	Start(ctx context.Context) error

	// Stop stops the service.
	Stop(ctx context.Context) error

	// HealthCheck performs a health check.
	HealthCheck(ctx context.Context) (bool, error)
}

// Client represents a client for making requests.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     Logger
}

// NewClient creates a new Client.
//
// NewClient initializes a client with the given base URL and options.
//
// Example:
//
//	client := NewClient("https://api.example.com")
//	resp, err := client.Get("/users")
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Get makes a GET request.
//
// Get sends a GET request to the specified path and returns the response.
// The path should be relative to the base URL.
func (c *Client) Get(path string) (*http.Response, error) {
	return c.httpClient.Get(c.baseURL + path)
}

// Post makes a POST request.
func (c *Client) Post(path string, body io.Reader) (*http.Response, error) {
	return c.httpClient.Post(c.baseURL+path, "application/json", body)
}

// SetLogger sets the logger for the client.
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
}

// NewConfig creates a new Config with default values.
//
// NewConfig initializes a configuration with sensible defaults.
// The name parameter is required and cannot be empty.
//
// Example:
//
//	cfg := NewConfig("my-app")
//	cfg.Port = 9090
func NewConfig(name string) (*Config, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}

	return &Config{
		Name:    name,
		Port:    DefaultPort,
		Debug:   false,
		Timeout: DefaultTimeout,
	}, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("invalid port")
	}
	return nil
}

// String returns a string representation of the config.
func (c *Config) String() string {
	return c.Name
}
