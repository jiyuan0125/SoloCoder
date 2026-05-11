package common

import "time"

type CreateTaskRequest struct {
	DeviceID       string `json:"device_id"`
	CargoType      string `json:"cargo_type"`
	StartWarehouse string `json:"start_warehouse"`
	EndWarehouse   string `json:"end_warehouse"`
}

type CreateTaskResponse struct {
	TaskID         string    `json:"task_id"`
	DeviceID       string    `json:"device_id"`
	CargoType      string    `json:"cargo_type"`
	StartWarehouse string    `json:"start_warehouse"`
	EndWarehouse   string    `json:"end_warehouse"`
	StartTime      time.Time `json:"start_time"`
	Status         string    `json:"status"`
	TempMin        float64   `json:"temp_min"`
	TempMax        float64   `json:"temp_max"`
}

type SubmitSensorDataRequest struct {
	DeviceID    string    `json:"device_id"`
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Speed       float64   `json:"speed"`
}

type SubmitSensorDataResponse struct {
	Success bool       `json:"success"`
	Alert   *AlertInfo `json:"alert,omitempty"`
}

type AlertInfo struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Level       string    `json:"level"`
}

type TaskResponse struct {
	TaskID         string     `json:"task_id"`
	DeviceID       string     `json:"device_id"`
	CargoType      string     `json:"cargo_type"`
	StartWarehouse string     `json:"start_warehouse"`
	EndWarehouse   string     `json:"end_warehouse"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	Status         string     `json:"status"`
	TempMin        float64    `json:"temp_min"`
	TempMax        float64    `json:"temp_max"`
}

type ListTasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

type TaskStatsResponse struct {
	TaskID       string  `json:"task_id"`
	TotalSamples int     `json:"total_samples"`
	ValidSamples int     `json:"valid_samples"`
	PassRate     float64 `json:"pass_rate"`
	NeedRecheck  bool    `json:"need_recheck"`
	TempMax      float64 `json:"temp_max"`
	TempMin      float64 `json:"temp_min"`
	TempAvg      float64 `json:"temp_avg"`
	TempStdDev   float64 `json:"temp_stddev"`
}

type SensorDataResponse struct {
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Speed       float64   `json:"speed"`
}

type ListSensorDataResponse struct {
	Data []SensorDataResponse `json:"data"`
}

type ListAlertsResponse struct {
	Alerts []AlertInfo `json:"alerts"`
}

type TaskReportResponse struct {
	TaskID     string              `json:"task_id"`
	Stats      TaskStatsResponse   `json:"stats"`
	Alerts     []AlertInfo         `json:"alerts"`
	ExportedAt time.Time           `json:"exported_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}
