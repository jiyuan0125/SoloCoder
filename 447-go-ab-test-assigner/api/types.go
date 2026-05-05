package api

import (
	"time"
)

type ExperimentStatus string

const (
	ExperimentStatusDraft    ExperimentStatus = "draft"
	ExperimentStatusRunning  ExperimentStatus = "running"
	ExperimentStatusPaused   ExperimentStatus = "paused"
	ExperimentStatusArchived ExperimentStatus = "archived"
)

type ExperimentType string

const (
	ExperimentTypeNormal ExperimentType = "normal"
	ExperimentTypeGray   ExperimentType = "gray"
)

type GroupType string

const (
	GroupControl   GroupType = "control"
	GroupTreatment GroupType = "treatment"
)

type Experiment struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	Type               ExperimentType   `json:"type"`
	Status             ExperimentStatus `json:"status"`
	ControlPercentage  int              `json:"control_percentage"`
	TreatmentPercentage int              `json:"treatment_percentage"`
	CreatedAt          time.Time        `json:"created_at"`
	StartedAt          *time.Time       `json:"started_at,omitempty"`
	EndedAt            *time.Time       `json:"ended_at,omitempty"`
	MinDurationDays    int              `json:"min_duration_days"`
}

type CreateExperimentRequest struct {
	Name               string         `json:"name"`
	Type               ExperimentType `json:"type"`
	ControlPercentage  int            `json:"control_percentage"`
	TreatmentPercentage int            `json:"treatment_percentage"`
}

type CreateExperimentResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message,omitempty"`
	Experiment Experiment `json:"experiment,omitempty"`
}

type UpdateTrafficRequest struct {
	ExperimentID       string `json:"experiment_id"`
	ControlPercentage  int    `json:"control_percentage"`
	TreatmentPercentage int    `json:"treatment_percentage"`
}

type UpdateTrafficResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message,omitempty"`
	Experiment Experiment `json:"experiment,omitempty"`
}

type AssignUserRequest struct {
	ExperimentID string `json:"experiment_id"`
	UserID       string `json:"user_id"`
}

type AssignUserResponse struct {
	Success    bool      `json:"success"`
	Message    string    `json:"message,omitempty"`
	Group      GroupType `json:"group,omitempty"`
	ExperimentID string  `json:"experiment_id,omitempty"`
	UserID       string  `json:"user_id,omitempty"`
}

type RecordMetricsRequest struct {
	ExperimentID string    `json:"experiment_id"`
	UserID       string    `json:"user_id"`
	Converted    bool      `json:"converted"`
	StayDuration float64   `json:"stay_duration"`
}

type RecordMetricsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type GroupStats struct {
	Group          GroupType `json:"group"`
	TotalUsers     int       `json:"total_users"`
	ConvertedUsers int       `json:"converted_users"`
	ConversionRate float64   `json:"conversion_rate"`
	AvgStayDuration float64  `json:"avg_stay_duration"`
}

type SignificanceTest struct {
	Metric            string  `json:"metric"`
	IsSignificant     bool    `json:"is_significant"`
	PValue            float64 `json:"p_value"`
	ConfidenceLevel   float64 `json:"confidence_level"`
}

type ExperimentStatsResponse struct {
	Success         bool               `json:"success"`
	Message         string             `json:"message,omitempty"`
	Experiment      Experiment         `json:"experiment,omitempty"`
	ControlStats    GroupStats         `json:"control_stats,omitempty"`
	TreatmentStats  GroupStats         `json:"treatment_stats,omitempty"`
	Significance    []SignificanceTest `json:"significance,omitempty"`
	SampleSufficient bool              `json:"sample_sufficient"`
}

type UserAssignment struct {
	ExperimentID string    `json:"experiment_id"`
	ExperimentName string  `json:"experiment_name"`
	Group        GroupType `json:"group"`
}

type GetUserAssignmentsResponse struct {
	Success     bool             `json:"success"`
	Message     string           `json:"message,omitempty"`
	UserID      string           `json:"user_id,omitempty"`
	Assignments []UserAssignment `json:"assignments,omitempty"`
}

type StartExperimentRequest struct {
	ExperimentID string `json:"experiment_id"`
}

type StartExperimentResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message,omitempty"`
	Experiment Experiment `json:"experiment,omitempty"`
}

type EndExperimentRequest struct {
	ExperimentID string `json:"experiment_id"`
}

type EndExperimentResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message,omitempty"`
	Experiment Experiment `json:"experiment,omitempty"`
}

type ArchiveExperimentRequest struct {
	ExperimentID string `json:"experiment_id"`
}

type ArchiveExperimentResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message,omitempty"`
	Experiment Experiment `json:"experiment,omitempty"`
}

type ListExperimentsRequest struct {
	Status *ExperimentStatus `json:"status,omitempty"`
}

type ListExperimentsResponse struct {
	Success     bool         `json:"success"`
	Message     string       `json:"message,omitempty"`
	Experiments []Experiment `json:"experiments,omitempty"`
}

type GetExperimentRequest struct {
	ExperimentID string `json:"experiment_id"`
}

type GetExperimentResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message,omitempty"`
	Experiment Experiment `json:"experiment,omitempty"`
}
