package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadMapFile(filepath string) ([][]int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open map file: %w", err)
	}
	defer file.Close()

	var grid [][]int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var row []int
		for _, ch := range line {
			if ch == '0' {
				row = append(row, 0)
			} else if ch == '1' {
				row = append(row, 1)
			} else if ch == ' ' || ch == '\t' {
				continue
			} else {
				return nil, fmt.Errorf("invalid character in map file: %c", ch)
			}
		}

		if len(row) > 0 {
			grid = append(grid, row)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading map file: %w", err)
	}

	if len(grid) == 0 {
		return nil, fmt.Errorf("map file is empty")
	}

	width := len(grid[0])
	for i, row := range grid {
		if len(row) != width {
			return nil, fmt.Errorf("inconsistent row width at line %d: expected %d, got %d", i+1, width, len(row))
		}
	}

	return grid, nil
}
