package models

type Department string

const (
	DeptImplant      Department = "种植科"
	DeptOrthodontics Department = "正畸科"
	DeptEndodontics  Department = "牙体牙髓科"
	DeptPeriodontics Department = "牙周科"
	DeptPediatric    Department = "儿童口腔科"
)

var AllDepartments = []Department{
	DeptImplant,
	DeptOrthodontics,
	DeptEndodontics,
	DeptPeriodontics,
	DeptPediatric,
}

type WorkShift struct {
	StartHour   int `json:"start_hour"`
	StartMinute int `json:"start_minute"`
	EndHour     int `json:"end_hour"`
	EndMinute   int `json:"end_minute"`
}

type Doctor struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Department Department  `json:"department"`
	WorkShifts []WorkShift `json:"work_shifts"`
}

type Patient struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Appointment struct {
	ID         string   `json:"id"`
	PatientID  string   `json:"patient_id"`
	DoctorID   string   `json:"doctor_id"`
	Date       string   `json:"date"`
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
	Treatments []string `json:"treatments"`
}

type StepStatus string

const (
	StepStatusNotStarted StepStatus = "not_started"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
)

type Step struct {
	ID           string     `json:"id"`
	PlanID       string     `json:"plan_id"`
	Index        int        `json:"index"`
	ExpectedDate string     `json:"expected_date"`
	ActualDate   *string    `json:"actual_date"`
	DoctorID     string     `json:"doctor_id"`
	Description  string     `json:"description"`
	Fee          int        `json:"fee"`
	Status       StepStatus `json:"status"`
	IsOverdue    bool       `json:"is_overdue"`
}

type PlanStatus string

const (
	PlanStatusNotStarted PlanStatus = "not_started"
	PlanStatusInProgress PlanStatus = "in_progress"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusCanceled   PlanStatus = "canceled"
)

type TreatmentPlan struct {
	ID              string       `json:"id"`
	PatientID       string       `json:"patient_id"`
	DiscountPercent int          `json:"discount_percent"`
	Steps           []*Step      `json:"steps"`
	Status          PlanStatus   `json:"status"`
	IsSurgical      bool         `json:"is_surgical"`
	CreatedAt       string       `json:"created_at"`
}

type FollowUpStatus string

const (
	FollowUpPending   FollowUpStatus = "pending"
	FollowUpCompleted FollowUpStatus = "completed"
)

type FollowUpMethod string

const (
	FollowUpPhone  FollowUpMethod = "电话"
	FollowUpWeChat FollowUpMethod = "微信"
	FollowUpSMS    FollowUpMethod = "短信"
)

type FeedbackType string

const (
	FeedbackSatisfied    FeedbackType = "满意"
	FeedbackNeutral      FeedbackType = "一般"
	FeedbackDissatisfied FeedbackType = "不满意"
)

type FollowUp struct {
	ID          string         `json:"id"`
	PatientID   string         `json:"patient_id"`
	DoctorID    string         `json:"doctor_id"`
	PlanID      string         `json:"plan_id"`
	StepID      string         `json:"step_id"`
	DueDate     string         `json:"due_date"`
	Status      FollowUpStatus `json:"status"`
	Method      FollowUpMethod `json:"method"`
	Feedback    FeedbackType   `json:"feedback"`
	Notes       string         `json:"notes"`
	CompletedAt *string        `json:"completed_at"`
}

type TodoPriority string

const (
	TodoPriorityHigh   TodoPriority = "high"
	TodoPriorityMedium TodoPriority = "medium"
	TodoPriorityLow    TodoPriority = "low"
)

type Todo struct {
	ID         string       `json:"id"`
	DoctorID   string       `json:"doctor_id"`
	Title      string       `json:"title"`
	Priority   TodoPriority `json:"priority"`
	PatientID  string       `json:"patient_id"`
	FollowUpID string       `json:"follow_up_id"`
	CreatedAt  string       `json:"created_at"`
	Done       bool         `json:"done"`
}
