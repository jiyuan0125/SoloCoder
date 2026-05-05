// Package main demonstrates the usage of the config loader library.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"config-loader/config"
)

// DatabaseConfig represents database connection configuration.
type DatabaseConfig struct {
	Host     string `config:"host,required" default:"localhost"`
	Port     int    `config:"port" default:"5432"`
	Username string `config:"username,required"`
	Password string `config:"password,required"`
	MaxConns int    `config:"max_conns" default:"10"`
	SSLMode  string `config:"ssl_mode" default:"disable"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Host    string `config:"host" default:"0.0.0.0"`
	Port    int    `config:"port" default:"8080"`
	Timeout int    `config:"timeout" default:"30"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Level  string `config:"level" default:"info"`
	Format string `config:"format" default:"json"`
}

// AppConfig represents the complete application configuration.
type AppConfig struct {
	Database DatabaseConfig `config:"database"`
	Server   ServerConfig   `config:"server"`
	Logging  LoggingConfig  `config:"logging"`
	Debug    bool           `config:"debug" default:"false"`
	Features []string       `config:"features"`
}

func main() {
	fmt.Println("=== Config Loader Library Demo ===")
	fmt.Println()

	demoBasicUsage()
	fmt.Println()

	demoPriorityOverride()
	fmt.Println()

	demoTypeConversion()
	fmt.Println()

	demoNestedAccess()
	fmt.Println()

	demoHotReload()
	fmt.Println()

	demoRequiredFields()
	fmt.Println()

	fmt.Println("=== All demos completed ===")
}

func demoBasicUsage() {
	fmt.Println("--- Demo 1: Basic Usage ---")

	loader := config.NewLoader()
	loader.AddFile("examples/config.yaml")

	var cfg AppConfig
	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	printConfig(&cfg)
}

func demoPriorityOverride() {
	fmt.Println("--- Demo 2: Priority Override ---")

	os.Setenv("APP_DATABASE_HOST", "env-host.example.com")
	os.Setenv("APP_SERVER_PORT", "9090")
	defer os.Unsetenv("APP_DATABASE_HOST")
	defer os.Unsetenv("APP_SERVER_PORT")

	loader := config.NewLoader()
	loader.AddFile("examples/config.yaml")
	loader.SetEnvPrefix("APP")
	loader.SetCLIArgs([]string{"--debug=true", "--logging=level=debug"})

	var cfg AppConfig
	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Println("Priority: CLI Args > Env Vars > Config File > Defaults")
	fmt.Printf("Database.Host (from env): %s\n", cfg.Database.Host)
	fmt.Printf("Server.Port (from env): %d\n", cfg.Server.Port)
	fmt.Printf("Debug (from CLI): %v\n", cfg.Debug)
}

func demoTypeConversion() {
	fmt.Println("--- Demo 3: Type Conversion ---")

	cfg := config.New()
	cfg.Set("string_val", "123", config.SourceEnv)
	cfg.Set("bool_val", "true", config.SourceEnv)
	cfg.Set("float_val", "3.14159", config.SourceEnv)
	cfg.Set("int_from_string", "42", config.SourceEnv)

	intVal, _ := cfg.GetInt("string_val")
	boolVal, _ := cfg.GetBool("bool_val")
	floatVal, _ := cfg.GetFloat64("float_val")
	intFromStr, _ := cfg.GetInt("int_from_string")

	fmt.Printf("String \"123\" -> int: %d\n", intVal)
	fmt.Printf("String \"true\" -> bool: %v\n", boolVal)
	fmt.Printf("String \"3.14159\" -> float64: %f\n", floatVal)
	fmt.Printf("String \"42\" -> int: %d\n", intFromStr)

	cfg.Set("invalid_int", "not-a-number", config.SourceCLI)
	_, err := cfg.GetInt("invalid_int")
	if err != nil {
		fmt.Printf("Expected error for invalid conversion: %v\n", err)
	}
}

func demoNestedAccess() {
	fmt.Println("--- Demo 4: Nested Path Access ---")

	cfg := config.New()
	cfg.Set("database.host", "db.example.com", config.SourceFile)
	cfg.Set("database.port", 5432, config.SourceFile)
	cfg.Set("server.host", "0.0.0.0", config.SourceFile)
	cfg.Set("server.port", 8080, config.SourceFile)

	host, _ := cfg.GetString("database.host")
	port, _ := cfg.GetInt("database.port")
	serverPort, _ := cfg.GetInt("server.port")

	fmt.Printf("database.host: %s\n", host)
	fmt.Printf("database.port: %d\n", port)
	fmt.Printf("server.port: %d\n", serverPort)

	all := cfg.All()
	fmt.Printf("All config keys: %v\n", keys(all))
}

func demoHotReload() {
	fmt.Println("--- Demo 5: Hot Reload (Simulated) ---")

	loader := config.NewLoader()

	loader.OnChange(func(oldCfg, newCfg *config.Config, err error) {
		if err != nil {
			fmt.Printf("Reload error: %v\n", err)
			return
		}
		fmt.Println("Configuration changed! New config loaded.")
	})

	fmt.Println("Watch callback registered. In real usage, config file changes would trigger reload.")
	fmt.Println("Watcher uses polling to detect file changes (default: 2 seconds).")
}

func demoRequiredFields() {
	fmt.Println("--- Demo 6: Required Fields Validation ---")

	type TestConfig struct {
		RequiredField string `config:"required_field,required"`
		OptionalField string `config:"optional_field" default:"default_value"`
	}

	loader := config.NewLoader()

	var cfg TestConfig
	err := loader.Load(&cfg)
	if err != nil {
		var missingErr *config.MissingRequiredError
		if ok := errorsAs(err, &missingErr); ok {
			fmt.Printf("Expected error: missing required fields: %v\n", missingErr.Fields)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	loader2 := config.NewLoader()
	loader2.SetCLIArgs([]string{"--required-field=provided"})

	var cfg2 TestConfig
	if err := loader2.Load(&cfg2); err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
		return
	}
	fmt.Printf("Success! RequiredField: %s, OptionalField (default): %s\n",
		cfg2.RequiredField, cfg2.OptionalField)
}

func printConfig(cfg *AppConfig) {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println("Loaded Configuration:")
	fmt.Println(string(data))
}

func keys(m map[string]interface{}) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func errorsAs(err error, target interface{}) bool {
	switch v := target.(type) {
	case **config.MissingRequiredError:
		if mr, ok := err.(*config.MissingRequiredError); ok {
			*v = mr
			return true
		}
	}
	return false
}

func runHotReloadDemo() {
	fmt.Println("\n=== Hot Reload Demo (Interactive) ===")
	fmt.Println("This demo simulates config file watching.")
	fmt.Println("Press Ctrl+C to exit.")

	loader := config.NewLoader()
	loader.AddFile("examples/config.yaml")
	loader.EnableWatch()
	loader.SetWatchInterval(2 * time.Second)

	var cfg AppConfig
	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	loader.OnChange(func(oldCfg, newCfg *config.Config, err error) {
		if err != nil {
			fmt.Printf("\nReload error: %v\n", err)
			return
		}
		fmt.Println("\n=== Configuration Changed ===")
		fmt.Println("Old config (snapshot):")
		oldData, _ := json.MarshalIndent(oldCfg.All(), "", "  ")
		fmt.Println(string(oldData))
		fmt.Println("\nNew config:")
		newData, _ := json.MarshalIndent(newCfg.All(), "", "  ")
		fmt.Println(string(newData))
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nExiting hot reload demo...")

	if w := loader.GetWatcher(); w != nil {
		w.Stop()
	}
}
