package common

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(err string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error:   err,
	}
}

type VehicleInfoResponse struct {
	Vehicle
	NextInspectionDate string `json:"next_inspection_date"`
	IsDueForInspection bool   `json:"is_due_for_inspection"`
	DaysUntilDue       int    `json:"days_until_due"`
}
