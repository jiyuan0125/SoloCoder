package common

import "time"

type EnergyType string

const (
	EnergyTypeElectricity EnergyType = "electricity"
	EnergyTypeWater       EnergyType = "water"
	EnergyTypeGas         EnergyType = "gas"
	EnergyTypeSteam       EnergyType = "steam"
)

type Point struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Area       string     `json:"area"`
	EnergyType EnergyType `json:"energy_type"`
	Unit       string     `json:"unit"`
	AreaSize   float64    `json:"area_size"`
}

type EnergyData struct {
	ID        string    `json:"id"`
	PointID   string    `json:"point_id"`
	Timestamp time.Time `json:"timestamp"`
	Reading   float64   `json:"reading"`
	Increment float64   `json:"increment"`
}

type SuggestionType string

const (
	SuggestionTypeUsageGrowth SuggestionType = "usage_growth"
	SuggestionTypeHighConsumption SuggestionType = "high_consumption"
)

type Suggestion struct {
	ID         string         `json:"id"`
	PointID    string         `json:"point_id"`
	Area       string         `json:"area"`
	Type       SuggestionType `json:"type"`
	Message    string         `json:"message"`
	Month      string         `json:"month"`
	CreatedAt  time.Time      `json:"created_at"`
}

type TimeGranularity string

const (
	GranularityHour  TimeGranularity = "hour"
	GranularityDay   TimeGranularity = "day"
	GranularityMonth TimeGranularity = "month"
)

type EnergyConsumptionItem struct {
	Time      string  `json:"time"`
	Value     float64 `json:"value"`
	HasData   bool    `json:"has_data"`
}

type AreaConsumptionItem struct {
	Area      string  `json:"area"`
	Value     float64 `json:"value"`
	Ratio     float64 `json:"ratio"`
}

type OverviewMetrics struct {
	ByEnergyType     map[EnergyType]float64 `json:"by_energy_type"`
	YoYChange        map[EnergyType]float64 `json:"yoy_change"`
	YoYIsNew         map[EnergyType]bool    `json:"yoy_is_new"`
	MoMChange        map[EnergyType]float64 `json:"mom_change"`
	MoMIsNew         map[EnergyType]bool    `json:"mom_is_new"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
