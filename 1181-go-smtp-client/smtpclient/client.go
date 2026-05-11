package smtpclient

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
	From     string
}

type Client struct {
	conn       net.Conn
	text       *textproto.Conn
	config     *Config
	extensions map[string]string
	serverName string
}

func NewClient(cfg *Config) *Client {
	return &Client{
		config:     cfg,
		extensions: make(map[string]string),
	}
}

func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	
	var err error
	c.conn, err = net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.text = textproto.NewConn(c.conn)

	_, _, err = c.readResponse(220)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to read greeting: %w", err)
	}

	if err := c.ehlo(); err != nil {
		c.Close()
		return err
	}

	if c.config.UseTLS {
		if err := c.startTLS(); err != nil {
			c.Close()
			return err
		}
	}

	if c.config.Username != "" && c.config.Password != "" {
		if err := c.authenticate(); err != nil {
			c.Close()
			return err
		}
	}

	return nil
}

func (c *Client) ehlo() error {
	_, lines, err := c.sendCommand(250, "EHLO localhost")
	if err != nil {
		return fmt.Errorf("EHLO failed: %w", err)
	}

	c.extensions = make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		key := strings.ToUpper(parts[0])
		if len(parts) == 2 {
			c.extensions[key] = parts[1]
		} else {
			c.extensions[key] = ""
		}
	}

	return nil
}

func (c *Client) startTLS() error {
	_, ok := c.extensions["STARTTLS"]
	if !ok {
		return nil
	}

	_, _, err := c.sendCommand(220, "STARTTLS")
	if err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	tlsConfig := &tls.Config{
		ServerName:         c.config.Host,
		InsecureSkipVerify: true,
	}

	tlsConn := tls.Client(c.conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("TLS handshake failed: %w", err)
	}

	c.conn = tlsConn
	c.text = textproto.NewConn(tlsConn)

	return c.ehlo()
}

func (c *Client) authenticate() error {
	authStr, ok := c.extensions["AUTH"]
	if !ok {
		return fmt.Errorf("server does not support AUTH")
	}

	mechs := strings.Split(authStr, " ")

	var authErr error
	for _, mech := range mechs {
		switch strings.ToUpper(mech) {
		case "PLAIN":
			authErr = c.authPlain()
		case "LOGIN":
			authErr = c.authLogin()
		case "CRAM-MD5":
			authErr = c.authCRAMMD5()
		default:
			continue
		}
		if authErr == nil {
			return nil
		}
	}

	if authErr != nil {
		return authErr
	}
	return fmt.Errorf("no supported authentication mechanism found: %s", authStr)
}

func (c *Client) authPlain() error {
	resp := "\x00" + c.config.Username + "\x00" + c.config.Password
	encoded := encodeBase64([]byte(resp))
	
	_, _, err := c.sendCommand(235, "AUTH PLAIN %s", encoded)
	return err
}

func (c *Client) authLogin() error {
	_, _, err := c.sendCommand(334, "AUTH LOGIN")
	if err != nil {
		return err
	}

	_, _, err = c.sendCommand(334, encodeBase64([]byte(c.config.Username)))
	if err != nil {
		return err
	}

	_, _, err = c.sendCommand(235, encodeBase64([]byte(c.config.Password)))
	return err
}

func (c *Client) authCRAMMD5() error {
	_, lines, err := c.sendCommand(334, "AUTH CRAM-MD5")
	if err != nil {
		return err
	}

	challenge := lines[0]
	decoded, _ := decodeBase64(challenge)
	
	mac := hmacMD5([]byte(c.config.Password), []byte(decoded))
	resp := c.config.Username + " " + fmt.Sprintf("%x", mac)
	encoded := encodeBase64([]byte(resp))

	_, _, err = c.sendCommand(235, encoded)
	return err
}

func (c *Client) Supports8BITMIME() bool {
	_, ok := c.extensions["8BITMIME"]
	return ok
}

func (c *Client) Send(msg *Message) error {
	from := msg.From
	if from == "" {
		from = c.config.From
	}

	cmd := fmt.Sprintf("MAIL FROM:<%s>", from)
	if c.Supports8BITMIME() && msg.HasNonASCII() {
		cmd += " BODY=8BITMIME"
	}

	_, _, err := c.sendCommand(250, cmd)
	if err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	allRecipients := make([]string, 0, len(msg.To)+len(msg.Cc)+len(msg.Bcc))
	allRecipients = append(allRecipients, msg.To...)
	allRecipients = append(allRecipients, msg.Cc...)
	allRecipients = append(allRecipients, msg.Bcc...)

	for _, rcpt := range allRecipients {
		_, _, err := c.sendCommand(250, "RCPT TO:<%s>", rcpt)
		if err != nil {
			return fmt.Errorf("RCPT TO failed for %s: %w", rcpt, err)
		}
	}

	_, _, err = c.sendCommand(354, "DATA")
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	data, err := msg.Build(c.Supports8BITMIME())
	if err != nil {
		c.text.PrintfLine(".")
		return fmt.Errorf("failed to build message: %w", err)
	}

	w := c.text.DotWriter()
	if _, err := w.Write(data); err != nil {
		w.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	_, _, err = c.readResponse(250)
	return err
}

func (c *Client) Close() error {
	if c.text != nil {
		c.sendCommand(221, "QUIT")
		c.text.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

func (c *Client) sendCommand(expectedCode int, format string, args ...interface{}) (int, []string, error) {
	if err := c.text.PrintfLine(format, args...); err != nil {
		return 0, nil, err
	}
	return c.readResponse(expectedCode)
}

func (c *Client) readResponse(expectedCode int) (int, []string, error) {
	code, lines, err := c.readResponseLines()
	if err != nil {
		return 0, nil, err
	}

	if code != expectedCode {
		return code, lines, &textproto.Error{Code: code, Msg: strings.Join(lines, "; ")}
	}
	return code, lines, nil
}

func (c *Client) readResponseLines() (int, []string, error) {
	var lines []string
	code := 0

	for {
		line, err := c.text.ReadLine()
		if err != nil {
			if err == io.EOF && len(lines) > 0 {
				break
			}
			return 0, nil, err
		}

		if len(line) < 4 {
			continue
		}

		codeStr := line[:3]
		sep := line[3]

		var lineCode int
		if c, err := strconv.Atoi(codeStr); err == nil {
			lineCode = c
		}

		if code == 0 {
			code = lineCode
		}

		var msg string
		if len(line) > 4 {
			msg = line[4:]
		}
		lines = append(lines, msg)

		if sep == ' ' {
			break
		}
	}

	return code, lines, nil
}
