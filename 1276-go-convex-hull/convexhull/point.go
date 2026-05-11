package convexhull

import (
	"fmt"
	"math"
)

const epsilon = 1e-12

type Point struct {
	X, Y float64
}

func (p Point) String() string {
	return fmt.Sprintf("(%.4f, %.4f)", p.X, p.Y)
}

func (p Point) Equal(o Point) bool {
	return math.Abs(p.X-o.X) < epsilon && math.Abs(p.Y-o.Y) < epsilon
}

func (p Point) Less(o Point) bool {
	if math.Abs(p.X-o.X) > epsilon {
		return p.X < o.X
	}
	return p.Y < o.Y
}

func (p Point) Sub(o Point) Point {
	return Point{p.X - o.X, p.Y - o.Y}
}

func (p Point) Cross(o Point) float64 {
	return p.X*o.Y - p.Y*o.X
}

func (p Point) Dot(o Point) float64 {
	return p.X*o.X + p.Y*o.Y
}

func (p Point) Distance(o Point) float64 {
	dx := p.X - o.X
	dy := p.Y - o.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func Sign(x float64) int {
	if x > epsilon {
		return 1
	}
	if x < -epsilon {
		return -1
	}
	return 0
}

func Cross(a, b, c Point) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func IsCollinear(a, b, c Point) bool {
	return Sign(Cross(a, b, c)) == 0
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
