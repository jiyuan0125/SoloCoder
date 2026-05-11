package common

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type StudentListResponse struct {
	Students []Student `json:"students"`
}

type CoachListResponse struct {
	Coaches []Coach `json:"coaches"`
}

type ExamPlanListResponse struct {
	ExamPlans []ExamPlan `json:"exam_plans"`
}

type ExamBookingListResponse struct {
	Bookings []ExamBooking `json:"bookings"`
}

type PracticeBookingListResponse struct {
	Bookings []PracticeBooking `json:"bookings"`
}

func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(message string) Response {
	return Response{
		Success: false,
		Message: message,
	}
}
