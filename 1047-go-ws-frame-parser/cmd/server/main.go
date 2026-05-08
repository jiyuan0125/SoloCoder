package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"wsframe/wsframe"
)

const (
	websocketGUID       = "258EAFA5-E914-47DA-95CA-5AB5FCDB9CDC"
	defaultPort         = ":8080"
	pingInterval        = 30 * time.Second
	pongTimeout         = 10 * time.Second
	maxMessageSize      = 1 << 20
)

type Client struct {
	conn          *wsframe.Connection
	lastPongTime  time.Time
	mu            sync.Mutex
	pingTicker    *time.Ticker
	done          chan struct{}
}

func main() {
	fmt.Printf("WebSocket Echo Server starting on %s\n", defaultPort)
	
	http.HandleFunc("/ws", handleWebSocket)
	
	if err := http.ListenAndServe(defaultPort, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	h, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Web server doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	if !isWebSocketUpgrade(r) {
		http.Error(w, "Not a WebSocket upgrade request", http.StatusBadRequest)
		return
	}

	secWebSocketKey := r.Header.Get("Sec-WebSocket-Key")
	if secWebSocketKey == "" {
		http.Error(w, "Missing Sec-WebSocket-Key header", http.StatusBadRequest)
		return
	}

	acceptKey := computeAcceptKey(secWebSocketKey)

	conn, _, err := h.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := buildUpgradeResponse(acceptKey)
	if _, err := conn.Write([]byte(response)); err != nil {
		conn.Close()
		log.Printf("Failed to write upgrade response: %v", err)
		return
	}

	wsConn := wsframe.NewServerConnection(conn, maxMessageSize)
	client := &Client{
		conn:         wsConn,
		lastPongTime: time.Now(),
		done:         make(chan struct{}),
	}

	go handleClient(client)
}

func isWebSocketUpgrade(r *http.Request) bool {
	upgrade := r.Header.Get("Upgrade")
	connection := r.Header.Get("Connection")
	
	return strings.Contains(strings.ToLower(connection), "upgrade") &&
		strings.ToLower(upgrade) == "websocket"
}

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func buildUpgradeResponse(acceptKey string) string {
	var sb strings.Builder
	sb.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	sb.WriteString("Upgrade: websocket\r\n")
	sb.WriteString("Connection: Upgrade\r\n")
	sb.WriteString(fmt.Sprintf("Sec-WebSocket-Accept: %s\r\n", acceptKey))
	sb.WriteString("\r\n")
	return sb.String()
}

func handleClient(client *Client) {
	defer client.close()

	log.Printf("Client connected: %s", client.conn.RemoteAddr())

	client.pingTicker = time.NewTicker(pingInterval)
	defer client.pingTicker.Stop()

	go client.pingLoop()

	for {
		client.conn.SetReadDeadline(time.Now().Add(pingInterval + pongTimeout))
		msg, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("Read error from %s: %v", client.conn.RemoteAddr(), err)
			return
		}

		switch msg.OpCode {
		case wsframe.OpCodeText:
			log.Printf("Received text message from %s: %s", client.conn.RemoteAddr(), string(msg.Payload))
			if err := client.conn.WriteText(string(msg.Payload)); err != nil {
				log.Printf("Write error: %v", err)
				return
			}

		case wsframe.OpCodeBinary:
			log.Printf("Received binary message from %s: %d bytes", client.conn.RemoteAddr(), len(msg.Payload))
			if err := client.conn.WriteBinary(msg.Payload); err != nil {
				log.Printf("Write error: %v", err)
				return
			}

		case wsframe.OpCodePing:
			log.Printf("Received Ping from %s", client.conn.RemoteAddr())
			if err := client.conn.WritePong(msg.Payload); err != nil {
				log.Printf("Write Pong error: %v", err)
				return
			}

		case wsframe.OpCodePong:
			log.Printf("Received Pong from %s", client.conn.RemoteAddr())
			client.mu.Lock()
			client.lastPongTime = time.Now()
			client.mu.Unlock()

		case wsframe.OpCodeClose:
			code, reason := wsframe.ParseClosePayload(msg.Payload)
			log.Printf("Received Close from %s: code=%d, reason=%s", client.conn.RemoteAddr(), code, reason)
			client.conn.WriteClose(wsframe.CloseNormalClosure, "")
			return
		}
	}
}

func (c *Client) pingLoop() {
	for {
		select {
		case <-c.pingTicker.C:
			c.mu.Lock()
			if time.Since(c.lastPongTime) > pingInterval+pongTimeout {
				c.mu.Unlock()
				log.Printf("Client %s timed out", c.conn.RemoteAddr())
				c.close()
				return
			}
			c.mu.Unlock()

			if err := c.conn.WritePing([]byte("ping")); err != nil {
				log.Printf("Failed to send ping to %s: %v", c.conn.RemoteAddr(), err)
				return
			}

		case <-c.done:
			return
		}
	}
}

func (c *Client) close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}
	
	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}
	
	c.conn.WriteClose(wsframe.CloseNormalClosure, "")
	c.conn.Close()
	log.Printf("Client disconnected: %s", c.conn.RemoteAddr())
}

func readHTTPRequest(conn net.Conn) (*http.Request, error) {
	reader := bufio.NewReader(conn)
	return http.ReadRequest(reader)
}
