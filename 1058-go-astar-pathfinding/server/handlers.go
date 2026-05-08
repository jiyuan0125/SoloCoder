package main

import (
	"astar-pathfinding/astar"
	"astar-pathfinding/common"
	"encoding/json"
	"net/http"
)

type Handler struct {
	state *ServerState
}

func NewHandler(state *ServerState) *Handler {
	return &Handler{state: state}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	type errorResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	writeJSON(w, status, errorResp{Success: false, Message: message})
}

func (h *Handler) HandleSetMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.SetMapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	grid := toAstarGrid(req.Map)
	if grid.Height() == 0 || grid.Width() == 0 {
		writeError(w, http.StatusBadRequest, "empty or invalid map")
		return
	}

	start := toAstarPoint(req.Start)
	goal := toAstarPoint(req.Goal)

	if !grid.InBounds(start.X, start.Y) || !grid.IsPassable(start.X, start.Y) {
		writeError(w, http.StatusBadRequest, astar.ErrInvalidStart.Error())
		return
	}

	if !grid.InBounds(goal.X, goal.Y) || !grid.IsPassable(goal.X, goal.Y) {
		writeError(w, http.StatusBadRequest, astar.ErrInvalidGoal.Error())
		return
	}

	h.state.SetMap(grid, start, goal, req.Octile)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "map set successfully",
	})
}

func (h *Handler) HandleRandomMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.RandomMapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	grid, err := astar.NewRandomGrid(req.Width, req.Height, req.ObstacleDensity, req.Seed)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	start := astar.Point{X: 0, Y: 0}
	goal := astar.Point{X: req.Width - 1, Y: req.Height - 1}

	for !grid.IsPassable(start.X, start.Y) {
		start.X++
		if start.X >= req.Width {
			start.X = 0
			start.Y++
		}
		if start.Y >= req.Height {
			writeError(w, http.StatusInternalServerError, "could not find passable start point")
			return
		}
	}

	for !grid.IsPassable(goal.X, goal.Y) {
		goal.X--
		if goal.X < 0 {
			goal.X = req.Width - 1
			goal.Y--
		}
		if goal.Y < 0 {
			writeError(w, http.StatusInternalServerError, "could not find passable goal point")
			return
		}
	}

	h.state.SetMap(grid, start, goal, req.Octile)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "random map generated",
		"map": map[string]interface{}{
			"map":    toCommonGrid(grid),
			"width":  req.Width,
			"height": req.Height,
			"start":  toCommonPoint(start),
			"goal":   toCommonPoint(goal),
			"octile": req.Octile,
		},
	})
}

func (h *Handler) HandleGetMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !h.state.HasMap() {
		writeError(w, http.StatusNotFound, "no map set")
		return
	}

	grid, start, goal, octile := h.state.GetMap()
	resp := common.GetMapResponse{
		Map:    toCommonGrid(grid),
		Width:  grid.Width(),
		Height: grid.Height(),
		Start:  toCommonPoint(start),
		Goal:   toCommonPoint(goal),
		Octile: octile,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandlePathfind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !h.state.HasMap() {
		writeError(w, http.StatusBadRequest, "no map set")
		return
	}

	var req common.RunPathfindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	grid, storedStart, storedGoal, storedOctile := h.state.GetMap()

	start := storedStart
	if req.Start.X != 0 || req.Start.Y != 0 {
		start = toAstarPoint(req.Start)
	}

	goal := storedGoal
	if req.Goal.X != 0 || req.Goal.Y != 0 {
		goal = toAstarPoint(req.Goal)
	}

	octile := storedOctile
	if req.Octile {
		octile = true
	}

	result, err := astar.FindPath(grid, start, goal, octile)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.state.SetResult(result)

	resp := common.PathfindingResponse{
		Success:      result.Found,
		Path:         toCommonPoints(result.Path),
		ExploreOrder: toCommonPoints(result.ExploreOrder),
		TotalCost:    result.TotalCost,
	}

	if !result.Found {
		resp.Message = "no path found"
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleGetResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	result := h.state.GetResult()
	if result == nil {
		writeError(w, http.StatusNotFound, "no result available; run pathfind first")
		return
	}

	resp := common.PathfindingResponse{
		Success:      result.Found,
		Path:         toCommonPoints(result.Path),
		ExploreOrder: toCommonPoints(result.ExploreOrder),
		TotalCost:    result.TotalCost,
	}

	if !result.Found {
		resp.Message = "no path found"
	}

	writeJSON(w, http.StatusOK, resp)
}
