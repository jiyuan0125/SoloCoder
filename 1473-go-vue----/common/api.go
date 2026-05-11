package common

type StationDTO struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Type      string  `json:"type"`
}

type LineDTO struct {
	Code          string       `json:"code"`
	Name          string       `json:"name"`
	Direction     string       `json:"direction"`
	FirstTime     string       `json:"first_time"`
	LastTime      string       `json:"last_time"`
	Stations      []StationDTO `json:"stations"`
	IsOperational bool         `json:"is_operational"`
	HasError      bool         `json:"has_error"`
	ErrorReason   string       `json:"error_reason"`
}

type VehicleDTO struct {
	ID             string  `json:"id"`
	LineCode       string  `json:"line_code"`
	LineDirection  string  `json:"line_direction"`
	CurrentStation int     `json:"current_station"`
	Status         string  `json:"status"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
}

type CreateStationRequest struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CreateLineRequest struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Direction   string   `json:"direction"`
	FirstTime   string   `json:"first_time"`
	LastTime    string   `json:"last_time"`
	StationCodes []string `json:"station_codes"`
}

type UpdateLineRequest struct {
	Name         *string  `json:"name,omitempty"`
	FirstTime    *string  `json:"first_time,omitempty"`
	LastTime     *string  `json:"last_time,omitempty"`
	StationCodes []string `json:"station_codes,omitempty"`
}

type CreateVehicleRequest struct {
	ID          string `json:"id"`
	LineCode    string `json:"line_code"`
	Direction   string `json:"direction"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

func SuccessMessage(message string) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
	}
}

func ErrorResponseMessage(message string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error:   message,
	}
}
