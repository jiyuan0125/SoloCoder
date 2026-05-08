package astar

import (
	"container/heap"
	"math"
)

type Point struct {
	X int
	Y int
}

type Result struct {
	Path         []Point
	ExploreOrder []Point
	TotalCost    float64
	Found        bool
}

const (
	DiagonalCost = math.Sqrt2
	StraightCost = 1.0
)

var (
	fourDirs = [][2]int{
		{0, -1},
		{1, 0},
		{0, 1},
		{-1, 0},
	}

	eightDirs = [][2]int{
		{0, -1},
		{1, 0},
		{0, 1},
		{-1, 0},
		{-1, -1},
		{1, -1},
		{1, 1},
		{-1, 1},
	}

	diagonalChecks = map[[2]int][2][2]int{
		{-1, -1}: {{-1, 0}, {0, -1}},
		{1, -1}:  {{1, 0}, {0, -1}},
		{1, 1}:   {{1, 0}, {0, 1}},
		{-1, 1}:  {{-1, 0}, {0, 1}},
	}
)

func manhattanDistance(x1, y1, x2, y2 int) float64 {
	dx := math.Abs(float64(x1 - x2))
	dy := math.Abs(float64(y1 - y2))
	return dx + dy
}

func chebyshevDistance(x1, y1, x2, y2 int) float64 {
	dx := math.Abs(float64(x1 - x2))
	dy := math.Abs(float64(y1 - y2))
	return math.Max(dx, dy)
}

type nodeState struct {
	x, y int
	g    float64
	h    float64
	f    float64
	prev *nodeState
}

func FindPath(grid Grid, start, goal Point, octile bool) (*Result, error) {
	if grid == nil || grid.Height() == 0 {
		return nil, ErrEmptyGrid
	}

	if !grid.InBounds(start.X, start.Y) || !grid.IsPassable(start.X, start.Y) {
		return nil, ErrInvalidStart
	}

	if !grid.InBounds(goal.X, goal.Y) || !grid.IsPassable(goal.X, goal.Y) {
		return nil, ErrInvalidGoal
	}

	if start.X == goal.X && start.Y == goal.Y {
		return &Result{
			Path:         []Point{start},
			ExploreOrder: []Point{start},
			TotalCost:    0,
			Found:        true,
		}, nil
	}

	var h func(x1, y1, x2, y2 int) float64
	if octile {
		h = chebyshevDistance
	} else {
		h = manhattanDistance
	}

	dirs := fourDirs
	if octile {
		dirs = eightDirs
	}

	type mapKey struct {
		x, y int
	}

	openSet := make(priorityQueue, 0)
	openMap := make(map[mapKey]*node)

	cameFrom := make(map[mapKey]*nodeState)
	gScore := make(map[mapKey]float64)
	exploreOrder := make([]Point, 0)

	startKey := mapKey{start.X, start.Y}
	startH := h(start.X, start.Y, goal.X, goal.Y)
	startNode := &node{
		x: start.X,
		y: start.Y,
		g: 0,
		h: startH,
		f: startH,
	}

	heap.Push(&openSet, startNode)
	openMap[startKey] = startNode
	gScore[startKey] = 0

	for openSet.Len() > 0 {
		current := heap.Pop(&openSet).(*node)
		currentKey := mapKey{current.x, current.y}
		delete(openMap, currentKey)

		exploreOrder = append(exploreOrder, Point{current.x, current.y})

		if current.x == goal.X && current.y == goal.Y {
			path := []Point{}
			cur := &nodeState{x: current.x, y: current.y}
			for {
				path = append([]Point{{cur.x, cur.y}}, path...)
				key := mapKey{cur.x, cur.y}
				if prev, ok := cameFrom[key]; ok {
					cur = prev
				} else {
					break
				}
			}

			totalCost, _ := gScore[currentKey]
			return &Result{
				Path:         path,
				ExploreOrder: exploreOrder,
				TotalCost:    totalCost,
				Found:        true,
			}, nil
		}

		for _, dir := range dirs {
			neighborX := current.x + dir[0]
			neighborY := current.y + dir[1]
			neighborKey := mapKey{neighborX, neighborY}

			if !grid.InBounds(neighborX, neighborY) {
				continue
			}

			if !grid.IsPassable(neighborX, neighborY) {
				continue
			}

			isDiagonal := dir[0] != 0 && dir[1] != 0
			if isDiagonal && octile {
				checks, hasChecks := diagonalChecks[[2]int{dir[0], dir[1]}]
				if hasChecks {
					c1x := current.x + checks[0][0]
					c1y := current.y + checks[0][1]
					c2x := current.x + checks[1][0]
					c2y := current.y + checks[1][1]

					if !grid.IsPassable(c1x, c1y) || !grid.IsPassable(c2x, c2y) {
						continue
					}
				}
			}

			var stepCost float64
			if isDiagonal {
				stepCost = DiagonalCost
			} else {
				stepCost = StraightCost
			}

			currentG, _ := gScore[currentKey]
			tentativeG := currentG + stepCost

			neighborG, exists := gScore[neighborKey]
			if !exists || tentativeG < neighborG {
				neighborH := h(neighborX, neighborY, goal.X, goal.Y)
				neighborF := tentativeG + neighborH

				cameFrom[neighborKey] = &nodeState{
					x: current.x,
					y: current.y,
				}
				gScore[neighborKey] = tentativeG

				if existingNode, ok := openMap[neighborKey]; ok {
					openSet.update(existingNode, tentativeG, neighborH, neighborF)
				} else {
					newNode := &node{
						x: neighborX,
						y: neighborY,
						g: tentativeG,
						h: neighborH,
						f: neighborF,
					}
					heap.Push(&openSet, newNode)
					openMap[neighborKey] = newNode
				}
			}
		}
	}

	return &Result{
		Path:         []Point{},
		ExploreOrder: exploreOrder,
		TotalCost:    0,
		Found:        false,
	}, nil
}
