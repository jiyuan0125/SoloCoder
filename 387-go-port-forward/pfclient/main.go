package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
	
	"port-forward/protocol"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	
	cmd := strings.ToLower(os.Args[1])
	
	globalFs := flag.NewFlagSet("global", flag.ExitOnError)
	controlPort := globalFs.Int("control-port", protocol.DefaultControlPort, "Control port of the server")
	verbose := globalFs.Bool("verbose", false, "Enable verbose output")
	
	cmdFs := flag.NewFlagSet(cmd, flag.ExitOnError)
	localPort := cmdFs.Int("local", 0, "Local port to forward (for add/remove commands)")
	remoteAddr := cmdFs.String("remote", "", "Remote address (host:port) to forward to (for add command)")
	timeout := cmdFs.Int("timeout", protocol.DefaultTimeout, "Idle timeout in seconds (for add command)")
	bufferSize := cmdFs.Int("buffer-size", protocol.DefaultBufferSize, "Buffer size in bytes (for add command)")
	
	cmdFs.Usage = usage
	
	var cmdArgs []string
	if len(os.Args) > 2 {
		cmdArgs = os.Args[2:]
	}
	
	if err := cmdFs.Parse(cmdArgs); err != nil {
		os.Exit(1)
	}
	
	parseGlobalFlags(globalFs, cmdFs, controlPort, verbose)
	
	switch cmd {
	case "add":
		handleAdd(*localPort, *remoteAddr, *timeout, *bufferSize, *verbose, *controlPort)
	case "remove":
		handleRemove(*localPort, *controlPort)
	case "list":
		handleList(*controlPort)
	case "stats":
		handleStats(*localPort, *verbose, *controlPort)
	case "shutdown":
		handleShutdown(*controlPort)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func parseGlobalFlags(globalFs *flag.FlagSet, cmdFs *flag.FlagSet, controlPort *int, verbose *bool) {
	cmdFs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "control-port":
			if gp := globalFs.Lookup("control-port"); gp != nil {
				gp.Value.Set(f.Value.String())
			}
		case "verbose":
			if gp := globalFs.Lookup("verbose"); gp != nil {
				gp.Value.Set(f.Value.String())
			}
		}
	})
}

func usage() {
	fmt.Println(`Port Forward Client

Usage:
  pfclient [global-options] command [command-options]

Commands:
  add       Add a new port forward rule
  remove    Remove a port forward rule
  list      List all active forward rules
  stats     Show connection statistics
  shutdown  Shut down the server

Global Options:
  --control-port int   Control port of the server (default 57329)
  --verbose            Enable verbose output

Command Options (add):
  --local int          Local port to forward (required)
  --remote string      Remote address (host:port) to forward to (required)
  --timeout int        Idle timeout in seconds (default 300)
  --buffer-size int    Buffer size in bytes (default 4096)

Command Options (remove):
  --local int          Local port to remove (required)

Command Options (stats):
  --local int          Filter stats by local port (optional)

Examples:
  pfclient add --local 3306 --remote 192.168.1.100:3306
  pfclient add --local 6379 --remote redis-server:6379 --timeout 600 --verbose
  pfclient remove --local 3306
  pfclient list
  pfclient stats
  pfclient stats --local 3306 --verbose
  pfclient shutdown`)
}

func connectToServer(controlPort int) (net.Conn, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", controlPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server on port %d: %v", controlPort, err)
	}
	return conn, nil
}

func handleAdd(localPort int, remoteAddr string, timeout int, bufferSize int, verbose bool, controlPort int) {
	if localPort <= 0 || localPort > 65535 {
		fmt.Println("Error: valid --local port is required (1-65535)")
		os.Exit(1)
	}
	
	if remoteAddr == "" {
		fmt.Println("Error: --remote address is required (format: host:port)")
		os.Exit(1)
	}
	
	conn, err := connectToServer(controlPort)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	req := protocol.AddForwardRequest{
		LocalPort:  localPort,
		RemoteAddr: remoteAddr,
		Verbose:    verbose,
		Timeout:    timeout,
		BufferSize: bufferSize,
	}
	
	payload, err := protocol.EncodeJSON(req)
	if err != nil {
		fmt.Printf("Error: failed to encode request: %v\n", err)
		os.Exit(1)
	}
	
	msg := protocol.Message{
		Type:    protocol.MsgTypeAddForward,
		Payload: payload,
	}
	
	if err := protocol.SendMessage(conn, msg); err != nil {
		fmt.Printf("Error: failed to send request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := protocol.ReceiveResponse(conn)
	if err != nil {
		fmt.Printf("Error: failed to receive response: %v\n", err)
		os.Exit(1)
	}
	
	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleRemove(localPort int, controlPort int) {
	if localPort <= 0 || localPort > 65535 {
		fmt.Println("Error: valid --local port is required (1-65535)")
		os.Exit(1)
	}
	
	conn, err := connectToServer(controlPort)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	req := protocol.RemoveForwardRequest{
		LocalPort: localPort,
	}
	
	payload, err := protocol.EncodeJSON(req)
	if err != nil {
		fmt.Printf("Error: failed to encode request: %v\n", err)
		os.Exit(1)
	}
	
	msg := protocol.Message{
		Type:    protocol.MsgTypeRemoveForward,
		Payload: payload,
	}
	
	if err := protocol.SendMessage(conn, msg); err != nil {
		fmt.Printf("Error: failed to send request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := protocol.ReceiveResponse(conn)
	if err != nil {
		fmt.Printf("Error: failed to receive response: %v\n", err)
		os.Exit(1)
	}
	
	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleList(controlPort int) {
	conn, err := connectToServer(controlPort)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	msg := protocol.Message{
		Type: protocol.MsgTypeListForwards,
	}
	
	if err := protocol.SendMessage(conn, msg); err != nil {
		fmt.Printf("Error: failed to send request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := protocol.ReceiveResponse(conn)
	if err != nil {
		fmt.Printf("Error: failed to receive response: %v\n", err)
		os.Exit(1)
	}
	
	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
	
	var rules []protocol.ForwardRule
	if len(resp.Data) > 0 {
		if err := protocol.DecodeJSON(resp.Data, &rules); err != nil {
			fmt.Printf("Error: failed to decode response: %v\n", err)
			os.Exit(1)
		}
	}
	
	if len(rules) == 0 {
		fmt.Println("No active forward rules")
		return
	}
	
	fmt.Println("Active forward rules:")
	fmt.Println("---------------------")
	for _, rule := range rules {
		fmt.Printf("  :%d -> %s\n", rule.LocalPort, rule.RemoteAddr)
	}
}

func handleStats(localPort int, verbose bool, controlPort int) {
	conn, err := connectToServer(controlPort)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	var payload []byte
	if localPort > 0 {
		payload, _ = protocol.EncodeJSON(localPort)
	}
	
	msg := protocol.Message{
		Type:    protocol.MsgTypeGetStats,
		Payload: payload,
	}
	
	if err := protocol.SendMessage(conn, msg); err != nil {
		fmt.Printf("Error: failed to send request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := protocol.ReceiveResponse(conn)
	if err != nil {
		fmt.Printf("Error: failed to receive response: %v\n", err)
		os.Exit(1)
	}
	
	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
	
	var statsList []protocol.ForwardStats
	if len(resp.Data) > 0 {
		if err := protocol.DecodeJSON(resp.Data, &statsList); err != nil {
			fmt.Printf("Error: failed to decode response: %v\n", err)
			os.Exit(1)
		}
	}
	
	if len(statsList) == 0 {
		fmt.Println("No statistics available")
		return
	}
	
	for _, stats := range statsList {
		fmt.Printf("\nForward Rule: :%d -> %s\n", stats.Rule.LocalPort, stats.Rule.RemoteAddr)
		fmt.Printf("  Active connections: %d\n", stats.ActiveConns)
		fmt.Printf("  Total connections:  %d\n", stats.TotalConns)
		fmt.Printf("  Total bytes:        %d\n", stats.TotalBytes)
		
		if verbose && len(stats.Connections) > 0 {
			fmt.Println("\n  Connections:")
			for _, conn := range stats.Connections {
				status := "ACTIVE"
				if !conn.IsActive {
					status = "CLOSED"
				}
				connectedAt := time.Unix(conn.ConnectedAt, 0).Format("2006-01-02 15:04:05")
				fmt.Printf("    [%s] ID=%d %s\n", status, conn.ID, conn.ClientAddr)
				fmt.Printf("      Connected at: %s\n", connectedAt)
				fmt.Printf("      Sent:         %d bytes\n", conn.BytesSent)
				fmt.Printf("      Received:     %d bytes\n", conn.BytesReceived)
			}
		}
	}
}

func handleShutdown(controlPort int) {
	conn, err := connectToServer(controlPort)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	msg := protocol.Message{
		Type: protocol.MsgTypeShutdown,
	}
	
	if err := protocol.SendMessage(conn, msg); err != nil {
		fmt.Printf("Error: failed to send request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := protocol.ReceiveResponse(conn)
	if err != nil {
		fmt.Printf("Error: failed to receive response: %v\n", err)
		os.Exit(1)
	}
	
	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}
