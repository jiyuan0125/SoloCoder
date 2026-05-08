package astar

import (
	"math/rand"
	"time"
)

type Grid [][]int

const (
	Passable = 0
	Obstacle = 1
)

func NewGrid(width, height int) Grid {
	g := make(Grid, height)
	for i := range g {
		g[i] = make([]int, width)
	}
	return g
}

func NewRandomGrid(width, height int, density float64, seed int64) (Grid, error) {
	if width <= 0 || height <= 0 {
		return nil, ErrInvalidSize
	}
	if density < 0 || density > 1 {
		return nil, ErrInvalidDensity
	}

	g := NewGrid(width, height)
	r := rand.New(rand.NewSource(seed))

	if seed == 0 {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	total := width * height
	numObstacles := int(float64(total) * density)

	positions := make([]struct{ x, y int }, 0, total)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			positions = append(positions, struct{ x, y int }{x, y})
		}
	}

	for i := range positions {
		j := r.Intn(len(positions))
		positions[i], positions[j] = positions[j], positions[i]
	}

	for i := 0; i < numObstacles && i < len(positions); i++ {
		g[positions[i].y][positions[i].x] = Obstacle
	}

	return g, nil
}

func (g Grid) InBounds(x, y int) bool {
	if y < 0 || y >= len(g) {
		return false
	}
	if x < 0 || x >= len(g[0]) {
		return false
	}
	return true
}

func (g Grid) IsPassable(x, y int) bool {
	if !g.InBounds(x, y) {
		return false
	}
	return g[y][x] == Passable
}

func (g Grid) Width() int {
	if len(g) == 0 {
		return 0
	}
	return len(g[0])
}

func (g Grid) Height() int {
	return len(g)
}
