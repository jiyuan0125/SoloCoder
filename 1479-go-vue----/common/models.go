package common

import "time"

type VehicleType string

const (
	VehicleTypeC1 VehicleType = "C1"
	VehicleTypeC2 VehicleType = "C2"
	VehicleTypeA1 VehicleType = "A1"
	VehicleTypeB2 VehicleType = "B2"
)

type Subject int

const (
	Subject1 Subject = 1
	Subject2 Subject = 2
	Subject3 Subject = 3
	Subject4 Subject = 4
)

type SubjectStatus string

const (
	SubjectStatusNotStarted SubjectStatus = "not_started"
	SubjectStatusStudying   SubjectStatus = "studying"
	SubjectStatusPassed     SubjectStatus = "passed"
)

type Student struct {
	ID            string                      `json:"id"`
	Name          string                      `json:"name"`
	IDCard        string                      `json:"id_card"`
	Phone         string                      `json:"phone"`
	VehicleType   VehicleType                 `json:"vehicle_type"`
	SubjectStatus map[Subject]SubjectStatus   `json:"subject_status"`
	StudyHours    map[Subject]int             `json:"study_hours"`
	CoachID       string                      `json:"coach_id,omitempty"`
}

type Coach struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	TeachingType VehicleType `json:"teaching_type"`
	Phone        string      `json:"phone"`
	CurrentStudents int     `json:"current_students"`
	MaxStudents  int         `json:"max_students"`
}

type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type CoachSchedule struct {
	CoachID   string                 `json:"coach_id"`
	WeekStart time.Time              `json:"week_start"`
	Slots     map[time.Weekday][]TimeSlot `json:"slots"`
}

type PracticeBooking struct {
	ID        string    `json:"id"`
	StudentID string    `json:"student_id"`
	CoachID   string    `json:"coach_id"`
	TimeSlot  TimeSlot  `json:"time_slot"`
	Subject   Subject   `json:"subject"`
	CreatedAt time.Time `json:"created_at"`
}

type ExamPlan struct {
	ID         string    `json:"id"`
	Date       time.Time `json:"date"`
	Subject    Subject   `json:"subject"`
	Venue      string    `json:"venue"`
	TotalQuota int       `json:"total_quota"`
	UsedQuota  int       `json:"used_quota"`
}

type ExamBooking struct {
	ID        string    `json:"id"`
	StudentID string    `json:"student_id"`
	ExamPlanID string   `json:"exam_plan_id"`
	Status    string    `json:"status"`
	IsWaiting bool      `json:"is_waiting"`
	WaitOrder int       `json:"wait_order"`
	CreatedAt time.Time `json:"created_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
}

const (
	ExamBookingStatusPending   = "pending"
	ExamBookingStatusConfirmed = "confirmed"
	ExamBookingStatusCancelled = "cancelled"
	ExamBookingStatusExpired   = "expired"
)

var SubjectNames = map[Subject]string{
	Subject1: "科目一",
	Subject2: "科目二",
	Subject3: "科目三",
	Subject4: "科目四",
}

var VehicleTypeNames = map[VehicleType]string{
	VehicleTypeC1: "C1手动挡",
	VehicleTypeC2: "C2自动挡",
	VehicleTypeA1: "A1大巴",
	VehicleTypeB2: "B2中巴",
}
