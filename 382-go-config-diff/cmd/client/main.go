package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "net"
    "os"

    "github.com/config-diff/internal/comparator"
    "github.com/config-diff/internal/protocol"
)

const (
    defaultServerHost = "127.0.0.1"
    defaultServerPort = 8765
)

type ignoreFlags []string

func (i *ignoreFlags) String() string {
    return fmt.Sprintf("%v", *i)
}

func (i *ignoreFlags) Set(value string) error {
    *i = append(*i, value)
    return nil
}

func main() {
    var ignoreList ignoreFlags
    serverHost := flag.String("server-host", defaultServerHost, "Server host")
    serverPort := flag.Int("server-port", defaultServerPort, "Server port")
    help := flag.Bool("h", false, "Show help message")

    flag.Var(&ignoreList, "ignore", "Keys to ignore (can be specified multiple times)")
    flag.Usage = showHelp
    flag.Parse()

    if *help {
        showHelp()
        os.Exit(0)
    }

    args := flag.Args()
    if len(args) != 2 {
        fmt.Fprintln(os.Stderr, "Error: Must provide exactly two configuration files")
        fmt.Fprintln(os.Stderr, "Use -h for help")
        os.Exit(1)
    }

    file1 := args[0]
    file2 := args[1]

    if _, err := os.Stat(file1); os.IsNotExist(err) {
        fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", file1)
        os.Exit(1)
    }
    if _, err := os.Stat(file2); os.IsNotExist(err) {
        fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", file2)
        os.Exit(1)
    }

    resp, err := sendCompareRequest(*serverHost, *serverPort, file1, file2, ignoreList)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if !resp.Success {
        fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
        os.Exit(1)
    }

    comp := comparator.NewComparator()
    if resp.Identical {
        fmt.Println("配置文件一致")
        os.Exit(0)
    }

    formattedDiffs := comp.FormatDiffs(resp.Diffs)
    for _, diff := range formattedDiffs {
        fmt.Println(diff)
    }

    os.Exit(0)
}

func sendCompareRequest(host string, port int, file1, file2 string, ignoreKeys []string) (protocol.CompareResponse, error) {
    addr := fmt.Sprintf("%s:%d", host, port)
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return protocol.CompareResponse{}, fmt.Errorf("failed to connect to server: %v", err)
    }
    defer conn.Close()

    req := protocol.CompareRequest{
        File1Path:  file1,
        File2Path:  file2,
        IgnoreKeys: ignoreKeys,
    }

    msg, err := protocol.EncodeRequest(req)
    if err != nil {
        return protocol.CompareResponse{}, fmt.Errorf("failed to encode request: %v", err)
    }

    msgData, err := json.Marshal(msg)
    if err != nil {
        return protocol.CompareResponse{}, fmt.Errorf("failed to marshal request: %v", err)
    }

    writer := bufio.NewWriter(conn)
    writer.Write(msgData)
    writer.WriteByte('\n')
    writer.Flush()

    reader := bufio.NewReader(conn)
    respLine, err := reader.ReadString('\n')
    if err != nil {
        return protocol.CompareResponse{}, fmt.Errorf("failed to read response: %v", err)
    }

    var respMsg protocol.Message
    if err := json.Unmarshal([]byte(respLine), &respMsg); err != nil {
        return protocol.CompareResponse{}, fmt.Errorf("failed to unmarshal response: %v", err)
    }

    return protocol.DecodeResponse(respMsg)
}

func showHelp() {
    fmt.Println("Config Diff - Compare two configuration files")
    fmt.Println()
    fmt.Println("Usage: config-diff [options] <file1> <file2>")
    fmt.Println()
    fmt.Println("Options:")
    fmt.Println("  -ignore string      Keys to ignore (can be specified multiple times)")
    fmt.Println("  -server-host string Server host (default \"127.0.0.1\")")
    fmt.Println("  -server-port int    Server port (default 8765)")
    fmt.Println("  -h                  Show this help message")
    fmt.Println()
    fmt.Println("Description:")
    fmt.Println("  Compare two configuration files and show the differences.")
    fmt.Println("  Supports key=value format and simple YAML format (flat key:value structure).")
    fmt.Println()
    fmt.Println("Output:")
    fmt.Println("  + key: value   - Key exists in file2 but not in file1 (added)")
    fmt.Println("  - key: value   - Key exists in file1 but not in file2 (removed)")
    fmt.Println("  ~ key: old -> new - Key exists in both files but values differ (modified)")
    fmt.Println()
    fmt.Println("Notes:")
    fmt.Println("  - Numbers and strings are compared equally (8080 == \"8080\")")
    fmt.Println("  - Order of keys does not affect comparison")
    fmt.Println("  - Empty lines and comment lines (#) are ignored")
    fmt.Println("  - Duplicate keys are overridden by later values")
    fmt.Println("  - Requires config-diff-server to be running")
}
