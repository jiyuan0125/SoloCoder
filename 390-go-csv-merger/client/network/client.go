package network

import (
	"fmt"
	"net"

	"csv-merger/client/cli"
	"csv-merger/common"
)

type Client struct {
	host string
	port string
}

func NewClient(host, port string) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c *Client) SendMergeRequest(config *cli.Config) (*common.MergeResponse, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", c.host, c.port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	request := &common.MergeRequest{
		Command:      common.CmdMergeCSV,
		InputFiles:   config.InputFiles,
		OutputFile:   config.OutputFile,
		DedupColumns: config.DedupColumns,
		SortColumn:   config.SortColumn,
		SortOrder:    config.SortOrder,
		CustomHeader: config.CustomHeader,
	}

	if err := common.SendMessage(conn, request); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var response common.MergeResponse
	if err := common.ReceiveMessage(conn, &response); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &response, nil
}

func (c *Client) Ping() (*common.PingResponse, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", c.host, c.port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	request := &common.PingRequest{
		Command: common.CmdPing,
	}

	if err := common.SendMessage(conn, request); err != nil {
		return nil, fmt.Errorf("failed to send ping: %w", err)
	}

	var response common.PingResponse
	if err := common.ReceiveMessage(conn, &response); err != nil {
		return nil, fmt.Errorf("failed to receive pong: %w", err)
	}

	return &response, nil
}
