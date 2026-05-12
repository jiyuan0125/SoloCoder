package annealing

import (
	"testing"
)

func TestTwoCities(t *testing.T) {
	cities := []City{
		{Name: "A", X: 0, Y: 0, CoordinateType: CoordinateTypeCartesian},
		{Name: "B", X: 3, Y: 4, CoordinateType: CoordinateTypeCartesian},
	}

	config := SolverConfig{RandomSeed: 42}
	solver, err := NewSolver(cities, config)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	result, err := solver.Solve()
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expectedPathLen := 2
	if len(result.Path) != expectedPathLen {
		t.Errorf("Expected path length %d, got %d", expectedPathLen, len(result.Path))
	}

	if result.Path[0] != 0 || result.Path[1] != 1 {
		t.Errorf("Expected path [0, 1], got %v", result.Path)
	}

	expectedDistance := 10.0
	if result.TotalDistance != expectedDistance {
		t.Errorf("Expected total distance %.2f, got %.2f", expectedDistance, result.TotalDistance)
	}
}

func TestSingleCity(t *testing.T) {
	cities := []City{
		{Name: "A", X: 0, Y: 0, CoordinateType: CoordinateTypeCartesian},
	}

	config := SolverConfig{RandomSeed: 42}
	solver, err := NewSolver(cities, config)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	result, err := solver.Solve()
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	if len(result.Path) != 1 {
		t.Errorf("Expected path length 1, got %d", len(result.Path))
	}

	if result.TotalDistance != 0.0 {
		t.Errorf("Expected total distance 0, got %.2f", result.TotalDistance)
	}
}

func TestSimulatedAnnealingBetterOrEqual(t *testing.T) {
	cities := []City{
		{Name: "A", X: 0, Y: 0, CoordinateType: CoordinateTypeCartesian},
		{Name: "B", X: 10, Y: 0, CoordinateType: CoordinateTypeCartesian},
		{Name: "C", X: 10, Y: 10, CoordinateType: CoordinateTypeCartesian},
		{Name: "D", X: 0, Y: 10, CoordinateType: CoordinateTypeCartesian},
		{Name: "E", X: 5, Y: 5, CoordinateType: CoordinateTypeCartesian},
	}

	config := SolverConfig{
		InitialTemperature:       100.0,
		FinalTemperature:         0.01,
		CoolingRate:              0.95,
		IterationsPerTemperature: 50,
		RandomSeed:               42,
	}

	solver, err := NewSolver(cities, config)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	result, err := solver.Solve()
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	if result.TotalDistance > result.InitialDistance+1e-10 {
		t.Errorf("Simulated annealing result (%.2f) should not be worse than initial solution (%.2f)", 
			result.TotalDistance, result.InitialDistance)
	}

	if result.ImprovementPct < -1e-10 {
		t.Errorf("Improvement percentage should not be negative, got %.2f%%", result.ImprovementPct)
	}
}
