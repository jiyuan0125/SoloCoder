package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"json-extractor/pkg/common"
)

type Client struct {
	host       string
	port       int
	filename   string
	fields     []string
	outputJSON bool
	compact    bool
}

type fieldFlag []string

func (f *fieldFlag) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *fieldFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func NewClientFromArgs() *Client {
	var fields fieldFlag
	outputJSON := flag.Bool("json", false, "Output as JSON format")
	compact := flag.Bool("compact", false, "Output in compact format")
	host := flag.String("host", "127.0.0.1", "Server host")
	port := flag.Int("port", 8080, "Server port")

	flag.Var(&fields, "f", "Field path to extract (can be specified multiple times)")
	flag.Var(&fields, "field", "Field path to extract (can be specified multiple times)")

	flag.Parse()

	args := flag.Args()
	filename := ""
	if len(args) > 0 {
		filename = args[0]
	}

	return &Client{
		host:       *host,
		port:       *port,
		filename:   filename,
		fields:     fields,
		outputJSON: *outputJSON,
		compact:    *compact,
	}
}

func (c *Client) Run() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	useStdin := c.filename == "-" || c.filename == ""

	req := &common.Request{
		Filename:     c.filename,
		Fields:       c.fields,
		OutputJSON:   c.outputJSON,
		Compact:      c.compact,
		HasStdinData: useStdin,
	}

	if err := common.SendRequest(conn, req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if useStdin {
		if err := common.SendStdinData(conn, os.Stdin); err != nil {
			return fmt.Errorf("failed to send stdin data: %w", err)
		}
	}

	resp, err := common.ReceiveResponse(conn)
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	for _, result := range resp.Results {
		fmt.Println(result)
	}

	return nil
}
