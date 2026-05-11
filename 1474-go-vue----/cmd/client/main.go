package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerURL string
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	config := &Config{ServerURL: *serverURL}
	client := NewClient(config)

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "fence":
		handleFenceCommand(client, cmdArgs)
	case "bike":
		handleBikeCommand(client, cmdArgs)
	case "ride":
		handleRideCommand(client, cmdArgs)
	case "dispatch":
		handleDispatchCommand(client, cmdArgs)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("BikeShare CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  bike-client [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string    Server URL (default \"http://localhost:8080\")")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  fence list")
	fmt.Println("  fence create <name> <type> <min_lat> <max_lat> <min_lng> <max_lng> <capacity>")
	fmt.Println("  fence get <id>")
	fmt.Println("  fence delete <id>")
	fmt.Println()
	fmt.Println("  bike list")
	fmt.Println("  bike add <lat> <lng>")
	fmt.Println("  bike get <id>")
	fmt.Println("  bike move <id> <lat> <lng>")
	fmt.Println()
	fmt.Println("  ride list [--active]")
	fmt.Println("  ride start <bike_id> <start_lat> <start_lng>")
	fmt.Println("  ride end <ride_id> <end_lat> <end_lng>")
	fmt.Println()
	fmt.Println("  dispatch analyze")
	fmt.Println("  dispatch tasks [--pending]")
	fmt.Println("  dispatch generate")
	fmt.Println("  dispatch complete <task_id>")
}

func handleFenceCommand(client *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: bike-client fence <command>")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		resp, err := client.ListFences()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "create":
		if len(args) < 8 {
			fmt.Println("Usage: bike-client fence create <name> <type> <min_lat> <max_lat> <min_lng> <max_lng> <capacity>")
			os.Exit(1)
		}
		resp, err := client.CreateFence(args[1], args[2], parseFloat(args[3]), parseFloat(args[4]), parseFloat(args[5]), parseFloat(args[6]), parseInt(args[7]))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: bike-client fence get <id>")
			os.Exit(1)
		}
		resp, err := client.GetFence(args[1])
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "delete":
		if len(args) < 2 {
			fmt.Println("Usage: bike-client fence delete <id>")
			os.Exit(1)
		}
		resp, err := client.DeleteFence(args[1])
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	default:
		fmt.Printf("Unknown fence command: %s\n", args[0])
		os.Exit(1)
	}
}

func handleBikeCommand(client *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: bike-client bike <command>")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		resp, err := client.ListBikes()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "add":
		if len(args) < 3 {
			fmt.Println("Usage: bike-client bike add <lat> <lng>")
			os.Exit(1)
		}
		resp, err := client.AddBike(parseFloat(args[1]), parseFloat(args[2]))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: bike-client bike get <id>")
			os.Exit(1)
		}
		resp, err := client.GetBike(args[1])
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "move":
		if len(args) < 4 {
			fmt.Println("Usage: bike-client bike move <id> <lat> <lng>")
			os.Exit(1)
		}
		resp, err := client.MoveBike(args[1], parseFloat(args[2]), parseFloat(args[3]))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	default:
		fmt.Printf("Unknown bike command: %s\n", args[0])
		os.Exit(1)
	}
}

func handleRideCommand(client *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: bike-client ride <command>")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		activeOnly := len(args) > 1 && args[1] == "--active"
		resp, err := client.ListRides(activeOnly)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "start":
		if len(args) < 4 {
			fmt.Println("Usage: bike-client ride start <bike_id> <start_lat> <start_lng>")
			os.Exit(1)
		}
		resp, err := client.StartRide(args[1], parseFloat(args[2]), parseFloat(args[3]))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "end":
		if len(args) < 4 {
			fmt.Println("Usage: bike-client ride end <ride_id> <end_lat> <end_lng>")
			os.Exit(1)
		}
		resp, err := client.EndRide(args[1], parseFloat(args[2]), parseFloat(args[3]))
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	default:
		fmt.Printf("Unknown ride command: %s\n", args[0])
		os.Exit(1)
	}
}

func handleDispatchCommand(client *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: bike-client dispatch <command>")
		os.Exit(1)
	}

	switch args[0] {
	case "analyze":
		resp, err := client.AnalyzeCapacity()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "tasks":
		pendingOnly := len(args) > 1 && args[1] == "--pending"
		resp, err := client.ListDispatchTasks(pendingOnly)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "generate":
		resp, err := client.GenerateDispatchTasks()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	case "complete":
		if len(args) < 2 {
			fmt.Println("Usage: bike-client dispatch complete <task_id>")
			os.Exit(1)
		}
		resp, err := client.CompleteDispatchTask(args[1])
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		printJSON(resp)

	default:
		fmt.Printf("Unknown dispatch command: %s\n", args[0])
		os.Exit(1)
	}
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(data))
}
