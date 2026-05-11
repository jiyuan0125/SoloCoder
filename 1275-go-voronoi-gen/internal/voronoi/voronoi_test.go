package voronoi

import (
	"math"
	"testing"

	"voronoi/internal/geom"
)

func Test4SquarePoints(t *testing.T) {
	seeds := []geom.Point{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 10},
	}

	diagram := Generate(seeds)
	if len(diagram.Seeds) != 4 {
		t.Errorf("Expected 4 seeds, got %d", len(diagram.Seeds))
	}

	bounds := geom.Rectangle{
		Min: geom.Point{X: -5, Y: -5},
		Max: geom.Point{X: 15, Y: 15},
	}

	cells := BuildCells(diagram, bounds)
	if len(cells) != 4 {
		t.Errorf("Expected 4 cells, got %d", len(cells))
	}

	for i, cell := range cells {
		if len(cell.Points) < 3 {
			t.Errorf("Cell %d has only %d vertices, expected at least 3", i, len(cell.Points))
		}
		t.Logf("Cell %d: seed=(%.2f, %.2f), vertices=%d, area=%.2f",
			i, cell.Seed.X, cell.Seed.Y, len(cell.Points), cell.Area)
	}

	totalArea := 0.0
	for _, cell := range cells {
		totalArea += cell.Area
	}
	expectedTotal := bounds.Width() * bounds.Height()
	if math.Abs(totalArea-expectedTotal) > 10.0 {
		t.Errorf("Total area %.2f != expected %.2f", totalArea, expectedTotal)
	}
	t.Logf("Total area: %.2f (expected %.2f)", totalArea, expectedTotal)
}

func Test2Points(t *testing.T) {
	seeds := []geom.Point{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
	}

	diagram := Generate(seeds)
	if len(diagram.Seeds) != 2 {
		t.Errorf("Expected 2 seeds, got %d", len(diagram.Seeds))
	}

	bounds := geom.Rectangle{
		Min: geom.Point{X: -5, Y: -5},
		Max: geom.Point{X: 15, Y: 5},
	}

	cells := BuildCells(diagram, bounds)
	if len(cells) != 2 {
		t.Errorf("Expected 2 cells, got %d", len(cells))
	}

	for i, cell := range cells {
		if len(cell.Points) < 3 {
			t.Errorf("Cell %d has only %d vertices", i, len(cell.Points))
		}
		t.Logf("Cell %d: seed=(%.2f, %.2f), vertices=%d, area=%.2f",
			i, cell.Seed.X, cell.Seed.Y, len(cell.Points), cell.Area)
	}

	totalArea := 0.0
	for _, cell := range cells {
		totalArea += cell.Area
	}
	t.Logf("Total area: %.2f", totalArea)
}

func Test3ColinearPoints(t *testing.T) {
	seeds := []geom.Point{
		{X: 0, Y: 0},
		{X: 5, Y: 0},
		{X: 10, Y: 0},
	}

	diagram := Generate(seeds)
	if len(diagram.Seeds) != 3 {
		t.Errorf("Expected 3 seeds, got %d", len(diagram.Seeds))
	}

	t.Logf("Generated %d edges", len(diagram.Edges))

	bounds := geom.Rectangle{
		Min: geom.Point{X: -2, Y: -5},
		Max: geom.Point{X: 12, Y: 5},
	}

	cells := BuildCells(diagram, bounds)
	if len(cells) != 3 {
		t.Errorf("Expected 3 cells, got %d", len(cells))
	}

	for i, cell := range cells {
		t.Logf("Cell %d: seed=(%.2f, %.2f), vertices=%d, edges=%d, area=%.2f",
			i, cell.Seed.X, cell.Seed.Y, len(cell.Points), len(cell.Edges), cell.Area)
	}

	totalArea := 0.0
	for _, cell := range cells {
		totalArea += cell.Area
	}
	t.Logf("Total area: %.2f", totalArea)
}

func TestNearestNeighbor(t *testing.T) {
	seeds := []geom.Point{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 10},
	}

	diagram := Generate(seeds)

	testCases := []struct {
		query    geom.Point
		expected int
	}{
		{geom.Point{X: 1, Y: 1}, 0},
		{geom.Point{X: 9, Y: 1}, 1},
		{geom.Point{X: 9, Y: 9}, 2},
		{geom.Point{X: 1, Y: 9}, 3},
	}

	for _, tc := range testCases {
		_, _, idx := FindNearestNeighbor(diagram, tc.query)
		if idx != tc.expected {
			t.Errorf("Query (%.0f,%.0f): expected index %d, got %d",
				tc.query.X, tc.query.Y, tc.expected, idx)
		}
	}
}
