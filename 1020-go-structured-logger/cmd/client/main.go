package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"structured-logger/pkg/api"
)

const defaultServerURL = "http://localhost:8080"

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/")}
}

func (c *Client) request(method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) WriteLog(level, message string, fields map[string]any) error {
	req := api.WriteLogRequest{
		Level:   api.LogLevel(strings.ToUpper(level)),
		Message: message,
		Fields:  fields,
	}
	var resp api.WriteLogResponse
	if err := c.request(http.MethodPost, "/log/write", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *Client) SetLevel(level string) error {
	req := api.UpdateLevelRequest{
		Level: api.LogLevel(strings.ToUpper(level)),
	}
	var resp api.UpdateLevelResponse
	if err := c.request(http.MethodPut, "/log/level", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *Client) GetConfig() (api.GetConfigResponse, error) {
	var config api.GetConfigResponse
	err := c.request(http.MethodGet, "/log/config", nil, &config)
	return config, err
}

func (c *Client) GetRecentLogs(n int) ([]api.LogEntry, error) {
	path := "/log/recent?n=" + url.QueryEscape(strconv.Itoa(n))
	var resp api.GetRecentLogsResponse
	if err := c.request(http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	return resp.Logs, nil
}

func parseFields(args []string) map[string]any {
	fields := make(map[string]any)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			fields[key] = intVal
		} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			fields[key] = floatVal
		} else if boolVal, err := strconv.ParseBool(value); err == nil {
			fields[key] = boolVal
		} else {
			fields[key] = value
		}
	}
	return fields
}

func printUsage() {
	fmt.Println(`Usage: log-client [global-options] <command> [command-options]

Global options:
  -server URL    Server URL (default: http://localhost:8080)

Commands:
  write [options] <message>    Write a log entry
    -level LEVEL    Log level (DEBUG, INFO, WARN, ERROR) (default: INFO)
    key=value...    Additional fields

  level <LEVEL>    Set log level (DEBUG, INFO, WARN, ERROR)
  
  config           Show current configuration
  
  tail [options]   Show recent logs
    -n COUNT       Number of logs to show (default: 100)
`)
}

func main() {
	serverFlag := flag.String("server", defaultServerURL, "Server URL")
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverFlag)
	command := args[0]

	switch command {
	case "write":
		writeCmd := flag.NewFlagSet("write", flag.ExitOnError)
		levelFlag := writeCmd.String("level", "INFO", "Log level")
		writeCmd.Parse(args[1:])
		writeArgs := writeCmd.Args()
		if len(writeArgs) == 0 {
			fmt.Println("Error: message is required")
			os.Exit(1)
		}
		message := writeArgs[0]
		var fields map[string]any
		if len(writeArgs) > 1 {
			fields = parseFields(writeArgs[1:])
		}
		if err := client.WriteLog(*levelFlag, message, fields); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Log written successfully")

	case "level":
		if len(args) < 2 {
			fmt.Println("Error: level is required (DEBUG, INFO, WARN, ERROR)")
			os.Exit(1)
		}
		level := args[1]
		if err := client.SetLevel(level); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Log level set to %s\n", strings.ToUpper(level))

	case "config":
		config, err := client.GetConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Level:        %s\n", config.Level)
		fmt.Printf("OutputToStd:  %v\n", config.OutputToStd)
		fmt.Printf("OutputToFile: %v\n", config.OutputToFile)
		fmt.Printf("MaxFileSize:  %d bytes\n", config.MaxFileSize)
		fmt.Printf("MaxBackups:   %d\n", config.MaxBackups)
		if len(config.FieldLimits) > 0 {
			fmt.Println("FieldLimits:")
			for k, v := range config.FieldLimits {
				fmt.Printf("  %s: %d\n", k, v)
			}
		}

	case "tail":
		tailCmd := flag.NewFlagSet("tail", flag.ExitOnError)
		nFlag := tailCmd.Int("n", 100, "Number of logs")
		tailCmd.Parse(args[1:])
		logs, err := client.GetRecentLogs(*nFlag)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for i := len(logs) - 1; i >= 0; i-- {
			entry := logs[i]
			fieldsJSON := ""
			if entry.Fields != nil {
				data, _ := json.Marshal(entry.Fields)
				fieldsJSON = " " + string(data)
			}
			fmt.Printf("%s [%s] %s:%d - %s%s\n",
				entry.Timestamp.Format("2006-01-02T15:04:05.000000000Z07:00"),
				entry.Level,
				entry.File,
				entry.Line,
				entry.Message,
				fieldsJSON,
			)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
