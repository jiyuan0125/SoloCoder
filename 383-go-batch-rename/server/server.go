package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
    "sync"

    "batch-rename/protocol"
)

type Server struct {
    addr     string
    listener net.Listener
    wg       sync.WaitGroup
    stopCh   chan struct{}
}

func NewServer(addr string) (*Server, error) {
    return &Server{
        addr:   addr,
        stopCh: make(chan struct{}),
    }, nil
}

func (s *Server) Start() error {
    ln, err := net.Listen("tcp", s.addr)
    if err != nil {
        return err
    }
    s.listener = ln

    for {
        select {
        case <-s.stopCh:
            return nil
        default:
        }

        conn, err := ln.Accept()
        if err != nil {
            select {
            case <-s.stopCh:
                return nil
            default:
                fmt.Printf("Accept error: %v\n", err)
                continue
            }
        }

        s.wg.Add(1)
        go s.handleConnection(conn)
    }
}

func (s *Server) Stop() {
    close(s.stopCh)
    if s.listener != nil {
        s.listener.Close()
    }
    s.wg.Wait()
}

func (s *Server) handleConnection(conn net.Conn) {
    defer s.wg.Done()
    defer conn.Close()

    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    for {
        select {
        case <-s.stopCh:
            return
        default:
        }

        line, err := reader.ReadString('\n')
        if err != nil {
            return
        }

        var req protocol.Request
        if err := json.Unmarshal([]byte(line), &req); err != nil {
            s.sendError(writer, fmt.Sprintf("Invalid request: %v", err))
            continue
        }

        switch req.Type {
        case protocol.MsgTypePreview:
            s.handlePreview(writer, &req)
        case protocol.MsgTypeExecute:
            s.handleExecute(writer, &req)
        case protocol.MsgTypeHistory:
            s.handleHistory(writer, &req)
        default:
            s.sendError(writer, fmt.Sprintf("Unknown request type: %s", req.Type))
        }
    }
}

func (s *Server) sendError(writer *bufio.Writer, message string) {
    resp := protocol.ErrorResponse{Message: message}
    data, _ := json.Marshal(resp)
    writer.Write(data)
    writer.WriteByte('\n')
    writer.Flush()
}

func (s *Server) handlePreview(writer *bufio.Writer, req *protocol.Request) {
    if len(req.Rules) == 0 {
        s.sendError(writer, "At least one rule is required")
        return
    }

    files, err := collectFiles(req)
    if err != nil {
        s.sendError(writer, fmt.Sprintf("Failed to collect files: %v", err))
        return
    }

    if len(files) == 0 {
        resp := protocol.PreviewResponse{
            Items:   []protocol.RenameItem{},
            Total:   0,
            Skipped: 0,
        }
        s.sendPreviewResponse(writer, &resp)
        return
    }

    items, err := generatePreview(files, req.Rules)
    if err != nil {
        s.sendError(writer, fmt.Sprintf("Failed to generate preview: %v", err))
        return
    }

    skipped := 0
    for _, item := range items {
        if item.Skipped {
            skipped++
        }
    }

    resp := protocol.PreviewResponse{
        Items:   items,
        Total:   len(items),
        Skipped: skipped,
    }

    s.sendPreviewResponse(writer, &resp)
}

func (s *Server) sendPreviewResponse(writer *bufio.Writer, resp *protocol.PreviewResponse) {
    data, _ := json.Marshal(resp)
    writer.Write(data)
    writer.WriteByte('\n')
    writer.Flush()
}

func (s *Server) handleExecute(writer *bufio.Writer, req *protocol.Request) {
    if len(req.Rules) == 0 {
        s.sendError(writer, "At least one rule is required")
        return
    }

    files, err := collectFiles(req)
    if err != nil {
        s.sendError(writer, fmt.Sprintf("Failed to collect files: %v", err))
        return
    }

    if len(files) == 0 {
        resp := protocol.ExecuteResponse{
            Items:   []protocol.RenameItem{},
            Success: 0,
            Failed:  0,
            Skipped: 0,
        }
        s.sendExecuteResponse(writer, &resp)
        return
    }

    items, err := generatePreview(files, req.Rules)
    if err != nil {
        s.sendError(writer, fmt.Sprintf("Failed to generate preview: %v", err))
        return
    }

    if req.DryRun {
        success := 0
        failed := 0
        skipped := 0
        for _, item := range items {
            if item.Skipped {
                skipped++
            } else {
                success++
            }
        }

        resp := protocol.ExecuteResponse{
            Items:   items,
            Success: success,
            Failed:  failed,
            Skipped: skipped,
        }
        s.sendExecuteResponse(writer, &resp)
        return
    }

    resultItems, success, failed, skipped := executeRenames(items)

    resp := protocol.ExecuteResponse{
        Items:   resultItems,
        Success: success,
        Failed:  failed,
        Skipped: skipped,
    }

    AddHistoryRecord(resultItems, success, failed, skipped)

    s.sendExecuteResponse(writer, &resp)
}

func (s *Server) sendExecuteResponse(writer *bufio.Writer, resp *protocol.ExecuteResponse) {
    data, _ := json.Marshal(resp)
    writer.Write(data)
    writer.WriteByte('\n')
    writer.Flush()
}

func (s *Server) handleHistory(writer *bufio.Writer, req *protocol.Request) {
    records := GetAllHistoryRecords()

    resp := protocol.HistoryResponse{
        Records: records,
        Total:   len(records),
    }

    s.sendHistoryResponse(writer, &resp)
}

func (s *Server) sendHistoryResponse(writer *bufio.Writer, resp *protocol.HistoryResponse) {
    data, _ := json.Marshal(resp)
    writer.Write(data)
    writer.WriteByte('\n')
    writer.Flush()
}
