package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	"goroutinelab/api"
)

func main() {
	var fileFlag, serverURLFlag, pidFlag string
	var thresholdFlag, suspectThresholdFlag int

	flag.StringVar(&fileFlag, "file", "", "Path to stack trace file")
	flag.StringVar(&serverURLFlag, "server", "", "Server URL (e.g., http://localhost:8080)")
	flag.StringVar(&pidFlag, "pid", "", "Running process PID to pull profile from")
	flag.IntVar(&thresholdFlag, "threshold", 5, "Wait threshold in minutes")
	flag.IntVar(&suspectThresholdFlag, "suspect-threshold", 1, "Number of suspects to trigger exit code 1")
	flag.Parse()

	var stackTrace string
	var err error

	if pidFlag != "" {
		stackTrace, err = pullFromPID(pidFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error pulling from PID: %v\n", err)
			os.Exit(1)
		}
	} else if fileFlag != "" {
		data, err := os.ReadFile(fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		stackTrace = string(data)
	} else {
		fmt.Fprintln(os.Stderr, "Usage: client -file <path> or -pid <pid>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	req := api.AnalyzeRequest{
		StackTrace: stackTrace,
		Options: &api.Options{
			WaitThresholdMinutes: thresholdFlag,
			SuspectThreshold:     suspectThresholdFlag,
		},
	}

	var report *api.Report

	if serverURLFlag != "" {
		report, err = analyzeRemote(serverURLFlag, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Remote analysis error: %v\n", err)
			os.Exit(1)
		}
	} else {
		report = analyzeLocal(stackTrace, req.Options)
	}

	printReport(report)

	totalSuspects := 0
	for _, g := range report.SuspectGroups {
		totalSuspects += g.Count
	}

	if totalSuspects >= suspectThresholdFlag {
		os.Exit(1)
	}
}

func pullFromPID(pid string) (string, error) {
	cmd := exec.Command("gdb", "-p", pid, "--batch", "--ex", "thread apply all bt")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func analyzeRemote(url string, req api.AnalyzeRequest) (*api.Report, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(strings.TrimRight(url, "/")+"/analyze", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ar api.AnalyzeResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, err
	}

	if !ar.Success {
		return nil, fmt.Errorf("server error: %s", ar.Error)
	}

	return &ar.Report, nil
}

func analyzeLocal(stackTrace string, opts *api.Options) *api.Report {
	a := newLocalAnalyzer(opts)
	return a.Analyze(stackTrace, opts)
}

func printReport(r *api.Report) {
	fmt.Printf("\n=== Goroutine Lab Report ===\n\n")
	fmt.Printf("Total goroutines: %d\n\n", r.TotalGoroutines)

	fmt.Printf("By block type:\n")
	if len(r.ByBlockType) == 0 {
		fmt.Println("  (none)")
	} else {
		keys := make([]string, 0, len(r.ByBlockType))
		for k := range r.ByBlockType {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %-15s: %d\n", k, r.ByBlockType[k])
		}
	}
	fmt.Println()

	if !r.HasSuspicious {
		fmt.Println("✅ No suspicious goroutines detected.")
		return
	}

	totalSuspects := 0
	for _, g := range r.SuspectGroups {
		totalSuspects += g.Count
	}

	fmt.Printf("⚠️  Detected %d suspicious goroutine(s) in %d group(s):\n\n", totalSuspects, len(r.SuspectGroups))

	for i, group := range r.SuspectGroups {
		fmt.Printf("--- Group %d: %s (%d goroutines) ---\n", i+1, group.BlockType, group.Count)
		for j, g := range group.Goroutines {
			fmt.Printf("\n  Goroutine %d (ID: %d, state: %s", j+1, g.ID, g.State)
			if g.HasWaitTime {
				fmt.Printf(", wait: %d min", g.WaitMinutes)
			}
			fmt.Println(")")

			if g.StackTruncated {
				fmt.Println("    [WARNING: stack truncated]")
			}

			if len(g.Reasons) > 0 {
				fmt.Println("    Reasons:")
				for _, reason := range g.Reasons {
					fmt.Printf("      - %s\n", reason)
				}
			}

			if len(g.UserStack) > 0 {
				fmt.Println("    User stack:")
				for _, f := range g.UserStack {
					fmt.Printf("      %s\n", f.Function)
					if f.File != "" {
						fmt.Printf("        %s:%d\n", f.File, f.Line)
					}
				}
			} else {
				fmt.Println("    (no user stack frames found)")
			}
		}
		fmt.Println()
	}
}
