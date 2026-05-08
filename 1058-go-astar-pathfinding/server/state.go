package main

import (
	"astar-pathfinding/astar"
	"sync"
)

type ServerState struct {
	mu          sync.RWMutex
	grid        astar.Grid
	start       astar.Point
	goal        astar.Point
	octile      bool
	lastResult  *astar.Result
}

func NewServerState() *ServerState {
	return &ServerState{
		grid:   nil,
		start:  astar.Point{},
		goal:   astar.Point{},
		octile: false,
	}
}

func (s *ServerState) SetMap(grid astar.Grid, start, goal astar.Point, octile bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grid = grid
	s.start = start
	s.goal = goal
	s.octile = octile
	s.lastResult = nil
}

func (s *ServerState) GetMap() (astar.Grid, astar.Point, astar.Point, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.grid, s.start, s.goal, s.octile
}

func (s *ServerState) HasMap() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.grid != nil
}

func (s *ServerState) SetResult(result *astar.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastResult = result
}

func (s *ServerState) GetResult() *astar.Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastResult
}
