package api

type Object struct {
	ID       string                 `json:"id"`
	MinX     float64                `json:"min_x"`
	MinY     float64                `json:"min_y"`
	MaxX     float64                `json:"max_x"`
	MaxY     float64                `json:"max_y"`
	Metadata map[string]interface{}   `json:"metadata,omitempty"`
}

type AddRequest struct {
	Objects []Object `json:"objects"`
}

type DeleteRequest struct {
	ID string `json:"id"`
}

type SearchRequest struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}

type KNNRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	K int     `json:"k"`
}

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type TreeInfo struct {
	NodeCount  int           `json:"node_count"`
	Height     int           `json:"height"`
	ObjectCount int          `json:"object_count"`
	AllNodesMBR []MBRInfo    `json:"all_nodes_mbr"`
}

type MBRInfo struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}
