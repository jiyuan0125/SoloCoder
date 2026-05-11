package main

import (
	"fmt"
	"strings"

	"github.com/example/externalsort/pkg/api"
)

func VisualizeStats(stats *api.StatsResponse) {
	fmt.Println("========================================")
	fmt.Println("          External Sort Stats          ")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("  Total Records:    %d\n", stats.TotalRecords)
	fmt.Printf("  Memory Limit:     %d records per chunk\n", stats.MemoryLimit)
	fmt.Printf("  Merge Ways:       %d-way merge\n", stats.MergeWays)
	fmt.Printf("  Chunks Created:   %d\n", stats.ChunksCreated)
	fmt.Printf("  Merge Passes:     %d\n", stats.MergePasses)
	fmt.Printf("  Disk Reads:       %d\n", stats.DiskReads)
	fmt.Printf("  Disk Writes:      %d\n", stats.DiskWrites)
	fmt.Printf("  Status:           %s\n", getStatusString(stats.IsSorted))
	fmt.Println()

	if len(stats.ChunkSizes) > 0 {
		fmt.Println("---------- Chunk Sizes ----------")
		for i, size := range stats.ChunkSizes {
			bar := generateBar(size, maxInt(stats.ChunkSizes...), 30)
			fmt.Printf("  Chunk %2d: %s (%d records)\n", i+1, bar, size)
		}
		fmt.Println()
	}

	if stats.MergeDetails != nil && len(stats.MergeDetails) > 1 {
		fmt.Println("------- Merge Process Visualization -------")
		for passIdx, pass := range stats.MergeDetails {
			if passIdx == 0 {
				fmt.Printf("  Pass 0 (Initial):   ")
			} else {
				fmt.Printf("  Pass %d (Merged):   ", passIdx)
			}
			for i, size := range pass {
				if i > 0 {
					fmt.Print(" + ")
				}
				bar := generateBar(size, maxInt(flattenInt(stats.MergeDetails)...), 10)
				fmt.Printf("[%s(%d)]", bar, size)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	fmt.Println("========================================")
}

func VisualizeSortingStarted() {
	fmt.Println("========================================")
	fmt.Println("      External Sort Started           ")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("  Sorting process has been triggered.")
	fmt.Println("  Use 'stats' command to check progress.")
	fmt.Println("  Use 'result' command to get sorted data.")
	fmt.Println()
	fmt.Println("========================================")
}

func VisualizeConfig(config *api.GetConfigResponse) {
	fmt.Println("========================================")
	fmt.Println("         Current Configuration         ")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("  Memory Limit:     %d records per chunk\n", config.MemoryLimit)
	fmt.Printf("  Merge Ways:       %d-way merge\n", config.MergeWays)
	fmt.Println()
	fmt.Println("========================================")
}

func getStatusString(isSorted bool) string {
	if isSorted {
		return "✓ Sorted"
	}
	return "○ In Progress / Not Started"
}

func generateBar(value int, maxValue int, width int) string {
	if maxValue == 0 {
		return strings.Repeat(" ", width)
	}
	filled := int(float64(value) / float64(maxValue) * float64(width))
	if filled < 1 && value > 0 {
		filled = 1
	}
	return strings.Repeat("█", filled) + strings.Repeat(" ", width-filled)
}

func maxInt(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

func flattenInt(arrays [][]int) []int {
	result := make([]int, 0)
	for _, arr := range arrays {
		result = append(result, arr...)
	}
	return result
}
