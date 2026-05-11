package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fifo-ipc/pkg/fifo"
)

func usage() {
	fmt.Fprintf(os.Stderr, `FIFO IPC Client

Usage: %s [global-options] <command> [command-options]

Commands:
  write <message>           Write a single message to the FIFO
  batch <msg1,msg2,...>     Write multiple messages to the FIFO
  flush [count]             Test concurrent write integrity (default 1000 messages)

Global Options:
  -fifo string              Path to FIFO pipe (default "/tmp/fifo-ipc.pipe")

Examples:
  %s write "hello world"
  %s batch "msg1,msg2,msg3"
  %s flush 10000
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	fifoPath := flag.String("fifo", "/tmp/fifo-ipc.pipe", "path to fifo pipe")
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}

	args := flag.Args()
	command := args[0]

	writer, err := fifo.NewWriter(*fifoPath)
	if err != nil {
		log.Fatalf("failed to create writer: %v", err)
	}
	defer writer.Close()

	switch command {
	case "write":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: write command requires a message argument")
			usage()
		}
		message := args[1]
		if err := writer.WriteMessage([]byte(message)); err != nil {
			log.Fatalf("failed to write message: %v", err)
		}
		fmt.Printf("wrote: %s\n", message)

	case "batch":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: batch command requires comma-separated messages")
			usage()
		}
		messages := strings.Split(args[1], ",")
		success := 0
		failed := 0
		for i, msg := range messages {
			msg = strings.TrimSpace(msg)
			if msg == "" {
				continue
			}
			if err := writer.WriteMessage([]byte(msg)); err != nil {
				log.Printf("message %d failed: %v", i+1, err)
				failed++
			} else {
				success++
			}
		}
		fmt.Printf("batch complete: %d success, %d failed\n", success, failed)

	case "flush":
		count := 1000
		if len(args) >= 2 {
			fmt.Sscanf(args[1], "%d", &count)
		}
		concurrency := 10
		if len(args) >= 3 {
			fmt.Sscanf(args[2], "%d", &concurrency)
		}

		runFlushTest(writer, count, concurrency)

	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n", command)
		usage()
	}
}

func runFlushTest(writer *fifo.Writer, count int, concurrency int) {
	fmt.Printf("starting flush test: %d messages, %d concurrency\n", count, concurrency)
	start := time.Now()

	var successCount int64
	var failedCount int64
	var wg sync.WaitGroup

	messagesPerGoroutine := count / concurrency
	remainder := count % concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		numMessages := messagesPerGoroutine
		if i < remainder {
			numMessages++
		}

		go func(id, n int) {
			defer wg.Done()
			for j := 0; j < n; j++ {
				msg := fmt.Sprintf("flush-test-worker-%d-msg-%d", id, j)
				if err := writer.WriteMessage([]byte(msg)); err != nil {
					atomic.AddInt64(&failedCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(i, numMessages)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("flush test complete:\n")
	fmt.Printf("  duration: %s\n", duration)
	fmt.Printf("  success:  %d\n", successCount)
	fmt.Printf("  failed:   %d\n", failedCount)
	fmt.Printf("  rate:     %.2f msg/s\n", float64(successCount)/duration.Seconds())
}
