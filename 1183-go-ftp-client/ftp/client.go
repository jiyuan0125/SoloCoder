package ftp

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Mode int

const (
	ModePassive Mode = iota
	ModeActive
)

const (
	defaultTimeout = 30 * time.Second
	defaultPort    = 21
)

type Client struct {
	host     string
	port     int
	username string
	password string
	mode     Mode

	controlConn net.Conn
	reader      *bufio.Reader
	writer      *bufio.Writer

	currentDir string
	connected  bool

	mu sync.Mutex
}

type Response struct {
	Code int
	Msg  string
	Lines []string
}

func NewClient(host string, port int, username, password string, mode Mode) *Client {
	if port == 0 {
		port = defaultPort
	}
	if mode != ModeActive {
		mode = ModePassive
	}
	return &Client{
		host:     host,
		port:     port,
		username: username,
		password: password,
		mode:     mode,
	}
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", addr, defaultTimeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.controlConn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)

	resp, err := c.readResponse()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read welcome message: %w", err)
	}
	if resp.Code/100 != 2 {
		conn.Close()
		return fmt.Errorf("unexpected welcome response: %d %s", resp.Code, resp.Msg)
	}

	err = c.login()
	if err != nil {
		conn.Close()
		return err
	}

	err = c.setBinaryMode()
	if err != nil {
		conn.Close()
		return err
	}

	dir, err := c.getCurrentDir()
	if err != nil {
		c.currentDir = "/"
	} else {
		c.currentDir = dir
	}

	c.connected = true
	return nil
}

func (c *Client) login() error {
	resp, err := c.sendCommand("USER %s", c.username)
	if err != nil {
		return fmt.Errorf("USER command failed: %w", err)
	}

	if resp.Code == 331 {
		resp, err = c.sendCommand("PASS %s", c.password)
		if err != nil {
			return fmt.Errorf("PASS command failed: %w", err)
		}
	}

	if resp.Code/100 != 2 {
		return fmt.Errorf("login failed: %d %s", resp.Code, resp.Msg)
	}

	return nil
}

func (c *Client) setBinaryMode() error {
	resp, err := c.sendCommand("TYPE I")
	if err != nil {
		return fmt.Errorf("TYPE I command failed: %w", err)
	}
	if resp.Code/100 != 2 {
		return fmt.Errorf("failed to set binary mode: %d %s", resp.Code, resp.Msg)
	}
	return nil
}

func (c *Client) getCurrentDir() (string, error) {
	resp, err := c.sendCommand("PWD")
	if err != nil {
		return "", err
	}
	if resp.Code/100 != 2 {
		return "", fmt.Errorf("PWD failed: %d %s", resp.Code, resp.Msg)
	}

	start := strings.Index(resp.Msg, "\"")
	if start == -1 {
		return "/", nil
	}
	end := strings.LastIndex(resp.Msg, "\"")
	if end == -1 || end <= start {
		return "/", nil
	}
	return resp.Msg[start+1 : end], nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.sendCommand("QUIT")
	err := c.controlConn.Close()
	c.connected = false
	c.controlConn = nil
	c.reader = nil
	c.writer = nil
	return err
}

func (c *Client) CurrentDir() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentDir
}

func (c *Client) ensureConnected() error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (c *Client) resolvePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return cleanPath(path)
	}
	if c.currentDir == "/" {
		return cleanPath("/" + path)
	}
	return cleanPath(c.currentDir + "/" + path)
}

func cleanPath(path string) string {
	parts := strings.Split(path, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
		case "..":
			if len(result) > 0 {
				result = result[:len(result)-1]
			}
		default:
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return "/"
	}
	return "/" + strings.Join(result, "/")
}

func (c *Client) sendCommand(format string, args ...interface{}) (*Response, error) {
	if err := c.sendRawCommand(fmt.Sprintf(format, args...)); err != nil {
		return nil, err
	}
	return c.readResponse()
}

func (c *Client) sendRawCommand(cmd string) error {
	if c.writer == nil {
		return fmt.Errorf("writer not initialized")
	}
	_, err := c.writer.WriteString(cmd + "\r\n")
	if err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *Client) readResponse() (*Response, error) {
	if c.reader == nil {
		return nil, fmt.Errorf("reader not initialized")
	}

	resp := &Response{}
	first := true

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")

		resp.Lines = append(resp.Lines, line)

		if first {
			if len(line) < 3 {
				return nil, fmt.Errorf("invalid response: %s", line)
			}
			code, err := strconv.Atoi(line[:3])
			if err != nil {
				return nil, fmt.Errorf("invalid response code: %s", line)
			}
			resp.Code = code

			if len(line) > 4 && line[3] == '-' {
				first = false
				continue
			} else if len(line) > 3 {
				resp.Msg = line[4:]
			}
			break
		} else {
			if len(line) >= 4 && line[3] == ' ' {
				if code, err := strconv.Atoi(line[:3]); err == nil && code == resp.Code {
					resp.Msg = line[4:]
					break
				}
			}
		}
	}

	return resp, nil
}

func (c *Client) executeCommand(format string, args ...interface{}) (*Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	if err := c.sendRawCommand(fmt.Sprintf(format, args...)); err != nil {
		return nil, err
	}

	return c.readResponse()
}
