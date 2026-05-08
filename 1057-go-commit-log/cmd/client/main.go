package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/baru/commitlog/api"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) postJSON(path string, req, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := c.client.Post(c.baseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("http %d: %s", httpResp.StatusCode, string(data))
	}

	if resp != nil {
		return json.NewDecoder(httpResp.Body).Decode(resp)
	}
	return nil
}

func (c *Client) Produce(topic string, value []byte) (int64, error) {
	var resp api.ProduceResponse
	err := c.postJSON("/produce", api.ProduceRequest{
		Topic: topic,
		Value: value,
	}, &resp)
	if err != nil {
		return 0, err
	}
	return resp.Offset, nil
}

func (c *Client) Fetch(topic string, offset int64, max int) ([]api.Message, int64, error) {
	var resp api.FetchResponse
	err := c.postJSON("/fetch", api.FetchRequest{
		Topic:  topic,
		Offset: offset,
		Max:    max,
	}, &resp)
	if err != nil {
		return nil, 0, err
	}
	return resp.Messages, resp.Next, nil
}

func (c *Client) CreateGroup(groupID, topic string) error {
	var resp api.CreateGroupResponse
	err := c.postJSON("/groups/create", api.CreateGroupRequest{
		GroupID: groupID,
		Topic:   topic,
	}, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) JoinGroup(groupID, consumerID string) (int64, error) {
	var resp api.JoinGroupResponse
	err := c.postJSON("/groups/join", api.JoinGroupRequest{
		GroupID:    groupID,
		ConsumerID: consumerID,
	}, &resp)
	if err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, fmt.Errorf("%s", resp.Error)
	}
	return resp.AssignedOffset, nil
}

func (c *Client) CommitOffset(groupID, consumerID string, offset int64) error {
	var resp api.CommitOffsetResponse
	err := c.postJSON("/groups/commit", api.CommitOffsetRequest{
		GroupID:    groupID,
		ConsumerID: consumerID,
		Offset:     offset,
	}, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) GetOffset(groupID, consumerID string) (int64, error) {
	var resp api.GetOffsetResponse
	err := c.postJSON("/groups/offset", api.GetOffsetRequest{
		GroupID:    groupID,
		ConsumerID: consumerID,
	}, &resp)
	if err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, fmt.Errorf("%s", resp.Error)
	}
	return resp.Offset, nil
}

func runProduce() {
	flagSet := flag.NewFlagSet("produce", flag.ExitOnError)
	server := flagSet.String("server", "http://localhost:8080", "server address")
	topic := flagSet.String("topic", "", "topic name")
	value := flagSet.String("value", "", "message value (if not provided, read from stdin)")
	interactive := flagSet.Bool("i", false, "interactive mode")
	flagSet.Parse(os.Args[2:])

	if *topic == "" {
		fmt.Println("error: -topic is required")
		os.Exit(1)
	}

	client := NewClient(*server)

	if *interactive {
		fmt.Printf("Interactive produce mode for topic: %s\n", *topic)
		fmt.Println("Enter messages (Ctrl+D to exit):")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			offset, err := client.Produce(*topic, []byte(line))
			if err != nil {
				fmt.Printf("error: %v\n", err)
				continue
			}
			fmt.Printf("produced: offset=%d\n", offset)
		}
		return
	}

	var msgValue []byte
	if *value != "" {
		msgValue = []byte(*value)
	} else {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("error reading stdin: %v\n", err)
			os.Exit(1)
		}
		msgValue = data
	}

	offset, err := client.Produce(*topic, msgValue)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(offset)
}

func runConsume() {
	flagSet := flag.NewFlagSet("consume", flag.ExitOnError)
	server := flagSet.String("server", "http://localhost:8080", "server address")
	topic := flagSet.String("topic", "", "topic name")
	offset := flagSet.Int64("offset", -1, "start offset (only for raw mode)")
	group := flagSet.String("group", "", "consumer group id")
	consumer := flagSet.String("consumer", "c1", "consumer id within group")
	follow := flagSet.Bool("f", false, "follow mode (keep polling)")
	count := flagSet.Int("n", 100, "max messages per fetch")
	flagSet.Parse(os.Args[2:])

	if *topic == "" {
		fmt.Println("error: -topic is required")
		os.Exit(1)
	}

	client := NewClient(*server)

	useGroup := *group != ""
	var currentOffset int64 = 0

	if useGroup {
		if err := client.CreateGroup(*group, *topic); err != nil {
			fmt.Printf("error creating group: %v\n", err)
			os.Exit(1)
		}
		assigned, err := client.JoinGroup(*group, *consumer)
		if err != nil {
			fmt.Printf("error joining group: %v\n", err)
			os.Exit(1)
		}
		currentOffset = assigned
	} else if *offset >= 0 {
		currentOffset = *offset
	}

	for {
		messages, next, err := client.Fetch(*topic, currentOffset, *count)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			if !*follow {
				os.Exit(1)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		for _, msg := range messages {
			fmt.Printf("%d\t%s\n", msg.Offset, string(msg.Value))
		}

		if useGroup && len(messages) > 0 {
			if err := client.CommitOffset(*group, *consumer, next); err != nil {
				fmt.Printf("error committing offset: %v\n", err)
			}
		}

		currentOffset = next

		if !*follow {
			break
		}

		if len(messages) == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: client <command> [options]")
		fmt.Println("commands: produce, consume")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "produce":
		runProduce()
	case "consume":
		runConsume()
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
