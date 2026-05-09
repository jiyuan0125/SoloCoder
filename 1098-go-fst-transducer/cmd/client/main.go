package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go-fst-transducer/common"
)

const defaultServer = "http://localhost:8080"

var serverAddr string

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	flagSet := flag.NewFlagSet(cmd, flag.ExitOnError)
	flagSet.StringVar(&serverAddr, "server", defaultServer, "FST server address")

	switch cmd {
	case "build":
		runBuild(flagSet)
	case "search":
		runSearch(flagSet)
	case "prefix":
		runPrefix(flagSet)
	case "bench":
		runBench(flagSet)
	case "compare":
		runCompare(flagSet)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("FST Dictionary Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fst-client [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build    - Build FST from text file (format: key value per line)")
	fmt.Println("  search   - Search for exact key")
	fmt.Println("  prefix   - Search for keys with prefix")
	fmt.Println("  bench    - Run benchmark with sample words")
	fmt.Println("  compare  - Compare memory usage with map")
	fmt.Println("  help     - Show this help message")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server <addr>   Server address (default: http://localhost:8080)")
}

func runBuild(fs *flag.FlagSet) {
	var dictName, filePath string
	fs.StringVar(&dictName, "dict", "default", "Dictionary name")
	fs.StringVar(&filePath, "file", "", "Input file path (required)")
	fs.Parse(os.Args[2:])

	if filePath == "" {
		fmt.Println("Error: -file is required")
		os.Exit(1)
	}

	items, err := readKeyValueFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Building dictionary '%s' with %d entries...\n", dictName, len(items))

	req := common.BuildRequest{
		DictName: dictName,
		Items:    items,
	}

	var resp common.BuildResponse
	if err := httpPost(serverAddr+"/api/dict/build", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success! Built %d entries\n", resp.Count)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

func runSearch(fs *flag.FlagSet) {
	var dictName, key string
	fs.StringVar(&dictName, "dict", "default", "Dictionary name")
	fs.StringVar(&key, "key", "", "Key to search (required)")
	fs.Parse(os.Args[2:])

	if key == "" {
		fmt.Println("Error: -key is required")
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/dict/search?dict_name=%s&key=%s", serverAddr, dictName, key)
	var resp common.SearchResponse
	if err := httpGet(url, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	if resp.Found {
		fmt.Printf("Found: key=%s, value=%d\n", key, resp.Value)
	} else {
		fmt.Printf("Not found: key=%s\n", key)
	}
}

func runPrefix(fs *flag.FlagSet) {
	var dictName, prefix string
	fs.StringVar(&dictName, "dict", "default", "Dictionary name")
	fs.StringVar(&prefix, "prefix", "", "Prefix to search (required)")
	fs.Parse(os.Args[2:])

	if prefix == "" {
		fmt.Println("Error: -prefix is required")
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/dict/prefix?dict_name=%s&prefix=%s", serverAddr, dictName, prefix)
	var resp common.PrefixResponse
	if err := httpGet(url, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("Found %d entries with prefix '%s':\n", len(resp.Items), prefix)
	for _, item := range resp.Items {
		fmt.Printf("  %s = %d\n", item.Key, item.Value)
	}
}

func runBench(fs *flag.FlagSet) {
	var dictName string
	fs.StringVar(&dictName, "dict", "benchmark", "Dictionary name")
	fs.Parse(os.Args[2:])

	sampleWords := generateSampleWords()
	items := make([]common.KeyValue, len(sampleWords))
	for i, word := range sampleWords {
		items[i] = common.KeyValue{
			Key:   word,
			Value: uint64(i),
		}
	}

	fmt.Printf("=== Building FST with %d words ===\n", len(sampleWords))
	start := time.Now()
	req := common.BuildRequest{
		DictName: dictName,
		Items:    items,
	}
	var resp common.BuildResponse
	if err := httpPost(serverAddr+"/api/dict/build", req, &resp); err != nil {
		fmt.Printf("Build error: %v\n", err)
		os.Exit(1)
	}
	buildTime := time.Since(start)
	fmt.Printf("Build time: %v\n", buildTime)
	fmt.Printf("Build rate: %.2f words/second\n", float64(len(sampleWords))/buildTime.Seconds())

	fmt.Println()
	fmt.Println("=== Query Benchmark ===")
	numQueries := 1000
	queryWords := make([]string, numQueries)
	for i := 0; i < numQueries; i++ {
		queryWords[i] = sampleWords[i%len(sampleWords)]
	}

	start = time.Now()
	foundCount := 0
	for _, word := range queryWords {
		url := fmt.Sprintf("%s/api/dict/search?dict_name=%s&key=%s", serverAddr, dictName, word)
		var resp common.SearchResponse
		if err := httpGet(url, &resp); err != nil {
			continue
		}
		if resp.Found {
			foundCount++
		}
	}
	queryTime := time.Since(start)
	fmt.Printf("Queries: %d, Found: %d\n", numQueries, foundCount)
	fmt.Printf("Query time: %v\n", queryTime)
	fmt.Printf("Query rate: %.2f queries/second\n", float64(numQueries)/queryTime.Seconds())
}

func runCompare(fs *flag.FlagSet) {
	fmt.Println("=== Memory Comparison: FST vs Map ===")
	sampleWords := generateSampleWords()

	var m1, m2 runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m1)

	mapData := make(map[string]uint64, len(sampleWords))
	for i, word := range sampleWords {
		mapData[word] = uint64(i)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)
	mapMem := m2.Alloc - m1.Alloc
	fmt.Printf("Map memory usage: %d bytes (%.2f KB)\n", mapMem, float64(mapMem)/1024)

	items := make([]common.KeyValue, len(sampleWords))
	for i, word := range sampleWords {
		items[i] = common.KeyValue{
			Key:   word,
			Value: uint64(i),
		}
	}

	req := common.BuildRequest{
		DictName: "compare",
		Items:    items,
	}
	var resp common.BuildResponse
	if err := httpPost(serverAddr+"/api/dict/build", req, &resp); err != nil {
		fmt.Printf("Build error: %v\n", err)
		os.Exit(1)
	}

	runtime.GC()
	runtime.ReadMemStats(&m1)

	var listResp common.ListResponse
	httpGet(serverAddr+"/api/dict/list", &listResp)
	fmt.Printf("FST built successfully: %d entries\n", len(sampleWords))
	fmt.Printf("Note: FST memory is managed on the server side")
}

func generateSampleWords() []string {
	words := []string{
		"apple", "app", "application", "apply",
		"banana", "band", "bandana",
		"cat", "category", "caterpillar",
		"dog", "dogma", "dogwood",
		"elephant", "elegant", "element",
		"fish", "fisher", "fishing",
		"goat", "goblin", "goal",
		"house", "housing", "household",
		"ice", "icecream", "icon",
		"jump", "jumbo", "jungle",
		"kite", "kitchen", "kitten",
		"lion", "lip", "liquid",
		"monkey", "money", "month",
		"night", "noble", "note",
		"ocean", "octopus", "odd",
		"panda", "pan", "pancake",
		"queen", "quick", "quiet",
		"rabbit", "race", "radio",
		"snake", "snow", "solar",
		"tiger", "time", "tiny",
		"umbrella", "uncle", "uniform",
		"violin", "victory", "village",
		"wolf", "woman", "wonder",
		"xray", "xylophone",
		"yacht", "yellow", "yogurt",
		"zebra", "zero", "zone",
	}

	result := make([]string, 0, len(words)*10)
	for _, word := range words {
		result = append(result, word)
		result = append(result, word+"s")
		result = append(result, word+"ed")
		result = append(result, word+"ing")
		result = append(result, word+"ly")
	}

	return result
}

func readKeyValueFile(filePath string) ([]common.KeyValue, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	items := make([]common.KeyValue, 0)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: invalid format, expected 'key value'", lineNum)
		}

		key := parts[0]
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid value: %v", lineNum, err)
		}

		items = append(items, common.KeyValue{
			Key:   key,
			Value: value,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func httpGet(url string, resp interface{}) error {
	httpResp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error %d: %s", httpResp.StatusCode, string(body))
	}

	return json.Unmarshal(body, resp)
}

func httpPost(url string, req interface{}, resp interface{}) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error %d: %s", httpResp.StatusCode, string(body))
	}

	return json.Unmarshal(body, resp)
}
