package vectorclock

import "time"

type VectorClock map[string]uint64

type Event struct {
	EventID      uint64      `json:"event_id"`
	NodeID       string      `json:"node_id"`
	Clock        VectorClock `json:"clock"`
	Content      string      `json:"content"`
	CreationTime time.Time   `json:"creation_time"`
}

type CausalityType int

const (
	CausalitySame CausalityType = iota
	CausalityBefore
	CausalityAfter
	CausalityConcurrent
)

const DefaultPruneThreshold = 100
const HistoricalKey = "__historical_max__"
