package main

import (
	"astar-pathfinding/common"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func printUsage() {
	fmt.Println("A* Pathfinding Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  map-file -f <file> [-s <start>] [-g <goal>] [-8]")
	fmt.Println("             Set map from file (0=passable, 1=obstacle)")
	fmt.Println("  random   -w <width> -h <height> -d <density> [-seed <n>] [-8]")
	fmt.Println("             Generate random map")
	fmt.Println("  show")
	fmt.Println("             Show current map")
	fmt.Println("  pathfind [-s <start>] [-g <goal>] [-8]")
	fmt.Println("             Run pathfinding on current map")
	fmt.Println("  result")
	fmt.Println("             Show last pathfinding result")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server <addr>  Server address (default: localhost:8080)")
	fmt.Println("  -s <x,y>        Start point (e.g., -s 0,0)")
	fmt.Println("  -g <x,y>        Goal point (e.g., -g 9,9)")
	fmt.Println("  -8              Enable 8-directional movement (octile)")
	fmt.Println("  -w <n>          Map width (for random)")
	fmt.Println("  -h <n>          Map height (for random)")
	fmt.Println("  -d <0.0-1.0>    Obstacle density (for random)")
	fmt.Println("  -seed <n>       Random seed (for random)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client random -w 10 -h 10 -d 0.2 -8")
	fmt.Println("  client pathfind")
	fmt.Println("  client result")
	fmt.Println("  client map-file -f map.txt -s 0,0 -g 9,9")
	os.Exit(1)
}

func parsePoint(s string) (common.Point, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return common.Point{}, fmt.Errorf("invalid point format: %s (expected x,y)", s)
	}

	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return common.Point{}, fmt.Errorf("invalid x coordinate: %s", parts[0])
	}

	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return common.Point{}, fmt.Errorf("invalid y coordinate: %s", parts[1])
	}

	return common.Point{X: x, Y: y}, nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}

	command := os.Args[1]
	args := os.Args[2:]

	fs := flag.NewFlagSet(command, flag.ExitOnError)
	server := fs.String("server", "localhost:8080", "server address")

	switch command {
	case "map-file":
		fileFlag := fs.String("f", "", "map file path")
		startFlag := fs.String("s", "", "start point x,y")
		goalFlag := fs.String("g", "", "goal point x,y")
		octileFlag := fs.Bool("8", false, "enable 8-directional movement")
		fs.Parse(args)

		if *fileFlag == "" {
			fmt.Println("Error: -f flag is required for map-file command")
			printUsage()
		}

		grid, err := ReadMapFile(*fileFlag)
		if err != nil {
			fmt.Printf("Error reading map file: %v\n", err)
			os.Exit(1)
		}

		start := common.Point{X: 0, Y: 0}
		if *startFlag != "" {
			start, err = parsePoint(*startFlag)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		goal := common.Point{X: len(grid[0]) - 1, Y: len(grid) - 1}
		if *goalFlag != "" {
			goal, err = parsePoint(*goalFlag)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		c := NewClient(*server)
		req := &common.SetMapRequest{
			Map:    grid,
			Width:  len(grid[0]),
			Height: len(grid),
			Start:  start,
			Goal:   goal,
			Octile: *octileFlag,
		}

		if err := c.SetMap(req); err != nil {
			fmt.Printf("Error setting map: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Map set successfully!")

		mapInfo, err := c.GetMap()
		if err != nil {
			fmt.Printf("Error getting map: %v\n", err)
			os.Exit(1)
		}
		DisplayMapInfo(mapInfo)

	case "random":
		widthFlag := fs.Int("w", 0, "map width")
		heightFlag := fs.Int("h", 0, "map height")
		densityFlag := fs.Float64("d", 0.0, "obstacle density (0.0-1.0)")
		seedFlag := fs.Int64("seed", 0, "random seed")
		octileFlag := fs.Bool("8", false, "enable 8-directional movement")
		fs.Parse(args)

		if *widthFlag <= 0 || *heightFlag <= 0 {
			fmt.Println("Error: -w and -h flags are required and must be positive")
			printUsage()
		}
		if *densityFlag < 0 || *densityFlag > 1 {
			fmt.Println("Error: -d flag must be between 0 and 1")
			os.Exit(1)
		}

		c := NewClient(*server)
		req := &common.RandomMapRequest{
			Width:           *widthFlag,
			Height:          *heightFlag,
			ObstacleDensity: *densityFlag,
			Octile:          *octileFlag,
			Seed:            *seedFlag,
		}

		mapInfo, err := c.RandomMap(req)
		if err != nil {
			fmt.Printf("Error generating random map: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Random map generated!")
		DisplayMapInfo(mapInfo)

	case "show":
		fs.Parse(args)
		c := NewClient(*server)

		mapInfo, err := c.GetMap()
		if err != nil {
			fmt.Printf("Error getting map: %v\n", err)
			os.Exit(1)
		}
		DisplayMapInfo(mapInfo)

	case "pathfind":
		startFlag := fs.String("s", "", "start point x,y")
		goalFlag := fs.String("g", "", "goal point x,y")
		octileFlag := fs.Bool("8", false, "enable 8-directional movement")
		fs.Parse(args)

		c := NewClient(*server)

		mapInfo, err := c.GetMap()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Set a map first using 'map-file' or 'random' command")
			os.Exit(1)
		}

		start := common.Point{}
		if *startFlag != "" {
			start, err = parsePoint(*startFlag)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		goal := common.Point{}
		if *goalFlag != "" {
			goal, err = parsePoint(*goalFlag)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		req := &common.RunPathfindingRequest{
			Start:  start,
			Goal:   goal,
			Octile: *octileFlag,
		}

		resp, err := c.Pathfind(req)
		if err != nil {
			fmt.Printf("Error running pathfinding: %v\n", err)
			os.Exit(1)
		}

		DisplayResult(mapInfo.Map, mapInfo.Start, mapInfo.Goal, resp)

	case "result":
		fs.Parse(args)
		c := NewClient(*server)

		mapInfo, err := c.GetMap()
		if err != nil {
			fmt.Printf("Error getting map: %v\n", err)
			os.Exit(1)
		}

		resp, err := c.GetResult()
		if err != nil {
			fmt.Printf("Error getting result: %v\n", err)
			os.Exit(1)
		}

		DisplayResult(mapInfo.Map, mapInfo.Start, mapInfo.Goal, resp)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}
