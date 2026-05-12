package models

import (
	"time"
)

type CourseType string

const (
	CourseTypeLive    CourseType = "live"
	CourseTypeRecord  CourseType = "record"
)

type LiveCourseStatus string

const (
	LiveStatusNotStarted   LiveCourseStatus = "not_started"
	LiveStatusInProgress   LiveCourseStatus = "in_progress"
	LiveStatusEnded        LiveCourseStatus = "ended"
	LiveStatusHasReplay    LiveCourseStatus = "has_replay"
)

type Grade string

const (
	GradePrimary1   Grade = "primary1"
	GradePrimary2   Grade = "primary2"
	GradePrimary3   Grade = "primary3"
	GradePrimary4   Grade = "primary4"
	GradePrimary5   Grade = "primary5"
	GradePrimary6   Grade = "primary6"
	GradeJunior1    Grade = "junior1"
	GradeJunior2    Grade = "junior2"
	GradeJunior3    Grade = "junior3"
	GradeSenior1    Grade = "senior1"
	GradeSenior2    Grade = "senior2"
	GradeSenior3    Grade = "senior3"
	GradeUniversity Grade = "university"
)

var AllGrades = []Grade{
	GradePrimary1, GradePrimary2, GradePrimary3, GradePrimary4, GradePrimary5, GradePrimary6,
	GradeJunior1, GradeJunior2, GradeJunior3,
	GradeSenior1, GradeSenior2, GradeSenior3,
	GradeUniversity,
}

type Subject string

const (
	SubjectChinese  Subject = "chinese"
	SubjectMath     Subject = "math"
	SubjectEnglish  Subject = "english"
	SubjectPhysics  Subject = "physics"
	SubjectChemistry Subject = "chemistry"
	SubjectBiology  Subject = "biology"
	SubjectHistory  Subject = "history"
	SubjectGeography Subject = "geography"
	SubjectPolitics Subject = "politics"
)

var AllSubjects = []Subject{
	SubjectChinese, SubjectMath, SubjectEnglish, SubjectPhysics, SubjectChemistry,
	SubjectBiology, SubjectHistory, SubjectGeography, SubjectPolitics,
}

type Course struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Instructor    string     `json:"instructor"`
	CourseType    CourseType `json:"course_type"`
	Price         int64      `json:"price"`
	Subject       Subject    `json:"subject,omitempty"`
	
	// 直播课字段
	StartTime     *time.Time         `json:"start_time,omitempty"`
	Duration      *int               `json:"duration,omitempty"`
	MaxOnline     *int               `json:"max_online,omitempty"`
	LiveStatus    *LiveCourseStatus  `json:"live_status,omitempty"`
	
	// 录播课字段
	VideoDuration *int64             `json:"video_duration,omitempty"`
	Description   *string            `json:"description,omitempty"`
	
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Student struct {
	ID          int64     `json:"id"`
	Nickname    string    `json:"nickname"`
	Phone       string    `json:"phone"`
	Grade       Grade     `json:"grade"`
	Balance     int64     `json:"balance"`
	CreatedAt   time.Time `json:"created_at"`
}

type RechargeCard struct {
	ID          int64      `json:"id"`
	CardNumber  string     `json:"card_number"`
	Password    string     `json:"password"`
	Amount      int64      `json:"amount"`
	IsUsed      bool       `json:"is_used"`
	UsedBy      *int64     `json:"used_by,omitempty"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
	PurchaseAt  time.Time  `json:"purchase_at"`
	ExpireAt    time.Time  `json:"expire_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Enrollment struct {
	ID          int64     `json:"id"`
	StudentID   int64     `json:"student_id"`
	CourseID    int64     `json:"course_id"`
	PricePaid   int64     `json:"price_paid"`
	EnrolledAt  time.Time `json:"enrolled_at"`
}

type WatchProgress struct {
	ID           int64     `json:"id"`
	StudentID    int64     `json:"student_id"`
	CourseID     int64     `json:"course_id"`
	WatchedSeconds int64   `json:"watched_seconds"`
	IsCompleted  bool      `json:"is_completed"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Review struct {
	ID          int64     `json:"id"`
	StudentID   int64     `json:"student_id"`
	CourseID    int64     `json:"course_id"`
	Rating      int       `json:"rating"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}

type Danmaku struct {
	ID          int64     `json:"id"`
	CourseID    int64     `json:"course_id"`
	StudentID   int64     `json:"student_id"`
	Content     string    `json:"content"`
	SentAt      time.Time `json:"sent_at"`
}

type Vote struct {
	ID          int64      `json:"id"`
	CourseID    int64      `json:"course_id"`
	Title       string     `json:"title"`
	Options     []VoteOption `json:"options"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	IsActive    bool       `json:"is_active"`
}

type VoteOption struct {
	ID      int64  `json:"id"`
	VoteID  int64  `json:"vote_id"`
	Text    string `json:"text"`
	Votes   int    `json:"votes"`
}

type VoteResult struct {
	OptionID    int64   `json:"option_id"`
	OptionText  string  `json:"option_text"`
	Votes       int     `json:"votes"`
	Percentage  float64 `json:"percentage"`
}

type LiveStats struct {
	ID                int64      `json:"id"`
	CourseID          int64      `json:"course_id"`
	PeakOnline        int        `json:"peak_online"`
	AverageOnlineTime float64    `json:"average_online_time"`
	DanmakuCount      int64      `json:"danmaku_count"`
	VoteParticipation float64    `json:"vote_participation"`
	CreatedAt         time.Time  `json:"created_at"`
}

type CourseStats struct {
	ID             int64  `json:"id"`
	Subject        string `json:"subject"`
	SubjectName    string `json:"subject_name"`
	CourseCount    int    `json:"course_count"`
	EnrollmentCount int64  `json:"enrollment_count"`
}

type InstructorStats struct {
	ID             int64   `json:"id"`
	InstructorName string  `json:"instructor_name"`
	CourseCount    int     `json:"course_count"`
	AvgRating      float64 `json:"avg_rating"`
}
