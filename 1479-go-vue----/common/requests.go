package common

import "time"

type RegisterStudentRequest struct {
	Name        string      `json:"name"`
	IDCard      string      `json:"id_card"`
	Phone       string      `json:"phone"`
	VehicleType VehicleType `json:"vehicle_type"`
}

type UpdateSubjectStatusRequest struct {
	StudentID string        `json:"student_id"`
	Subject   Subject       `json:"subject"`
	Status    SubjectStatus `json:"status"`
}

type AddStudyHoursRequest struct {
	StudentID string `json:"student_id"`
	Subject   Subject `json:"subject"`
	Hours     int    `json:"hours"`
}

type AddCoachRequest struct {
	Name         string      `json:"name"`
	TeachingType VehicleType `json:"teaching_type"`
	Phone        string      `json:"phone"`
}

type SetCoachScheduleRequest struct {
	CoachID   string                     `json:"coach_id"`
	WeekStart time.Time                  `json:"week_start"`
	Slots     map[time.Weekday][]TimeSlot `json:"slots"`
}

type BookPracticeRequest struct {
	StudentID string   `json:"student_id"`
	CoachID   string   `json:"coach_id"`
	TimeSlot  TimeSlot `json:"time_slot"`
	Subject   Subject  `json:"subject"`
}

type CreateExamPlanRequest struct {
	Date       time.Time `json:"date"`
	Subject    Subject   `json:"subject"`
	Venue      string    `json:"venue"`
	TotalQuota int       `json:"total_quota"`
}

type BookExamRequest struct {
	StudentID  string `json:"student_id"`
	ExamPlanID string `json:"exam_plan_id"`
}

type ConfirmWaitingExamRequest struct {
	ExamBookingID string `json:"exam_booking_id"`
}

type CancelExamBookingRequest struct {
	StudentID     string `json:"student_id"`
	ExamBookingID string `json:"exam_booking_id"`
}

type ApproveCancelExamRequest struct {
	ExamBookingID string `json:"exam_booking_id"`
	Approved      bool   `json:"approved"`
}
