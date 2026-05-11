package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"wsframe/wsframe"
)

const (
	defaultHost       = "localhost:8410"
	defaultPath       = "/ws"
	maxMessageSize    = 1 << 20
)

var (
	host       = flag.String("host", defaultHost, "WebSocket server host:port")
	path       = flag.String("path", defaultPath, "WebSocket server path")
	message    = flag.String("message", "", "Message to send (text mode)")
	binaryFile = flag.String("binary", "", "Binary file to send")
	interactive = flag.Bool("interactive", false, "Interactive mode")
)

func main() {
	flag.Parse()

	if *message == "" && *binaryFile == "" && !*interactive {
		fmt.Println("Usage:")
		fmt.Println("  client -message \"hello\"")
		fmt.Println("  client -binary /path/to/file")
		fmt.Println("  client -interactive")
		flag.PrintDefaults()
		os.Exit(1)
	}

	conn, err := net.Dial("tcp", *host)
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	wsKey := generateSecWebSocketKey()
	upgradeRequest := buildUpgradeRequest(*host, *path, wsKey)
	
	if _, err := conn.Write([]byte(upgradeRequest)); err != nil {
		fmt.Printf("Failed to send upgrade request: %v\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(conn)
	response, err := readUpgradeResponse(reader)
	if err != nil {
		fmt.Printf("Failed to read upgrade response: %v\n", err)
		os.Exit(1)
	}

	if !strings.Contains(response, "101 Switching Protocols") {
		fmt.Printf("Upgrade failed: %s\n", response)
		os.Exit(1)
	}

	wsConn := wsframe.NewClientConnection(conn, maxMessageSize)
	fmt.Println("WebSocket connection established")

	if *interactive {
		runInteractiveMode(wsConn)
	} else if *message != "" {
		runSingleMessageMode(wsConn, *message, false)
	} else if *binaryFile != "" {
		runSingleFileMode(wsConn, *binaryFile)
	}
}

func generateSecWebSocketKey() string {
	key := make([]byte, 16)
	rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}

func buildUpgradeRequest(host, path, key string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("GET %s HTTP/1.1\r\n", path))
	sb.WriteString(fmt.Sprintf("Host: %s\r\n", host))
	sb.WriteString("Upgrade: websocket\r\n")
	sb.WriteString("Connection: Upgrade\r\n")
	sb.WriteString(fmt.Sprintf("Sec-WebSocket-Key: %s\r\n", key))
	sb.WriteString("Sec-WebSocket-Version: 13\r\n")
	sb.WriteString("\r\n")
	return sb.String()
}

func readUpgradeResponse(reader *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		sb.WriteString(line)
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	return sb.String(), nil
}

func runInteractiveMode(wsConn *wsframe.Connection) {
	fmt.Println("Interactive mode. Type messages and press Enter to send. Type 'quit' to exit.")
	
	done := make(chan struct{})
	defer close(done)
	
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				msg, err := wsConn.ReadMessage()
				if err != nil {
					if err != io.EOF {
						fmt.Printf("\nRead error: %v\n", err)
					}
					return
				}

				switch msg.OpCode {
				case wsframe.OpCodeText:
					fmt.Printf("\nReceived: %s\n> ", string(msg.Payload))
				case wsframe.OpCodeBinary:
					fmt.Printf("\nReceived binary: %d bytes\n> ", len(msg.Payload))
				case wsframe.OpCodePing:
					wsConn.WritePong(msg.Payload)
				case wsframe.OpCodePong:
					fmt.Println("\nReceived Pong")
				case wsframe.OpCodeClose:
					fmt.Println("\nServer closed connection")
					return
				}
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		input := scanner.Text()
		if input == "quit" || input == "exit" {
			wsConn.WriteClose(wsframe.CloseNormalClosure, "")
			time.Sleep(100 * time.Millisecond)
			return
		}

		if input != "" {
			if err := wsConn.WriteText(input); err != nil {
				fmt.Printf("Write error: %v\n", err)
				return
			}
		}
		fmt.Print("> ")
	}
}

func runSingleMessageMode(wsConn *wsframe.Connection, message string, isBinary bool) {
	if isBinary {
		if err := wsConn.WriteBinary([]byte(message)); err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
			return
		}
		fmt.Println("Binary message sent")
	} else {
		if err := wsConn.WriteText(message); err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
			return
		}
		fmt.Println("Text message sent")
	}

	for {
		msg, err := wsConn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			return
		}

		switch msg.OpCode {
		case wsframe.OpCodeText:
			fmt.Printf("Received: %s\n", string(msg.Payload))
			wsConn.WriteClose(wsframe.CloseNormalClosure, "")
			return
		case wsframe.OpCodeBinary:
			fmt.Printf("Received binary: %d bytes\n", len(msg.Payload))
			wsConn.WriteClose(wsframe.CloseNormalClosure, "")
			return
		case wsframe.OpCodePing:
			wsConn.WritePong(msg.Payload)
		case wsframe.OpCodePong:
			fmt.Println("Received Pong")
		case wsframe.OpCodeClose:
			fmt.Println("Server closed connection")
			return
		}
	}
}

func runSingleFileMode(wsConn *wsframe.Connection, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
	}

	runSingleMessageMode(wsConn, string(data), true)
}
