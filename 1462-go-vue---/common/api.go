package common

import "time"

type CreatePointRequest struct {
	Name       string     `json:"name"`
	Area       string     `json:"area"`
	EnergyType EnergyType `json:"energy_type"`
	Unit       string     `json:"unit"`
	AreaSize   float64    `json:"area_size"`
}

type ReportDataRequest struct {
	PointID   string    `json:"point_id"`
	Timestamp time.Time `json:"timestamp"`
	Reading   float64   `json:"reading"`
}

type TimeRangeQuery struct {
	Start       time.Time       `json:"start"`
	End         time.Time       `json:"end"`
	Granularity TimeGranularity `json:"granularity"`
}

type PointTimeQuery struct {
	PointID     string          `json:"point_id"`
	Start       time.Time       `json:"start"`
	End         time.Time       `json:"end"`
	Granularity TimeGranularity `json:"granularity"`
}

type AreaTimeQuery struct {
	Area        string          `json:"area"`
	Start       time.Time       `json:"start"`
	End         time.Time       `json:"end"`
	Granularity TimeGranularity `json:"granularity"`
}
