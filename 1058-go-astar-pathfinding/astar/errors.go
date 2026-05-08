package astar

import "errors"

var (
	ErrInvalidStart  = errors.New("start point is not passable or out of bounds")
	ErrInvalidGoal   = errors.New("goal point is not passable or out of bounds")
	ErrInvalidSize   = errors.New("invalid grid size")
	ErrInvalidDensity = errors.New("obstacle density must be between 0 and 1")
	ErrNoPath        = errors.New("no path found")
	ErrEmptyGrid     = errors.New("grid is empty")
)
