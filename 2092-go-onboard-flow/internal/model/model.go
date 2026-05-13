package model

import "time"

type StepName string

const (
	StepOfferConfirm     StepName = "offer_confirm"
	StepBackgroundCheck  StepName = "background_check"
	StepContractSign     StepName = "contract_sign"
	StepWorkstationPrep  StepName = "workstation_prep"
	StepAccountSetup     StepName = "account_setup"
	StepMentorAssign     StepName = "mentor_assign"
	StepTrainingComplete StepName = "training_complete"
	StepProbationPeriod  StepName = "probation_period"
	StepProbationPass    StepName = "probation_pass"
)

type StepStatus string

const (
	StatusPending    StepStatus = "pending"
	StatusInProgress StepStatus = "in_progress"
	StatusCompleted  StepStatus = "completed"
	StatusFailed     StepStatus = "failed"
	StatusTerminated StepStatus = "terminated"
)

type Employee struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Department      string     `json:"department"`
	HireDate        time.Time  `json:"hire_date"`
	Status          StepStatus `json:"status"`
	CurrentStep     StepName   `json:"current_step"`
	TotalAmount     float64    `json:"total_amount"`
	TerminationNote string     `json:"termination_note,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Step struct {
	ID                 int64      `json:"id"`
	EmployeeID         int64      `json:"employee_id"`
	StepName           StepName   `json:"step_name"`
	StepOrder          int        `json:"step_order"`
	ResponsibleDept    string     `json:"responsible_dept"`
	PlannedStartDate   time.Time  `json:"planned_start_date"`
	PlannedEndDate     time.Time  `json:"planned_end_date"`
	ActualStartDate    *time.Time `json:"actual_start_date,omitempty"`
	ActualEndDate      *time.Time `json:"actual_end_date,omitempty"`
	Status             StepStatus `json:"status"`
	PlannedAmount      float64    `json:"planned_amount"`
	ActualAmount       float64    `json:"actual_amount"`
	DelayReason        string     `json:"delay_reason,omitempty"`
	Passed             *bool      `json:"passed,omitempty"`
	BackgroundCheckID  *int64     `json:"background_check_id,omitempty"`
	AccountSetupID     *int64     `json:"account_setup_id,omitempty"`
	MentorID           *int64     `json:"mentor_id,omitempty"`
	TrainingID         *int64     `json:"training_id,omitempty"`
	ProbationReviewID  *int64     `json:"probation_review_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type BackgroundCheck struct {
	ID         int64     `json:"id"`
	EmployeeID int64     `json:"employee_id"`
	Passed     bool      `json:"passed"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

type AccountSetup struct {
	ID         int64     `json:"id"`
	EmployeeID int64     `json:"employee_id"`
	EmailDone  bool      `json:"email_done"`
	IMDone     bool      `json:"im_done"`
	VPNDone    bool      `json:"vpn_done"`
	DevEnvDone bool      `json:"dev_env_done"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Mentor struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	EmployeeID   string    `json:"employee_id"`
	Department   string    `json:"department"`
	CreatedAt    time.Time `json:"created_at"`
}

type Exam struct {
	ID              int64     `json:"id"`
	TrainingID      int64     `json:"training_id"`
	ExamType        string    `json:"exam_type"`
	Score           int       `json:"score"`
	MaxScore        int       `json:"max_score"`
	Passed          bool      `json:"passed"`
	AttemptCount    int       `json:"attempt_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type Training struct {
	ID               int64      `json:"id"`
	EmployeeID       int64      `json:"employee_id"`
	CultureExamID    *int64     `json:"culture_exam_id"`
	SkillExamID      *int64     `json:"skill_exam_id"`
	Status           StepStatus `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ProbationReview struct {
	ID               int64     `json:"id"`
	EmployeeID       int64     `json:"employee_id"`
	MentorScore      int       `json:"mentor_score"`
	MentorComment    string    `json:"mentor_comment"`
	ManagerScore     int       `json:"manager_score"`
	ManagerComment   string    `json:"manager_comment"`
	TotalScore       int       `json:"total_score"`
	Passed           bool      `json:"passed"`
	Completed        bool      `json:"completed"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type DelayRecord struct {
	ID           int64     `json:"id"`
	StepID       int64     `json:"step_id"`
	EmployeeID   int64     `json:"employee_id"`
	StepName     StepName  `json:"step_name"`
	Reason       string    `json:"reason"`
	Notified     bool      `json:"notified"`
	CreatedAt    time.Time `json:"created_at"`
}

var StepOrder = map[StepName]int{
	StepOfferConfirm:     1,
	StepBackgroundCheck:  2,
	StepContractSign:     3,
	StepWorkstationPrep:  4,
	StepAccountSetup:     5,
	StepMentorAssign:     6,
	StepTrainingComplete: 7,
	StepProbationPeriod:  8,
	StepProbationPass:    9,
}

var StepDepartment = map[StepName]string{
	StepOfferConfirm:     "HR",
	StepBackgroundCheck:  "HR",
	StepContractSign:     "Legal",
	StepWorkstationPrep:  "Admin",
	StepAccountSetup:     "IT",
	StepMentorAssign:     "Department",
	StepTrainingComplete: "HR",
	StepProbationPeriod:  "HR",
	StepProbationPass:    "Department",
}

var StepsInOrder = []StepName{
	StepOfferConfirm,
	StepBackgroundCheck,
	StepContractSign,
	StepWorkstationPrep,
	StepAccountSetup,
	StepMentorAssign,
	StepTrainingComplete,
	StepProbationPeriod,
	StepProbationPass,
}
