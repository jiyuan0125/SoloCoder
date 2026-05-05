package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

const DefaultPort = "9876"

type Conn struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func Dial(addr string) (*Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewConn(conn), nil
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) SendRequest(req *Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if err := c.writeFrame(data); err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *Conn) SendResponse(resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if err := c.writeFrame(data); err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *Conn) ReceiveRequest() (*Request, error) {
	data, err := c.readFrame()
	if err != nil {
		return nil, err
	}
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (c *Conn) ReceiveResponse() (*Response, error) {
	data, err := c.readFrame()
	if err != nil {
		return nil, err
	}
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Conn) writeFrame(data []byte) error {
	_, err := fmt.Fprintf(c.writer, "%d\n", len(data))
	if err != nil {
		return err
	}
	_, err = c.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Conn) readFrame() ([]byte, error) {
	var length int
	_, err := fmt.Fscanln(c.reader, &length)
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read frame length: %v", err)
	}
	if length <= 0 {
		return nil, fmt.Errorf("invalid frame length: %d", length)
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(c.reader, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read frame data: %v", err)
	}
	return buf, nil
}
