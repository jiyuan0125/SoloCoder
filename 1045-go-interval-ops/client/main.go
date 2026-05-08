package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"intervalops/common"
)

func main() {
	cmd := flag.String("cmd", "", "command: merge|difference|intersect|query|batch")
	typ := flag.String("type", "integer", "interval type: integer|time")
	batchID := flag.String("batch", "", "batch id")
	server := flag.String("server", "http://localhost:8080", "server address")
	fileA := flag.String("a", "", "input file A (first set)")
	fileB := flag.String("b", "", "input file B (second set, for difference/intersect)")
	output := flag.String("o", "", "output file")
	point := flag.String("point", "", "query point: integer or time string")
	timezone := flag.String("tz", "UTC", "timezone for time intervals")
	flag.Parse()

	if *cmd == "" {
		fmt.Fprintln(os.Stderr, "error: -cmd is required")
		printUsage()
		os.Exit(1)
	}

	client := NewAPIClient(*server)

	switch strings.ToLower(*cmd) {
	case "merge", "difference", "intersect", "query":
		runOperation(client, *cmd, *typ, *batchID, *fileA, *fileB, *output, *point, *timezone)
	case "batch":
		runQueryBatch(client, *batchID)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %s\n", *cmd)
		os.Exit(1)
	}
}

func runOperation(client *APIClient, cmd, typ, batchID, fileA, fileB, output, point, tz string) {
	req := common.OperationRequest{
		BatchID:      batchID,
		Operation:    common.Operation(strings.ToLower(cmd)),
		IntervalType: common.IntervalType(typ),
		Timezone:     tz,
	}

	if typ == "integer" {
		if fileA == "" {
			fmt.Fprintln(os.Stderr, "error: -a is required for input")
			os.Exit(1)
		}
		intervalsA, err := readIntegerIntervalsFromFile(fileA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file A: %v\n", err)
			os.Exit(1)
		}
		req.IntegersA = intervalsA

		if cmd == "difference" || cmd == "intersect" {
			if fileB == "" {
				fmt.Fprintln(os.Stderr, "error: -b is required for difference/intersect")
				os.Exit(1)
			}
			intervalsB, err := readIntegerIntervalsFromFile(fileB)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading file B: %v\n", err)
				os.Exit(1)
			}
			req.IntegersB = intervalsB
		}

		if cmd == "query" {
			if point == "" {
				fmt.Fprintln(os.Stderr, "error: -point is required for query")
				os.Exit(1)
			}
			p, err := strconv.Atoi(point)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing integer point: %v\n", err)
				os.Exit(1)
			}
			req.IntegerPoint = &p
		}
	} else if typ == "time" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading timezone: %v\n", err)
			os.Exit(1)
		}
		if fileA == "" {
			fmt.Fprintln(os.Stderr, "error: -a is required for input")
			os.Exit(1)
		}
		intervalsA, err := readTimeIntervalsFromFile(fileA, loc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file A: %v\n", err)
			os.Exit(1)
		}
		req.TimesA = intervalsA

		if cmd == "difference" || cmd == "intersect" {
			if fileB == "" {
				fmt.Fprintln(os.Stderr, "error: -b is required for difference/intersect")
				os.Exit(1)
			}
			intervalsB, err := readTimeIntervalsFromFile(fileB, loc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading file B: %v\n", err)
				os.Exit(1)
			}
			req.TimesB = intervalsB
		}

		if cmd == "query" {
			if point == "" {
				fmt.Fprintln(os.Stderr, "error: -point is required for query")
				os.Exit(1)
			}
			p, err := parseTime(point, loc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing time point: %v\n", err)
				os.Exit(1)
			}
			req.TimePoint = &p
		}
	} else {
		fmt.Fprintf(os.Stderr, "error: unknown type %s\n", typ)
		os.Exit(1)
	}

	resp, err := client.Operate(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}

	if typ == "integer" {
		printIntegerResult(resp.Result.Integers)
		if output != "" {
			if err := writeIntegerIntervalsToFile(output, resp.Result.Integers); err != nil {
				fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("result written to %s\n", output)
		}
	} else {
		loc := time.UTC
		if tz != "" {
			loc, _ = time.LoadLocation(tz)
		}
		printTimeResult(resp.Result.Times, loc)
		if output != "" {
			if err := writeTimeIntervalsToFile(output, resp.Result.Times, loc); err != nil {
				fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("result written to %s\n", output)
		}
	}
}

func runQueryBatch(client *APIClient, batchID string) {
	if batchID == "" {
		fmt.Fprintln(os.Stderr, "error: -batch is required for batch query")
		os.Exit(1)
	}
	resp, err := client.QueryBatch(batchID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
		os.Exit(1)
	}
	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}
	if !resp.Found {
		fmt.Printf("batch %s not found\n", batchID)
		return
	}
	data := resp.Data
	fmt.Printf("batch: %s\n", data.BatchID)
	fmt.Printf("operation: %s\n", data.Request.Operation)
	fmt.Printf("interval_type: %s\n", data.Request.IntervalType)
	fmt.Printf("created_at: %s\n", data.CreatedAt.Format(time.RFC3339))
	fmt.Println("result:")
	if data.Request.IntervalType == common.TypeInteger {
		printIntegerResult(data.Result.Integers)
	} else {
		loc := time.UTC
		if data.Request.Timezone != "" {
			loc, _ = time.LoadLocation(data.Request.Timezone)
		}
		printTimeResult(data.Result.Times, loc)
	}
}

func printIntegerResult(intervals []common.IntegerInterval) {
	if len(intervals) == 0 {
		fmt.Println("(empty)")
		return
	}
	for _, iv := range intervals {
		fmt.Printf("[%d, %d]\n", iv.Min, iv.Max)
	}
}

func printTimeResult(intervals []common.TimeInterval, loc *time.Location) {
	if len(intervals) == 0 {
		fmt.Println("(empty)")
		return
	}
	for _, iv := range intervals {
		start := iv.Start.In(loc).Format(time.RFC3339)
		end := iv.End.In(loc).Format(time.RFC3339)
		fmt.Printf("[%s, %s]\n", start, end)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client -cmd merge -type integer -a intervals.txt [-o result.txt] [-batch batch001]")
	fmt.Println("  client -cmd difference -type time -a intervalsA.txt -b intervalsB.txt -tz Asia/Shanghai")
	fmt.Println("  client -cmd intersect -type integer -a a.txt -b b.txt")
	fmt.Println("  client -cmd query -type integer -a intervals.txt -point 5")
	fmt.Println("  client -cmd batch -batch batch001")
}
