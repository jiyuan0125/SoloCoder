package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	Command    string
	InputFile  string
	JSONPath   string
	StatsMode  string
	Format     string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	config := parseArgs()

	switch config.Command {
	case "query":
		runQuery(config)
	case "count":
		runCount(config)
	case "sum":
		runSum(config)
	case "keys":
		runKeys(config)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", config.Command)
		printUsage()
		os.Exit(1)
	}
}

func parseArgs() *Config {
	config := &Config{}
	config.Command = os.Args[1]

	flagSet := flag.NewFlagSet(config.Command, flag.ExitOnError)
	flagSet.StringVar(&config.InputFile, "f", "", "Input file path (default: stdin)")
	flagSet.StringVar(&config.JSONPath, "p", "$", "JSONPath query (default: $)")
	flagSet.StringVar(&config.StatsMode, "m", "count", "Stats mode: count, sum, avg (default: count)")
	flagSet.StringVar(&config.Format, "format", "json", "Output format: json or csv (default: json)")

	if len(os.Args) > 2 {
		flagSet.Parse(os.Args[2:])
	}

	return config
}

func printUsage() {
	fmt.Println("Usage: streamjson [command] [flags]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  query    Query values by JSONPath")
	fmt.Println("  count    Count matching values")
	fmt.Println("  sum      Sum numeric values")
	fmt.Println("  keys     List all keys in JSON")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -f string    Input file path (default: stdin)")
	fmt.Println("  -p string   JSONPath query (default: $)")
	fmt.Println("  -m string   Stats mode: count, sum, avg (default: count)")
	fmt.Println("  --format string   Output format: json or csv (default: json)")
}

func runQuery(config *Config) {
	reader, _, err := OpenInput(config.InputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	jp, err := ParseJSONPath(config.JSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSONPath: %v\n", err)
		os.Exit(1)
	}

	handler := &QueryHandler{
		Results: make([]interface{}, 0),
	}

	processor := NewStreamProcessor(reader, jp, handler)
	if err := processor.Process(); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing: %v\n", err)
		os.Exit(1)
	}

	outputResults(handler.Results, config.Format)
}

func runCount(config *Config) {
	runStats(config, ModeCount)
}

func runSum(config *Config) {
	mode, err := ParseStatsMode(config.StatsMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	runStats(config, mode)
}

func runStats(config *Config, defaultMode StatsMode) {
	reader, _, err := OpenInput(config.InputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	jp, err := ParseJSONPath(config.JSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSONPath: %v\n", err)
		os.Exit(1)
	}

	handler := NewStatsHandler(defaultMode)

	processor := NewStreamProcessor(reader, jp, handler)
	if err := processor.Process(); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing: %v\n", err)
		os.Exit(1)
	}

	outputStats(handler, config.Format)
}

func runKeys(config *Config) {
	reader, _, err := OpenInput(config.InputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	handler := NewKeysHandler()

	processor := NewKeysProcessor(reader, handler)
	if err := processor.Process(); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing: %v\n", err)
		os.Exit(1)
	}

	keys := handler.GetKeys()
	sort.Strings(keys)
	outputKeys(keys, config.Format)
}

func outputResults(results []interface{}, format string) {
	switch strings.ToLower(format) {
	case "csv":
		outputResultsCSV(results)
	default:
		outputResultsJSON(results)
	}
}

func outputResultsJSON(results []interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func outputResultsCSV(results []interface{}) {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	if err := w.Write([]string{"value"}); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	for _, v := range results {
		row := valueToString(v)
		if err := w.Write([]string{row}); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
			os.Exit(1)
		}
	}
}

func outputStats(handler *StatsHandler, format string) {
	switch strings.ToLower(format) {
	case "csv":
		outputStatsCSV(handler)
	default:
		outputStatsJSON(handler)
	}
}

func outputStatsJSON(handler *StatsHandler) {
	result := map[string]interface{}{
		"count": handler.Count(),
		"sum":   handler.Sum(),
		"avg":   handler.Avg(),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func outputStatsCSV(handler *StatsHandler) {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	if err := w.Write([]string{"metric", "value"}); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	if err := w.Write([]string{"count", strconv.FormatInt(handler.Count(), 10)}); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	if err := w.Write([]string{"sum", strconv.FormatFloat(handler.Sum(), 'f', -1, 64)}); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	if err := w.Write([]string{"avg", strconv.FormatFloat(handler.Avg(), 'f', -1, 64)}); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}
}

func outputKeys(keys []string, format string) {
	switch strings.ToLower(format) {
	case "csv":
		outputKeysCSV(keys)
	default:
		outputKeysJSON(keys)
	}
}

func outputKeysJSON(keys []string) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(keys); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func outputKeysCSV(keys []string) {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	if err := w.Write([]string{"key"}); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	for _, k := range keys {
		if err := w.Write([]string{k}); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
			os.Exit(1)
		}
	}
}

func valueToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case bool:
		return strconv.FormatBool(val)
	case nil:
		return "null"
	default:
		bytes, _ := json.Marshal(v)
		return string(bytes)
	}
}
