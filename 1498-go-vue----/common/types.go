package common

import "time"

type CourseLevel string

const (
	LevelBeginner CourseLevel = "beginner"
	LevelAdvanced CourseLevel = "advanced"
)

type AttendanceStatus string

const (
	AttendancePresent AttendanceStatus = "present"
	AttendanceLeave   AttendanceStatus = "leave"
	AttendanceAbsent  AttendanceStatus = "absent"
)

type ExamResultStatus string

const (
	ExamResultPassed  ExamResultStatus = "passed"
	ExamResultFailed  ExamResultStatus = "failed"
	ExamResultPending ExamResultStatus = "pending"
)

type InterviewResult string

const (
	InterviewPending  InterviewResult = "pending"
	InterviewAccepted InterviewResult = "accepted"
	InterviewRejected InterviewResult = "rejected"
)

type Course struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Level        CourseLevel `json:"level"`
	Hours        int         `json:"hours"`
	Instructor   string      `json:"instructor"`
	Fee          int         `json:"fee"`
	MaxCapacity  int         `json:"max_capacity"`
}

type ClassSchedule struct {
	ID          string    `json:"id"`
	CourseID    string    `json:"course_id"`
	StartDate   time.Time `json:"start_date"`
	TimeSlot    string    `json:"time_slot"`
	Classroom   string    `json:"classroom"`
	Enrolled    int       `json:"enrolled"`
	TotalClasses int       `json:"total_classes"`
}

type Student struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IDCard      string    `json:"id_card"`
	Phone       string    `json:"phone"`
	Education   string    `json:"education"`
	EnrollDate  time.Time `json:"enroll_date"`
}

type Enrollment struct {
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	ScheduleID   string    `json:"schedule_id"`
	EnrollDate   time.Time `json:"enroll_date"`
	FirstPayment int       `json:"first_payment"`
	SecondPayment int      `json:"second_payment"`
	PaidFirst    bool      `json:"paid_first"`
	PaidSecond   bool      `json:"paid_second"`
}

type AttendanceRecord struct {
	ID         string           `json:"id"`
	EnrollmentID string         `json:"enrollment_id"`
	ClassDate  time.Time        `json:"class_date"`
	Status     AttendanceStatus `json:"status"`
}

type Exam struct {
	ID            string           `json:"id"`
	EnrollmentID  string           `json:"enrollment_id"`
	ExamDate      time.Time        `json:"exam_date"`
	WrittenScore  int              `json:"written_score"`
	PracticalScore int             `json:"practical_score"`
	TotalScore    int              `json:"total_score"`
	Status        ExamResultStatus `json:"status"`
	IsRetake      bool             `json:"is_retake"`
	RetakeFee     int              `json:"retake_fee"`
}

type Certificate struct {
	ID        string    `json:"id"`
	StudentID string    `json:"student_id"`
	CourseID  string    `json:"course_id"`
	ExamID    string    `json:"exam_id"`
	CertNo    string    `json:"cert_no"`
	IssueDate time.Time `json:"issue_date"`
}

type Employer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
}

type JobPosting struct {
	ID           string    `json:"id"`
	EmployerID   string    `json:"employer_id"`
	Position     string    `json:"position"`
	RequiredCert string    `json:"required_cert"`
	MinSalary    int       `json:"min_salary"`
	MaxSalary    int       `json:"max_salary"`
	WorkLocation string    `json:"work_location"`
	PostDate     time.Time `json:"post_date"`
}

type Recommendation struct {
	ID            string          `json:"id"`
	JobPostingID  string          `json:"job_posting_id"`
	StudentID     string          `json:"student_id"`
	Consultant    string          `json:"consultant"`
	InterviewDate time.Time       `json:"interview_date"`
	Result        InterviewResult `json:"result"`
	Notes         string          `json:"notes"`
}

type RefundRequest struct {
	ID             string    `json:"id"`
	EnrollmentID   string    `json:"enrollment_id"`
	RequestDate    time.Time `json:"request_date"`
	RefundAmount   int       `json:"refund_amount"`
	Reason         string    `json:"reason"`
}

type CreateCourseRequest struct {
	Name        string      `json:"name"`
	Level       CourseLevel `json:"level"`
	Hours       int         `json:"hours"`
	Instructor  string      `json:"instructor"`
	Fee         int         `json:"fee"`
	MaxCapacity int         `json:"max_capacity"`
}

type CreateScheduleRequest struct {
	CourseID     string    `json:"course_id"`
	StartDate    string    `json:"start_date"`
	TimeSlot     string    `json:"time_slot"`
	Classroom    string    `json:"classroom"`
	TotalClasses int       `json:"total_classes"`
}

type CreateStudentRequest struct {
	Name      string `json:"name"`
	IDCard    string `json:"id_card"`
	Phone     string `json:"phone"`
	Education string `json:"education"`
}

type EnrollRequest struct {
	StudentID  string `json:"student_id"`
	ScheduleID string `json:"schedule_id"`
}

type RecordAttendanceRequest struct {
	EnrollmentID string           `json:"enrollment_id"`
	ClassDate    string           `json:"class_date"`
	Status       AttendanceStatus `json:"status"`
}

type RecordExamRequest struct {
	EnrollmentID  string `json:"enrollment_id"`
	ExamDate      string `json:"exam_date"`
	WrittenScore  int    `json:"written_score"`
	PracticalScore int   `json:"practical_score"`
	IsRetake      bool   `json:"is_retake"`
}

type CreateEmployerRequest struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
}

type CreateJobPostingRequest struct {
	EmployerID   string `json:"employer_id"`
	Position     string `json:"position"`
	RequiredCert string `json:"required_cert"`
	MinSalary    int    `json:"min_salary"`
	MaxSalary    int    `json:"max_salary"`
	WorkLocation string `json:"work_location"`
}

type ProcessRecommendationRequest struct {
	InterviewDate string          `json:"interview_date"`
	Result        InterviewResult `json:"result"`
	Notes         string          `json:"notes"`
}

type RefundRequestReq struct {
	EnrollmentID string `json:"enrollment_id"`
	Reason       string `json:"reason"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
