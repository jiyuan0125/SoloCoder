package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"disruptor/common"
)

const (
	defaultServer = "http://localhost:8080"
)

type config struct {
	server string
}

func getConfig() config {
	cfg := config{}
	flag.StringVar(&cfg.server, "server", "", "server address")
	flag.Parse()

	if cfg.server == "" {
		if envServer := os.Getenv("DISRUPTOR_SERVER"); envServer != "" {
			cfg.server = envServer
		} else {
			cfg.server = defaultServer
		}
	}

	return cfg
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := getConfig()
	cmd := os.Args[1]

	os.Args = append(os.Args[:1], os.Args[2:]...)

	switch cmd {
	case "create":
		handleCreate(cfg)
	case "destroy":
		handleDestroy(cfg)
	case "publish":
		handlePublish(cfg)
	case "consume":
		handleConsume(cfg)
	case "batch-publish":
		handleBatchPublish(cfg)
	case "batch-consume":
		handleBatchConsume(cfg)
	case "info":
		handleInfo(cfg)
	case "list":
		handleList(cfg)
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: disruptor-client [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create -name <name> -size <size>  Create a new buffer")
	fmt.Println("  destroy -name <name>              Destroy a buffer")
	fmt.Println("  publish -name <name> -msg <msg>   Publish a message")
	fmt.Println("  consume -name <name> -id <id>     Consume a message")
	fmt.Println("  batch-publish -name <name> -msgs <msg1,msg2>  Batch publish messages")
	fmt.Println("  batch-consume -name <name> -id <id> -count <n>  Batch consume messages")
	fmt.Println("  info -name <name>                 Get buffer info")
	fmt.Println("  list                              List all buffers")
	fmt.Println()
	fmt.Println("Global options:")
	fmt.Println("  -server <address>                 Server address (default: http://localhost:8080)")
}

func httpPost(cfg config, path string, body interface{}, resp interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := strings.TrimSuffix(cfg.server, "/") + path
	httpResp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d", httpResp.StatusCode)
	}

	if resp != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, resp)
	}

	return nil
}

func httpGet(cfg config, path string, resp interface{}) error {
	url := strings.TrimSuffix(cfg.server, "/") + path
	httpResp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d", httpResp.StatusCode)
	}

	if resp != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, resp)
	}

	return nil
}

func handleCreate(cfg config) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	name := fs.String("name", "", "buffer name")
	size := fs.Int64("size", 1024, "buffer size")
	fs.Parse(os.Args[1:])

	if *name == "" {
		fmt.Println("error: -name is required")
		os.Exit(1)
	}

	req := common.CreateBufferRequest{Name: *name, Size: *size}
	var resp common.CreateBufferResponse

	if err := httpPost(cfg, "/buffer/create", req, &resp); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Buffer created: %s, capacity: %d\n", resp.Name, resp.Capacity)
}

func handleDestroy(cfg config) {
	fs := flag.NewFlagSet("destroy", flag.ExitOnError)
	name := fs.String("name", "", "buffer name")
	fs.Parse(os.Args[1:])

	if *name == "" {
		fmt.Println("error: -name is required")
		os.Exit(1)
	}

	req := common.DestroyBufferRequest{Name: *name}

	if err := httpPost(cfg, "/buffer/destroy", req, nil); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Buffer destroyed: %s\n", *name)
}

func handlePublish(cfg config) {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	name := fs.String("name", "", "buffer name")
	msg := fs.String("msg", "", "message")
	fs.Parse(os.Args[1:])

	if *name == "" {
		fmt.Println("error: -name is required")
		os.Exit(1)
	}
	if *msg == "" {
		fmt.Println("error: -msg is required")
		os.Exit(1)
	}

	req := common.PublishRequest{BufferName: *name, Message: *msg}
	var resp common.PublishResponse

	if err := httpPost(cfg, "/buffer/publish", req, &resp); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Published at sequence: %d\n", resp.Sequence)
}

func handleConsume(cfg config) {
	fs := flag.NewFlagSet("consume", flag.ExitOnError)
	name := fs.String("name", "", "buffer name")
	consumerID := fs.String("id", "", "consumer ID")
	fs.Parse(os.Args[1:])

	if *name == "" {
		fmt.Println("error: -name is required")
		os.Exit(1)
	}
	if *consumerID == "" {
		fmt.Println("error: -id is required")
		os.Exit(1)
	}

	req := common.ConsumeRequest{BufferName: *name, ConsumerID: *consumerID}
	var resp common.ConsumeResponse

	err := httpPost(cfg, "/buffer/consume", req, &resp)
	if err != nil {
		if strings.Contains(err.Error(), "no data available") {
			fmt.Println("暂无数据")
			return
		}
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sequence: %d\nMessage: %s\n", resp.Sequence, resp.Message)
}

func handleBatchPublish(cfg config) {
	fs := flag.NewFlagSet("batch-publish", flag.ExitOnError)
	name := fs.String("name", "", "buffer name")
	msgsStr := fs.String("msgs", "", "messages (comma-separated)")
	fs.Parse(os.Args[1:])

	if *name == "" {
		fmt.Println("error: -name is required")
		os.Exit(1)
	}
	if *msgsStr == "" {
		fmt.Println("error: -msgs is required")
		os.Exit(1)
	}

	msgs := strings.Split(*msgsStr, ",")
	for i := range msgs {
		msgs[i] = strings.TrimSpace(msgs[i])
	}

	req := common.BatchPublishRequest{BufferName: *name, Messages: msgs}
	var resp common.BatchPublishResponse

	if err := httpPost(cfg, "/buffer/batch-publish", req, &resp); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Published %d messages starting at sequence: %d\n", resp.Count, resp.StartSequence)
}

func handleBatchConsume(cfg config) {
	fs := flag.NewFlagSet("batch-consume", flag.ExitOnError)
	name := fs.String("name", "", "buffer name")
	consumerID := fs.String("id", "", "consumer ID")
	count := fs.Int("count", 10, "max number of messages to consume")
	fs.Parse(os.Args[1:])

	if *name == "" {
		fmt.Println("error: -name is required")
		os.Exit(1)
	}
	if *consumerID == "" {
		fmt.Println("error: -id is required")
		os.Exit(1)
	}

	req := common.BatchConsumeRequest{BufferName: *name, ConsumerID: *consumerID, MaxCount: *count}
	var resp common.BatchConsumeResponse

	err := httpPost(cfg, "/buffer/batch-consume", req, &resp)
	if err != nil {
		if strings.Contains(err.Error(), "no data available") {
			fmt.Println("暂无数据")
			return
		}
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Consumed %d messages:\n", len(resp.Messages))
	for i, msg := range resp.Messages {
		fmt.Printf("[%d] seq=%d: %s\n", i+1, msg.Sequence, msg.Message)
	}
}

func handleInfo(cfg config) {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	name := fs.String("name", "", "buffer name")
	fs.Parse(os.Args[1:])

	if *name == "" {
		fmt.Println("error: -name is required")
		os.Exit(1)
	}

	var resp common.BufferInfoResponse
	path := fmt.Sprintf("/buffer/info?name=%s", *name)

	if err := httpGet(cfg, path, &resp); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Buffer: %s\n", resp.Name)
	fmt.Printf("Capacity: %d\n", resp.Capacity)
	fmt.Printf("Used: %d\n", resp.Used)
	fmt.Printf("Producer cursor: %d\n", resp.ProducerCursor)
	fmt.Println("Consumers:")
	for id, seq := range resp.Consumers {
		fmt.Printf("  %s: %d\n", id, seq)
	}
}

func handleList(cfg config) {
	var resp common.ListBufferResponse

	if err := httpGet(cfg, "/buffer/list", &resp); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Buffers) == 0 {
		fmt.Println("No buffers")
		return
	}

	fmt.Println("Buffers:")
	for _, name := range resp.Buffers {
		fmt.Printf("  %s\n", name)
	}
}
