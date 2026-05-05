package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"cron-tester/internal/cron"
	"cron-tester/internal/protocol"
)

func handleConnection(conn net.Conn, cache *cron.Cache) {
	defer conn.Close()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Read error: %v\n", err)
		return
	}

	var req protocol.ValidateRequest
	if err := json.Unmarshal(buf[:n], &req); err != nil {
		fmt.Printf("Unmarshal error: %v\n", err)
		return
	}

	var response interface{}

	switch req.Type {
	case protocol.RequestTypeValidate:
		response = handleValidate(req)
	case protocol.RequestTypeNextTime:
		var nextTimeReq protocol.NextTimeRequest
		if err := json.Unmarshal(buf[:n], &nextTimeReq); err != nil {
			fmt.Printf("Unmarshal error: %v\n", err)
			return
		}
		response = handleNextTime(nextTimeReq, cache)
	case protocol.RequestTypeNextNTimes:
		var nTimesReq protocol.NextNTimesRequest
		if err := json.Unmarshal(buf[:n], &nTimesReq); err != nil {
			fmt.Printf("Unmarshal error: %v\n", err)
			return
		}
		response = handleNextNTimes(nTimesReq, cache)
	default:
		response = protocol.ValidateResponse{
			Valid: false,
			Error: fmt.Sprintf("unknown request type: %s", req.Type),
		}
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return
	}

	_, err = conn.Write(respBytes)
	if err != nil {
		fmt.Printf("Write error: %v\n", err)
		return
	}
}

func handleValidate(req protocol.ValidateRequest) protocol.ValidateResponse {
	err := cron.Validate(req.Expr)
	if err != nil {
		return protocol.ValidateResponse{
			Valid: false,
			Error: err.Error(),
		}
	}
	return protocol.ValidateResponse{
		Valid: true,
	}
}

func handleNextTime(req protocol.NextTimeRequest, cache *cron.Cache) protocol.NextTimeResponse {
	expr, err := getOrParseExpr(req.Expr, cache)
	if err != nil {
		return protocol.NextTimeResponse{
			Valid: false,
			Error: err.Error(),
		}
	}

	if expr.WillNeverFire() {
		return protocol.NextTimeResponse{
			Valid:         true,
			WillNeverFire: true,
		}
	}

	now := time.Now()
	nextTime, found := expr.Next(now)

	if !found {
		return protocol.NextTimeResponse{
			Valid:         true,
			WillNeverFire: true,
		}
	}

	execTime := protocol.ExecutionTime{
		Time: nextTime.Format(time.RFC3339),
	}

	if req.Verbose {
		execTime.Delay = nextTime.Sub(now)
	}

	return protocol.NextTimeResponse{
		Valid:    true,
		NextTime: &execTime,
	}
}

func handleNextNTimes(req protocol.NextNTimesRequest, cache *cron.Cache) protocol.NextNTimesResponse {
	expr, err := getOrParseExpr(req.Expr, cache)
	if err != nil {
		return protocol.NextNTimesResponse{
			Valid: false,
			Error: err.Error(),
		}
	}

	if expr.WillNeverFire() {
		return protocol.NextNTimesResponse{
			Valid:         true,
			WillNeverFire: true,
			Count:         0,
		}
	}

	if req.N <= 0 {
		req.N = 10
	}

	now := time.Now()
	times, allFound := expr.NextN(now, req.N)

	if !allFound && len(times) == 0 {
		return protocol.NextNTimesResponse{
			Valid:         true,
			WillNeverFire: true,
			Count:         0,
		}
	}

	var execTimes []protocol.ExecutionTime
	for _, t := range times {
		execTime := protocol.ExecutionTime{
			Time: t.Format(time.RFC3339),
		}

		if req.Verbose {
			execTime.Delay = t.Sub(now)
		}

		execTimes = append(execTimes, execTime)
	}

	return protocol.NextNTimesResponse{
		Valid:         true,
		NextTimes:     execTimes,
		WillNeverFire: !allFound,
		Count:         len(execTimes),
	}
}

func getOrParseExpr(exprStr string, cache *cron.Cache) (*cron.CronExpr, error) {
	if parsedExpr, found := cache.Get(exprStr); found {
		return parsedExpr, nil
	}

	expr, err := cron.Parse(exprStr)
	if err != nil {
		return nil, err
	}

	cache.Set(exprStr, expr)
	return expr, nil
}
