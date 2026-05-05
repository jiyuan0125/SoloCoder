package server

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
    "time"

    "github.com/config-diff/internal/comparator"
    "github.com/config-diff/internal/history"
    "github.com/config-diff/internal/parser"
    "github.com/config-diff/internal/protocol"
)

type Server struct {
    host       string
    port       int
    comparator *comparator.Comparator
    history    *history.History
}

func NewServer(host string, port int) *Server {
    return &Server{
        host:       host,
        port:       port,
        comparator: comparator.NewComparator(),
        history:    history.NewHistory(100),
    }
}

func (s *Server) Start() error {
    addr := fmt.Sprintf("%s:%d", s.host, s.port)
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    defer listener.Close()

    fmt.Printf("Config Diff Server listening on %s\n", addr)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Error accepting connection: %v\n", err)
            continue
        }
        go s.handleConnection(conn)
    }
}

func (s *Server) handleConnection(conn net.Conn) {
    defer conn.Close()

    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    line, err := reader.ReadString('\n')
    if err != nil {
        fmt.Printf("Error reading request: %v\n", err)
        return
    }

    var msg protocol.Message
    if err := json.Unmarshal([]byte(line), &msg); err != nil {
        fmt.Printf("Error unmarshaling message: %v\n", err)
        s.sendError(writer, "Invalid request format")
        return
    }

    var resp protocol.CompareResponse
    switch msg.Type {
    case protocol.TypeRequestCompare:
        resp = s.handleCompareRequest(msg)
    default:
        resp = protocol.CompareResponse{
            Success: false,
            Error:   "Unknown message type",
        }
    }

    respMsg, err := protocol.EncodeResponse(resp)
    if err != nil {
        fmt.Printf("Error encoding response: %v\n", err)
        s.sendError(writer, "Internal server error")
        return
    }

    respData, err := json.Marshal(respMsg)
    if err != nil {
        fmt.Printf("Error marshaling response: %v\n", err)
        s.sendError(writer, "Internal server error")
        return
    }

    writer.Write(respData)
    writer.WriteByte('\n')
    writer.Flush()
}

func (s *Server) handleCompareRequest(msg protocol.Message) protocol.CompareResponse {
    req, err := protocol.DecodeRequest(msg)
    if err != nil {
        return protocol.CompareResponse{
            Success: false,
            Error:   "Invalid request payload",
        }
    }

    config1, err := parser.ParseFile(req.File1Path)
    if err != nil {
        return protocol.CompareResponse{
            Success: false,
            Error:   fmt.Sprintf("Error parsing file1: %v", err),
        }
    }

    config2, err := parser.ParseFile(req.File2Path)
    if err != nil {
        return protocol.CompareResponse{
            Success: false,
            Error:   fmt.Sprintf("Error parsing file2: %v", err),
        }
    }

    diffs, identical := s.comparator.Compare(config1, config2, req.IgnoreKeys)

    record := history.HistoryRecord{
        ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
        File1Path:  req.File1Path,
        File2Path:  req.File2Path,
        IgnoreKeys: req.IgnoreKeys,
        Identical:  identical,
        Diffs:      diffs,
        Timestamp:  time.Now(),
    }
    s.history.AddRecord(record)

    return protocol.CompareResponse{
        Success:   true,
        Identical: identical,
        Diffs:     diffs,
    }
}

func (s *Server) sendError(writer *bufio.Writer, errMsg string) {
    errorMsg, err := protocol.EncodeError(errMsg)
    if err != nil {
        return
    }
    errorData, _ := json.Marshal(errorMsg)
    writer.Write(errorData)
    writer.WriteByte('\n')
    writer.Flush()
}
