package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pingtool/pkg/model"
)

func main() {
	var serverURL string
	var target string
	var count int
	var interval time.Duration
	var ttl int

	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Ping server URL")
	flag.IntVar(&count, "c", 0, "Number of packets to send (0 for infinite)")
	flag.DurationVar(&interval, "i", 1*time.Second, "Interval between packets")
	flag.IntVar(&ttl, "t", 128, "TTL value")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: ping-cli [options] <target>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	target = flag.Arg(0)

	client := NewPingClient(serverURL)

	fmt.Printf("PING %s (%s):\n", target, target)

	taskID, err := client.StartPing(target, count, interval, ttl)
	if err != nil {
		fmt.Printf("Error starting ping: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task ID: %s\n", taskID)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	var lastSent int = -1
	packetLoss := make(map[int]bool)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				status, err := client.GetStatus(taskID)
				if err != nil {
					continue
				}

				displayCurrentResults(status, &lastSent, packetLoss)

				if !status.IsRunning {
					close(done)
					return
				}

				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	select {
	case <-sigChan:
		fmt.Println("\nStopping ping task...")
		client.StopTask(taskID)
		time.Sleep(100 * time.Millisecond)
		status, _ := client.GetStatus(taskID)
		printSummary(status)
	case <-done:
		status, _ := client.GetStatus(taskID)
		printSummary(status)
	}
}

func displayCurrentResults(status *model.TaskStatus, lastSent *int, packetLoss map[int]bool) {
	for i := *lastSent + 1; i < status.PacketsSent; i++ {
		_ = i
	}

	if status.PacketsSent > *lastSent {
		*lastSent = status.PacketsSent
	}
}

func printSummary(status *model.TaskStatus) {
	fmt.Println("")
	fmt.Println("--- ping statistics ---")
	fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n",
		status.PacketsSent,
		status.PacketsReceived,
		status.PacketLoss)

	if status.PacketsReceived > 0 {
		minMs := float64(status.MinRTT) / 1e6
		avgMs := float64(status.AvgRTT) / 1e6
		maxMs := float64(status.MaxRTT) / 1e6
		stdDevMs := float64(status.StdDevRTT) / 1e6

		fmt.Printf("rtt min/avg/max/stddev = %.3f/%.3f/%.3f/%.3f ms\n",
			minMs, avgMs, maxMs, stdDevMs)
	}

	if len(status.TimeExceeded) > 0 {
		fmt.Println("\nTime exceeded messages:")
		for _, te := range status.TimeExceeded {
			fmt.Printf("  %d: From %s icmp_seq=%d Time to live exceeded\n",
				te.Sequence, te.RouterIP, te.Sequence)
		}
	}
}
