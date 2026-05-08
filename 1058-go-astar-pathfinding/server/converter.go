package main

import (
	"astar-pathfinding/astar"
	"astar-pathfinding/common"
)

func toAstarPoint(p common.Point) astar.Point {
	return astar.Point{X: p.X, Y: p.Y}
}

func toCommonPoint(p astar.Point) common.Point {
	return common.Point{X: p.X, Y: p.Y}
}

func toCommonPoints(pts []astar.Point) []common.Point {
	result := make([]common.Point, len(pts))
	for i, p := range pts {
		result[i] = toCommonPoint(p)
	}
	return result
}

func toAstarGrid(grid [][]int) astar.Grid {
	return astar.Grid(grid)
}

func toCommonGrid(grid astar.Grid) [][]int {
	return [][]int(grid)
}
