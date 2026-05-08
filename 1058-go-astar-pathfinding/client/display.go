package main

import (
	"astar-pathfinding/common"
	"fmt"
	"strings"
)

func RenderMap(grid [][]int, start, goal common.Point, path []common.Point, explore []common.Point) string {
	if len(grid) == 0 {
		return "(empty map)"
	}

	height := len(grid)
	width := len(grid[0])

	isPath := make(map[common.Point]bool)
	for _, p := range path {
		isPath[p] = true
	}

	isExplored := make(map[common.Point]int)
	for i, p := range explore {
		isExplored[p] = i
	}

	var builder strings.Builder

	builder.WriteString("Map Legend:\n")
	builder.WriteString("  . = passable\n")
	builder.WriteString("  # = obstacle\n")
	builder.WriteString("  S = start\n")
	builder.WriteString("  G = goal\n")
	builder.WriteString("  * = path\n")
	builder.WriteString("  o = explored (not on path)\n")
	builder.WriteString("\n")

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pt := common.Point{X: x, Y: y}

			if pt == start {
				builder.WriteString("S")
			} else if pt == goal {
				builder.WriteString("G")
			} else if isPath[pt] {
				builder.WriteString("*")
			} else if _, explored := isExplored[pt]; explored {
				builder.WriteString("o")
			} else if grid[y][x] == 1 {
				builder.WriteString("#")
			} else {
				builder.WriteString(".")
			}

			if x < width-1 {
				builder.WriteString(" ")
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func DisplayResult(grid [][]int, start, goal common.Point, resp *common.PathfindingResponse) {
	fmt.Println()
	fmt.Println("============================================================")
	if resp.Success {
		fmt.Println("SUCCESS: Path found!")
		fmt.Printf("Total cost: %.3f\n", resp.TotalCost)
		fmt.Printf("Path length: %d steps\n", len(resp.Path))
		fmt.Printf("Nodes explored: %d\n", len(resp.ExploreOrder))
	} else {
		fmt.Println("FAILED: No path found")
		fmt.Printf("Nodes explored: %d\n", len(resp.ExploreOrder))
	}
	fmt.Println("============================================================")
	fmt.Println()

	path := resp.Path
	if path == nil {
		path = []common.Point{}
	}

	explore := resp.ExploreOrder
	if explore == nil {
		explore = []common.Point{}
	}

	fmt.Println(RenderMap(grid, start, goal, path, explore))

	if resp.Success && len(resp.Path) > 0 {
		fmt.Print("Path: ")
		for i, p := range resp.Path {
			if i > 0 {
				fmt.Print(" -> ")
			}
			fmt.Printf("(%d,%d)", p.X, p.Y)
		}
		fmt.Println()
	}
}

func DisplayMapInfo(resp *common.GetMapResponse) {
	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println("Current Map:")
	fmt.Printf("  Dimensions: %d x %d\n", resp.Width, resp.Height)
	fmt.Printf("  Start: (%d, %d)\n", resp.Start.X, resp.Start.Y)
	fmt.Printf("  Goal: (%d, %d)\n", resp.Goal.X, resp.Goal.Y)
	fmt.Printf("  Movement: ")
	if resp.Octile {
		fmt.Println("8-direction (octile)")
	} else {
		fmt.Println("4-direction")
	}
	fmt.Println("============================================================")
	fmt.Println()

	fmt.Println(RenderMap(resp.Map, resp.Start, resp.Goal, []common.Point{}, []common.Point{}))
}
