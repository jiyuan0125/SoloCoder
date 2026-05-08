package common

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type SetMapRequest struct {
	Map     [][]int `json:"map"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Start   Point   `json:"start"`
	Goal    Point   `json:"goal"`
	Octile  bool    `json:"octile"`
}

type RandomMapRequest struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	ObstacleDensity float64 `json:"obstacle_density"`
	Octile      bool    `json:"octile"`
	Seed        int64   `json:"seed,omitempty"`
}

type RunPathfindingRequest struct {
	Start  Point `json:"start"`
	Goal   Point `json:"goal"`
	Octile bool  `json:"octile"`
}

type PathfindingResponse struct {
	Success      bool    `json:"success"`
	Message      string  `json:"message,omitempty"`
	Path         []Point `json:"path"`
	ExploreOrder []Point `json:"explore_order"`
	TotalCost    float64 `json:"total_cost"`
}

type GetMapResponse struct {
	Map    [][]int `json:"map"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Start  Point   `json:"start,omitempty"`
	Goal   Point   `json:"goal,omitempty"`
	Octile bool    `json:"octile"`
}
