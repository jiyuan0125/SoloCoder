package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "strings"

    "batch-rename/protocol"
)

func sendPreviewRequest(writer *bufio.Writer, reader *bufio.Reader, req *protocol.Request) (*protocol.PreviewResponse, error) {
    req.Type = protocol.MsgTypePreview

    data, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    _, err = writer.Write(data)
    if err != nil {
        return nil, err
    }
    err = writer.WriteByte('\n')
    if err != nil {
        return nil, err
    }
    err = writer.Flush()
    if err != nil {
        return nil, err
    }

    respLine, err := reader.ReadString('\n')
    if err != nil {
        return nil, err
    }
    respLine = strings.TrimSpace(respLine)

    var errorResp protocol.ErrorResponse
    if err := json.Unmarshal([]byte(respLine), &errorResp); err == nil && errorResp.Message != "" {
        return nil, fmt.Errorf("%s", errorResp.Message)
    }

    var previewResp protocol.PreviewResponse
    if err := json.Unmarshal([]byte(respLine), &previewResp); err != nil {
        return nil, err
    }

    return &previewResp, nil
}

func sendExecuteRequest(writer *bufio.Writer, reader *bufio.Reader, req *protocol.Request) (*protocol.ExecuteResponse, error) {
    req.Type = protocol.MsgTypeExecute

    data, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    _, err = writer.Write(data)
    if err != nil {
        return nil, err
    }
    err = writer.WriteByte('\n')
    if err != nil {
        return nil, err
    }
    err = writer.Flush()
    if err != nil {
        return nil, err
    }

    respLine, err := reader.ReadString('\n')
    if err != nil {
        return nil, err
    }
    respLine = strings.TrimSpace(respLine)

    var errorResp protocol.ErrorResponse
    if err := json.Unmarshal([]byte(respLine), &errorResp); err == nil && errorResp.Message != "" {
        return nil, fmt.Errorf("%s", errorResp.Message)
    }

    var executeResp protocol.ExecuteResponse
    if err := json.Unmarshal([]byte(respLine), &executeResp); err != nil {
        return nil, err
    }

    return &executeResp, nil
}

func printPreview(resp *protocol.PreviewResponse) {
    if len(resp.Items) == 0 {
        fmt.Println("No files to process")
        return
    }

    fmt.Println("Preview of rename operations:")
    fmt.Println(strings.Repeat("-", 80))

    for i, item := range resp.Items {
        if item.Skipped {
            fmt.Printf("[%d] %s -> %s (SKIPPED: %s)\n", i+1, item.OldName, item.NewName, item.SkipReason)
        } else {
            fmt.Printf("[%d] %s -> %s\n", i+1, item.OldName, item.NewName)
        }
    }

    fmt.Println(strings.Repeat("-", 80))
    fmt.Printf("Total: %d files, Skipped: %d files\n", resp.Total, resp.Skipped)
}

func printExecuteResult(resp *protocol.ExecuteResponse) {
    fmt.Println("\nExecution Result:")
    fmt.Println(strings.Repeat("-", 80))

    for i, item := range resp.Items {
        if item.Skipped {
            fmt.Printf("[%d] %s -> %s (SKIPPED: %s)\n", i+1, item.OldName, item.NewName, item.SkipReason)
        } else {
            fmt.Printf("[%d] %s -> %s (SUCCESS)\n", i+1, item.OldName, item.NewName)
        }
    }

    fmt.Println(strings.Repeat("-", 80))
    fmt.Printf("Success: %d, Failed: %d, Skipped: %d\n", resp.Success, resp.Failed, resp.Skipped)
}

func confirmExecution() bool {
    fmt.Print("\nProceed with rename? (y/n): ")
    reader := bufio.NewReader(os.Stdin)
    response, err := reader.ReadString('\n')
    if err != nil {
        return false
    }
    response = strings.TrimSpace(strings.ToLower(response))
    return response == "y" || response == "yes"
}
