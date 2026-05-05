// Package config provides a flexible configuration loading library for Go applications.
// It supports loading configuration from multiple sources with priority merging,
// automatic type conversion, environment variable mapping, and hot reloading.
//
// Priority order (highest to lowest):
//   1. Command-line arguments
//   2. Environment variables
//   3. Configuration files (YAML or JSON)
//   4. Default values from struct tags
//
// Example usage:
//
//	type DatabaseConfig struct {
//	    Host     string `config:"host,required" default:"localhost"`
//	    Port     int    `config:"port" default:"5432"`
//	    Username string `config:"username,required"`
//	    Password string `config:"password,required"`
//	}
//
//	type AppConfig struct {
//	    Database DatabaseConfig `config:"database"`
//	    LogLevel string         `config:"log_level" default:"info"`
//	}
//
//	loader := config.NewLoader()
//	loader.AddFile("config.yaml")
//	loader.SetEnvPrefix("APP")
//
//	var cfg AppConfig
//	if err := loader.Load(&cfg); err != nil {
//	    log.Fatal(err)
//	}
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SourceType represents the type of configuration source.
type SourceType int

const (
	// SourceFile represents configuration from a file (YAML/JSON).
	SourceFile SourceType = iota
	// SourceEnv represents configuration from environment variables.
	SourceEnv
	// SourceCLI represents configuration from command-line arguments.
	SourceCLI
	// SourceDefault represents default values from struct tags.
	SourceDefault
)

// String returns the string representation of SourceType.
func (st SourceType) String() string {
	switch st {
	case SourceFile:
		return "file"
	case SourceEnv:
		return "environment"
	case SourceCLI:
		return "command-line"
	case SourceDefault:
		return "default"
	default:
		return "unknown"
	}
}

// Config holds the merged configuration data from all sources.
type Config struct {
	data     map[string]interface{}
	mu       sync.RWMutex
	metadata map[string]SourceType
}

// New creates a new empty Config instance.
func New() *Config {
	return &Config{
		data:     make(map[string]interface{}),
		metadata: make(map[string]SourceType),
	}
}

// Get retrieves a configuration value by its dot-separated path.
// Returns nil if the path does not exist.
//
// Example:
//
//	value := cfg.Get("database.host")
func (c *Config) Get(path string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := strings.Split(path, ".")
	current := c.data

	for i, key := range keys {
		if i == len(keys)-1 {
			return current[key]
		}

		next, ok := current[key].(map[string]interface{})
		if !ok {
			return nil
		}
		current = next
	}

	return nil
}

// GetString retrieves a configuration value as a string.
// Returns an error if the value cannot be converted to string.
//
// Example:
//
//	host, err := cfg.GetString("database.host")
func (c *Config) GetString(path string) (string, error) {
	val := c.Get(path)
	if val == nil {
		return "", fmt.Errorf("config path not found: %s", path)
	}

	switch v := val.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", fmt.Errorf("cannot convert %T to string for path: %s", val, path)
	}
}

// GetInt retrieves a configuration value as an int.
// Returns an error if the value cannot be converted to int.
//
// Example:
//
//	port, err := cfg.GetInt("database.port")
func (c *Config) GetInt(path string) (int, error) {
	val := c.Get(path)
	if val == nil {
		return 0, fmt.Errorf("config path not found: %s", path)
	}

	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, newConversionError(path, v, "int")
		}
		return i, nil
	default:
		return 0, newConversionError(path, v, "int")
	}
}

// GetInt64 retrieves a configuration value as an int64.
// Returns an error if the value cannot be converted to int64.
func (c *Config) GetInt64(path string) (int64, error) {
	val := c.Get(path)
	if val == nil {
		return 0, fmt.Errorf("config path not found: %s", path)
	}

	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return 0, newConversionError(path, v, "int64")
		}
		return i, nil
	default:
		return 0, newConversionError(path, v, "int64")
	}
}

// GetFloat64 retrieves a configuration value as a float64.
// Returns an error if the value cannot be converted to float64.
//
// Example:
//
//	ratio, err := cfg.GetFloat64("app.ratio")
func (c *Config) GetFloat64(path string) (float64, error) {
	val := c.Get(path)
	if val == nil {
		return 0, fmt.Errorf("config path not found: %s", path)
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, newConversionError(path, v, "float64")
		}
		return f, nil
	default:
		return 0, newConversionError(path, v, "float64")
	}
}

// GetBool retrieves a configuration value as a bool.
// String values "true", "false", "1", "0", "yes", "no", "on", "off" are supported (case-insensitive).
// Returns an error if the value cannot be converted to bool.
//
// Example:
//
//	debug, err := cfg.GetBool("app.debug")
func (c *Config) GetBool(path string) (bool, error) {
	val := c.Get(path)
	if val == nil {
		return false, fmt.Errorf("config path not found: %s", path)
	}

	switch v := val.(type) {
	case bool:
		return v, nil
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		switch s {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		default:
			return false, newConversionError(path, v, "bool")
		}
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, newConversionError(path, v, "bool")
	}
}

// GetMap retrieves a configuration value as a map[string]interface{}.
// Returns an error if the value is not a map.
func (c *Config) GetMap(path string) (map[string]interface{}, error) {
	val := c.Get(path)
	if val == nil {
		return nil, fmt.Errorf("config path not found: %s", path)
	}

	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("config path %s is not a map, got %T", path, val)
	}

	return m, nil
}

// GetSlice retrieves a configuration value as a []interface{}.
// Returns an error if the value is not a slice.
func (c *Config) GetSlice(path string) ([]interface{}, error) {
	val := c.Get(path)
	if val == nil {
		return nil, fmt.Errorf("config path not found: %s", path)
	}

	s, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("config path %s is not a slice, got %T", path, val)
	}

	return s, nil
}

// Set sets a configuration value at the given dot-separated path.
// Creates intermediate maps if necessary.
//
// Example:
//
//	cfg.Set("database.host", "localhost")
func (c *Config) Set(path string, value interface{}, source SourceType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := strings.Split(path, ".")
	current := c.data

	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]
		if _, ok := current[key]; !ok {
			current[key] = make(map[string]interface{})
		}
		next, ok := current[key].(map[string]interface{})
		if !ok {
			current[key] = make(map[string]interface{})
			next = current[key].(map[string]interface{})
		}
		current = next
	}

	current[keys[len(keys)-1]] = value
	c.metadata[path] = source
}

// Merge merges another Config into this one.
// Values from the other config take precedence.
func (c *Config) Merge(other *Config) {
	other.mu.RLock()
	defer other.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	mergeMaps(c.data, other.data)
}

// mergeMaps recursively merges src into dst.
func mergeMaps(dst, src map[string]interface{}) {
	for key, srcVal := range src {
		dstVal, exists := dst[key]

		if !exists {
			dst[key] = srcVal
			continue
		}

		srcMap, srcIsMap := srcVal.(map[string]interface{})
		dstMap, dstIsMap := dstVal.(map[string]interface{})

		if srcIsMap && dstIsMap {
			mergeMaps(dstMap, srcMap)
		} else {
			dst[key] = srcVal
		}
	}
}

// All returns all configuration data as a map.
func (c *Config) All() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return deepCopyMap(c.data)
}

// deepCopyMap creates a deep copy of a map.
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = deepCopyMap(val)
		case []interface{}:
			result[k] = deepCopySlice(val)
		default:
			result[k] = v
		}
	}
	return result
}

// deepCopySlice creates a deep copy of a slice.
func deepCopySlice(s []interface{}) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		switch val := v.(type) {
		case map[string]interface{}:
			result[i] = deepCopyMap(val)
		case []interface{}:
			result[i] = deepCopySlice(val)
		default:
			result[i] = v
		}
	}
	return result
}

// ConversionError represents an error that occurs when converting a configuration value to a specific type.
type ConversionError struct {
	Path     string
	Value    interface{}
	Expected string
}

// newConversionError creates a new ConversionError.
func newConversionError(path string, value interface{}, expected string) *ConversionError {
	return &ConversionError{
		Path:     path,
		Value:    value,
		Expected: expected,
	}
}

// Error returns the string representation of the ConversionError.
func (e *ConversionError) Error() string {
	return fmt.Sprintf("conversion error: cannot convert value '%v' (type %T) at path '%s' to %s",
		e.Value, e.Value, e.Path, e.Expected)
}

// Is returns true if the error is a ConversionError.
func (e *ConversionError) Is(target error) bool {
	_, ok := target.(*ConversionError)
	return ok
}

// MissingRequiredError represents an error that occurs when a required configuration field is missing.
type MissingRequiredError struct {
	Fields []string
}

// Error returns the string representation of the MissingRequiredError.
func (e *MissingRequiredError) Error() string {
	if len(e.Fields) == 0 {
		return "missing required configuration fields"
	}
	if len(e.Fields) == 1 {
		return fmt.Sprintf("missing required configuration field: %s", e.Fields[0])
	}
	return fmt.Sprintf("missing %d required configuration fields: %v", len(e.Fields), e.Fields)
}

// Is returns true if the error is a MissingRequiredError.
func (e *MissingRequiredError) Is(target error) bool {
	_, ok := target.(*MissingRequiredError)
	return ok
}

// LoadError represents an error that occurs during configuration loading.
type LoadError struct {
	Source  string
	Message string
	Err     error
}

// Error returns the string representation of the LoadError.
func (e *LoadError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("load error from %s: %s: %v", e.Source, e.Message, e.Err)
	}
	return fmt.Sprintf("load error from %s: %s", e.Source, e.Message)
}

// Unwrap returns the underlying error.
func (e *LoadError) Unwrap() error {
	return e.Err
}

// Is returns true if the error is a LoadError.
func (e *LoadError) Is(target error) bool {
	_, ok := target.(*LoadError)
	return ok
}

// WatchCallback is a function type that is called when a configuration file changes.
type WatchCallback func(oldCfg, newCfg *Config, err error)

// Watcher watches configuration files for changes.
type Watcher struct {
	files     []string
	callbacks []WatchCallback
	stopChan  chan struct{}
	running   bool
	mu        sync.Mutex
	interval  time.Duration
	modTimes  map[string]time.Time
}

// NewWatcher creates a new configuration file watcher.
func NewWatcher() *Watcher {
	return &Watcher{
		stopChan: make(chan struct{}),
		interval: 2 * time.Second,
		modTimes: make(map[string]time.Time),
	}
}

// AddFile adds a file to watch.
func (w *Watcher) AddFile(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, f := range w.files {
		if f == path {
			return
		}
	}
	w.files = append(w.files, path)
}

// SetInterval sets the polling interval for file changes.
func (w *Watcher) SetInterval(interval time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.interval = interval
}

// OnChange registers a callback to be called when any watched file changes.
func (w *Watcher) OnChange(callback WatchCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, callback)
}

// Start starts watching for file changes.
func (w *Watcher) Start(loader *Loader, target interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return errors.New("watcher is already running")
	}

	if len(w.files) == 0 {
		return errors.New("no files to watch")
	}

	for _, file := range w.files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		w.modTimes[file] = info.ModTime()
	}

	w.running = true

	go w.watchLoop(loader, target)

	return nil
}

// Stop stops watching for file changes.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	close(w.stopChan)
	w.running = false
}

// watchLoop polls for file changes.
func (w *Watcher) watchLoop(loader *Loader, target interface{}) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkChanges(loader, target)
		case <-w.stopChan:
			return
		}
	}
}

// checkChanges checks if any watched files have changed.
func (w *Watcher) checkChanges(loader *Loader, target interface{}) {
	w.mu.Lock()
	changed := false
	var changedFiles []string

	for _, file := range w.files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		prevTime, exists := w.modTimes[file]
		if !exists || info.ModTime().After(prevTime) {
			w.modTimes[file] = info.ModTime()
			changed = true
			changedFiles = append(changedFiles, file)
		}
	}
	w.mu.Unlock()

	if changed && len(w.callbacks) > 0 {
		oldConfig := New()
		if existing, ok := target.(interface{ GetConfig() *Config }); ok {
			oldConfig = existing.GetConfig()
		}

		newLoader := NewLoader()
		for _, f := range w.files {
			newLoader.AddFile(f)
		}

		newConfig, err := newLoader.LoadToConfig()
		for _, callback := range w.callbacks {
			callback(oldConfig, newConfig, err)
		}

		if err == nil && target != nil {
			_ = newConfig.Unmarshal(target)
		}
	}
}

// Loader handles loading configuration from multiple sources.
type Loader struct {
	files         []string
	envPrefix     string
	cliArgs       []string
	envEnabled    bool
	cliEnabled    bool
	envMap        map[string]string
	watcher       *Watcher
	watchEnabled  bool
	watchInterval time.Duration
}

// LoaderOption is a function type for configuring the Loader.
type LoaderOption func(*Loader)

// NewLoader creates a new configuration loader with default settings.
// By default, environment variables and command-line arguments are enabled.
//
// Example:
//
//	loader := config.NewLoader()
func NewLoader() *Loader {
	return &Loader{
		envEnabled:   true,
		cliEnabled:   true,
		cliArgs:      os.Args[1:],
		watchEnabled: false,
	}
}

// WithFiles sets the configuration files to load.
// Later files in the list have higher priority.
func WithFiles(files ...string) LoaderOption {
	return func(l *Loader) {
		l.files = files
	}
}

// WithEnvPrefix sets the prefix for environment variables.
// For example, with prefix "APP", "database.host" maps to "APP_DATABASE_HOST".
func WithEnvPrefix(prefix string) LoaderOption {
	return func(l *Loader) {
		l.envPrefix = prefix
	}
}

// WithCLIArgs sets the command-line arguments to parse.
// Defaults to os.Args[1:].
func WithCLIArgs(args []string) LoaderOption {
	return func(l *Loader) {
		l.cliArgs = args
	}
}

// WithEnvDisabled disables loading from environment variables.
func WithEnvDisabled() LoaderOption {
	return func(l *Loader) {
		l.envEnabled = false
	}
}

// WithCLIDisabled disables loading from command-line arguments.
func WithCLIDisabled() LoaderOption {
	return func(l *Loader) {
		l.cliEnabled = false
	}
}

// WithWatch enables watching configuration files for changes.
// The default polling interval is 2 seconds.
func WithWatch() LoaderOption {
	return func(l *Loader) {
		l.watchEnabled = true
	}
}

// WithWatchInterval sets the polling interval for file watching.
func WithWatchInterval(interval time.Duration) LoaderOption {
	return func(l *Loader) {
		l.watchInterval = interval
	}
}

// AddFile adds a configuration file to load.
// Files are loaded in the order they are added, with later files having higher priority.
//
// Example:
//
//	loader.AddFile("config.yaml")
//	loader.AddFile("config.local.yaml") // overrides values from config.yaml
func (l *Loader) AddFile(path string) {
	l.files = append(l.files, path)
}

// SetEnvPrefix sets the prefix for environment variables.
//
// Example:
//
//	loader.SetEnvPrefix("APP")
//	// Now "database.host" maps to "APP_DATABASE_HOST"
func (l *Loader) SetEnvPrefix(prefix string) {
	l.envPrefix = prefix
}

// SetCLIArgs sets the command-line arguments to parse.
func (l *Loader) SetCLIArgs(args []string) {
	l.cliArgs = args
}

// DisableEnv disables loading from environment variables.
func (l *Loader) DisableEnv() {
	l.envEnabled = false
}

// DisableCLI disables loading from command-line arguments.
func (l *Loader) DisableCLI() {
	l.cliEnabled = false
}

// EnableWatch enables watching configuration files for changes.
func (l *Loader) EnableWatch() {
	l.watchEnabled = true
}

// SetWatchInterval sets the polling interval for file watching.
func (l *Loader) SetWatchInterval(interval time.Duration) {
	l.watchInterval = interval
}

// OnChange registers a callback to be called when configuration files change.
// Only effective if watching is enabled.
func (l *Loader) OnChange(callback WatchCallback) {
	if l.watcher == nil {
		l.watcher = NewWatcher()
	}
	l.watcher.OnChange(callback)
}

// Load loads configuration from all sources and unmarshals it into the target struct.
//
// Priority order (highest to lowest):
//   1. Command-line arguments
//   2. Environment variables
//   3. Configuration files
//   4. Default values from struct tags
//
// Example:
//
//	var cfg AppConfig
//	if err := loader.Load(&cfg); err != nil {
//	    log.Fatal(err)
//	}
func (l *Loader) Load(target interface{}) error {
	if target == nil {
		return errors.New("target cannot be nil")
	}

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("target must be a non-nil pointer")
	}

	defaults, requiredFields := l.extractDefaultsAndRequired(v.Elem())

	loadedConfig, err := l.LoadToConfig()
	if err != nil {
		return err
	}

	defaults.Merge(loadedConfig)

	if err := defaults.Unmarshal(target); err != nil {
		return err
	}

	missing := l.checkRequiredFields(defaults, requiredFields)
	if len(missing) > 0 {
		return &MissingRequiredError{Fields: missing}
	}

	if l.watchEnabled && len(l.files) > 0 {
		if l.watcher == nil {
			l.watcher = NewWatcher()
		}
		for _, f := range l.files {
			l.watcher.AddFile(f)
		}
		if l.watchInterval > 0 {
			l.watcher.SetInterval(l.watchInterval)
		}
		if err := l.watcher.Start(l, target); err != nil {
			return err
		}
	}

	return nil
}

// LoadToConfig loads configuration from all sources and returns a Config instance.
// This is useful when you want to access configuration values dynamically without a struct.
//
// Example:
//
//	cfg, err := loader.LoadToConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	host, _ := cfg.GetString("database.host")
func (l *Loader) LoadToConfig() (*Config, error) {
	result := New()

	for _, file := range l.files {
		fileConfig, err := l.loadFromFile(file)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, &LoadError{
					Source:  "file",
					Message: fmt.Sprintf("failed to load file %s", file),
					Err:     err,
				}
			}
			continue
		}
		result.Merge(fileConfig)
	}

	if l.envEnabled {
		envConfig := l.loadFromEnv()
		result.Merge(envConfig)
	}

	if l.cliEnabled {
		cliConfig := l.loadFromCLI()
		result.Merge(cliConfig)
	}

	return result, nil
}

// loadFromFile loads configuration from a file (YAML or JSON).
func (l *Loader) loadFromFile(path string) (*Config, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return loadYAMLFile(path)
	case ".json":
		return loadJSONFile(path)
	default:
		return nil, fmt.Errorf("unsupported file format: %s (supported: .yaml, .yml, .json)", ext)
	}
}

// loadFromEnv loads configuration from environment variables.
// Only loads environment variables if envPrefix is set.
func (l *Loader) loadFromEnv() *Config {
	config := New()
	prefix := strings.ToUpper(l.envPrefix)

	if prefix == "" {
		return config
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if !strings.HasPrefix(key, prefix+"_") {
			continue
		}
		key = strings.TrimPrefix(key, prefix+"_")

		path := strings.ReplaceAll(strings.ToLower(key), "_", ".")
		config.Set(path, parseValue(value), SourceEnv)
	}

	return config
}

// loadFromCLI loads configuration from command-line arguments.
// Supports formats: --key=value, --key value, -key=value, -key value
func (l *Loader) loadFromCLI() *Config {
	config := New()

	for i := 0; i < len(l.cliArgs); i++ {
		arg := l.cliArgs[i]

		if strings.HasPrefix(arg, "--") {
			arg = strings.TrimPrefix(arg, "--")
		} else if strings.HasPrefix(arg, "-") {
			arg = strings.TrimPrefix(arg, "-")
		} else {
			continue
		}

		if idx := strings.Index(arg, "="); idx != -1 {
			key := arg[:idx]
			value := arg[idx+1:]
			path := strings.ReplaceAll(key, "-", ".")
			config.Set(path, parseValue(value), SourceCLI)
		} else {
			key := arg
			if i+1 < len(l.cliArgs) {
				value := l.cliArgs[i+1]
				if !strings.HasPrefix(value, "-") {
					path := strings.ReplaceAll(key, "-", ".")
					config.Set(path, parseValue(value), SourceCLI)
					i++
				} else {
					path := strings.ReplaceAll(key, "-", ".")
					config.Set(path, true, SourceCLI)
				}
			} else {
				path := strings.ReplaceAll(key, "-", ".")
				config.Set(path, true, SourceCLI)
			}
		}
	}

	return config
}

// parseValue attempts to parse a string value into its appropriate type.
func parseValue(s string) interface{} {
	s = strings.TrimSpace(s)

	if s == "true" || s == "TRUE" || s == "True" {
		return true
	}
	if s == "false" || s == "FALSE" || s == "False" {
		return false
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	return s
}

// extractDefaultsAndRequired extracts default values and required fields from struct tags.
func (l *Loader) extractDefaultsAndRequired(v reflect.Value) (*Config, []string) {
	config := New()
	var required []string

	l.extractDefaultsFromValue(v, "", config, &required)

	return config, required
}

// extractDefaultsFromValue recursively extracts defaults from a reflect.Value.
func (l *Loader) extractDefaultsFromValue(v reflect.Value, prefix string, config *Config, required *[]string) {
	t := v.Type()

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			if !fieldValue.CanSet() {
				continue
			}

			tag := field.Tag.Get("config")
			if tag == "" {
				continue
			}

			parts := strings.Split(tag, ",")
			fieldName := parts[0]

			var fullPath string
			if prefix == "" {
				fullPath = fieldName
			} else {
				fullPath = prefix + "." + fieldName
			}

			isRequired := false
			for _, p := range parts[1:] {
				if p == "required" {
					isRequired = true
				}
			}

			if isRequired {
				*required = append(*required, fullPath)
			}

			defaultTag := field.Tag.Get("default")
			if defaultTag != "" {
				config.Set(fullPath, parseValue(defaultTag), SourceDefault)
			}

			if fieldValue.Kind() == reflect.Struct ||
				(fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct) {
				var structValue reflect.Value
				if fieldValue.Kind() == reflect.Ptr {
					if fieldValue.IsNil() {
						fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
					}
					structValue = fieldValue.Elem()
				} else {
					structValue = fieldValue
				}
				l.extractDefaultsFromValue(structValue, fullPath, config, required)
			}
		}

	case reflect.Ptr:
		if !v.IsNil() {
			l.extractDefaultsFromValue(v.Elem(), prefix, config, required)
		}
	}
}

// checkRequiredFields checks if all required fields are present in the config.
func (l *Loader) checkRequiredFields(config *Config, required []string) []string {
	var missing []string

	for _, field := range required {
		if config.Get(field) == nil {
			missing = append(missing, field)
		}
	}

	return missing
}

// GetWatcher returns the underlying file watcher.
func (l *Loader) GetWatcher() *Watcher {
	return l.watcher
}

// Unmarshal unmarshals the Config into the target struct.
// Supports struct tags:
//   - `config:"fieldName"` - specifies the configuration key name
//   - `default:"value"` - specifies a default value
//   - `config:"fieldName,required"` - marks the field as required
func (c *Config) Unmarshal(target interface{}) error {
	if target == nil {
		return errors.New("target cannot be nil")
	}

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("target must be a non-nil pointer")
	}

	return c.unmarshalValue(v.Elem(), c.data)
}

// unmarshalValue recursively unmarshals a value into the target.
func (c *Config) unmarshalValue(v reflect.Value, data interface{}) error {
	if data == nil {
		return nil
	}

	switch v.Kind() {
	case reflect.Struct:
		return c.unmarshalStruct(v, data)
	case reflect.Map:
		return c.unmarshalMap(v, data)
	case reflect.Slice:
		return c.unmarshalSlice(v, data)
	case reflect.Ptr:
		return c.unmarshalPtr(v, data)
	case reflect.Interface:
		if v.IsNil() {
			v.Set(reflect.ValueOf(data))
			return nil
		}
		return c.unmarshalValue(v.Elem(), data)
	default:
		return c.unmarshalScalar(v, data)
	}
}

// unmarshalStruct unmarshals data into a struct.
func (c *Config) unmarshalStruct(v reflect.Value, data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot unmarshal %T into struct", data)
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		tag := field.Tag.Get("config")
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		fieldName := parts[0]

		fieldData, exists := dataMap[fieldName]
		if !exists {
			continue
		}

		if err := c.unmarshalValue(fieldValue, fieldData); err != nil {
			return fmt.Errorf("field %s: %w", fieldName, err)
		}
	}

	return nil
}

// unmarshalMap unmarshals data into a map.
func (c *Config) unmarshalMap(v reflect.Value, data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot unmarshal %T into map", data)
	}

	t := v.Type()
	keyType := t.Key()
	elemType := t.Elem()

	if keyType.Kind() != reflect.String {
		return fmt.Errorf("map keys must be string, got %s", keyType.Kind())
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(t))
	}

	for key, val := range dataMap {
		keyValue := reflect.ValueOf(key)
		elemValue := reflect.New(elemType).Elem()

		if err := c.unmarshalValue(elemValue, val); err != nil {
			return fmt.Errorf("map key %s: %w", key, err)
		}

		v.SetMapIndex(keyValue, elemValue)
	}

	return nil
}

// unmarshalSlice unmarshals data into a slice.
func (c *Config) unmarshalSlice(v reflect.Value, data interface{}) error {
	dataSlice, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("cannot unmarshal %T into slice", data)
	}

	t := v.Type()

	newSlice := reflect.MakeSlice(t, len(dataSlice), len(dataSlice))

	for i, val := range dataSlice {
		elemValue := newSlice.Index(i)
		if err := c.unmarshalValue(elemValue, val); err != nil {
			return fmt.Errorf("slice index %d: %w", i, err)
		}
	}

	v.Set(newSlice)
	return nil
}

// unmarshalPtr unmarshals data into a pointer.
func (c *Config) unmarshalPtr(v reflect.Value, data interface{}) error {
	if data == nil {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}

	return c.unmarshalValue(v.Elem(), data)
}

// unmarshalScalar unmarshals data into a scalar value.
func (c *Config) unmarshalScalar(v reflect.Value, data interface{}) error {
	switch v.Kind() {
	case reflect.String:
		s, err := convertToString(data)
		if err != nil {
			return err
		}
		v.SetString(s)
		return nil

	case reflect.Bool:
		b, err := convertToBool(data)
		if err != nil {
			return err
		}
		v.SetBool(b)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := convertToInt(data, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(i)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := convertToUint(data, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(u)
		return nil

	case reflect.Float32, reflect.Float64:
		f, err := convertToFloat(data, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(f)
		return nil

	default:
		return fmt.Errorf("unsupported type: %s", v.Kind())
	}
}

// convertToString converts any value to string.
func convertToString(data interface{}) (string, error) {
	switch v := data.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// convertToBool converts any value to bool.
func convertToBool(data interface{}) (bool, error) {
	switch v := data.(type) {
	case bool:
		return v, nil
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		switch s {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		default:
			return false, fmt.Errorf("cannot convert string '%s' to bool", v)
		}
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// convertToInt converts any value to int64.
func convertToInt(data interface{}, bits int) (int64, error) {
	switch v := data.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(v), 10, bits)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to int%d: %w", v, bits, err)
		}
		return i, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int%d", v, bits)
	}
}

// convertToUint converts any value to uint64.
func convertToUint(data interface{}, bits int) (uint64, error) {
	switch v := data.(type) {
	case uint:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative int %d to uint%d", v, bits)
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative int64 %d to uint%d", v, bits)
		}
		return uint64(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative float64 %f to uint%d", v, bits)
		}
		return uint64(v), nil
	case string:
		u, err := strconv.ParseUint(strings.TrimSpace(v), 10, bits)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to uint%d: %w", v, bits, err)
		}
		return u, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to uint%d", v, bits)
	}
}

// convertToFloat converts any value to float64.
func convertToFloat(data interface{}, bits int) (float64, error) {
	switch v := data.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), bits)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to float%d: %w", v, bits, err)
		}
		return f, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float%d", v, bits)
	}
}
