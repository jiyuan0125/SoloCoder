package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const defaultServer = "http://localhost:8512"

func main() {
	server := flag.String("server", defaultServer, "external sort server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*server)
	cmd := args[0]

	var err error
	switch cmd {
	case "add":
		if len(args) < 2 {
			log.Fatal("usage: add <filename>")
		}
		err = cmdAdd(client, args[1])
	case "sort":
		err = cmdSort(client)
	case "stats":
		err = cmdStats(client)
	case "config":
		err = cmdConfig(client)
	case "result":
		if len(args) >= 2 {
			err = cmdResultSave(client, args[1])
		} else {
			err = cmdResultShow(client)
		}
	case "memory":
		if len(args) < 2 {
			log.Fatal("usage: memory <limit>")
		}
		err = cmdSetMemory(client, args[1])
	case "ways":
		if len(args) < 2 {
			log.Fatal("usage: ways <count>")
		}
		err = cmdSetMergeWays(client, args[1])
	case "reset":
		err = cmdReset(client)
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func printUsage() {
	fmt.Println("External Sort Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string    server URL (default \"http://localhost:8512\")")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add <filename>        read data from file and add to server")
	fmt.Println("  sort                  trigger external sort")
	fmt.Println("  stats                 show sorting statistics and visualization")
	fmt.Println("  config                show current configuration")
	fmt.Println("  result [filename]     get sorted result, save to file if specified")
	fmt.Println("  memory <limit>        set memory limit (records per chunk)")
	fmt.Println("  ways <count>          set merge ways")
	fmt.Println("  reset                 reset all data and state")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client -server http://localhost:9090 add input.txt")
	fmt.Println("  client memory 1000")
	fmt.Println("  client ways 4")
	fmt.Println("  client sort")
	fmt.Println("  client stats")
	fmt.Println("  client result output.txt")
}

func cmdAdd(client *Client, filename string) error {
	data, err := ReadDataFromFile(filename)
	if err != nil {
		return err
	}

	if err := client.AddData(data); err != nil {
		return err
	}

	fmt.Printf("added %d records\n", len(data))
	return nil
}

func cmdSort(client *Client) error {
	if err := client.Sort(); err != nil {
		return err
	}

	VisualizeSortingStarted()
	return nil
}

func cmdStats(client *Client) error {
	stats, err := client.GetStats()
	if err != nil {
		return err
	}

	VisualizeStats(stats)
	return nil
}

func cmdConfig(client *Client) error {
	config, err := client.GetConfig()
	if err != nil {
		return err
	}

	VisualizeConfig(config)
	return nil
}

func cmdResultShow(client *Client) error {
	result, err := client.GetResult()
	if err != nil {
		return err
	}

	if !result.Success {
		fmt.Printf("error: %s\n", result.Message)
		return nil
	}

	for _, v := range result.Data {
		fmt.Println(v)
	}
	return nil
}

func cmdResultSave(client *Client, filename string) error {
	result, err := client.GetResult()
	if err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	if err := WriteDataToFile(filename, result.Data); err != nil {
		return err
	}

	fmt.Printf("saved %d records to %s\n", len(result.Data), filename)
	return nil
}

func cmdSetMemory(client *Client, limitStr string) error {
	var limit int
	if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
		return fmt.Errorf("invalid memory limit: %s", limitStr)
	}

	if err := client.SetMemoryLimit(limit); err != nil {
		return err
	}

	fmt.Printf("memory limit set to %d\n", limit)
	return nil
}

func cmdSetMergeWays(client *Client, waysStr string) error {
	var ways int
	if _, err := fmt.Sscanf(waysStr, "%d", &ways); err != nil {
		return fmt.Errorf("invalid merge ways: %s", waysStr)
	}

	if err := client.SetMergeWays(ways); err != nil {
		return err
	}

	fmt.Printf("merge ways set to %d\n", ways)
	return nil
}

func cmdReset(client *Client) error {
	if err := client.Reset(); err != nil {
		return err
	}

	fmt.Println("reset complete")
	return nil
}
