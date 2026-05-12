package models

import "time"

type Title string

const (
	TitleAssistant Title = "助教"
	TitleLecturer  Title = "讲师"
	TitleAssociate Title = "副教授"
	TitleProfessor Title = "教授"
)

var TitleOrder = map[Title]int{
	TitleAssistant: 0,
	TitleLecturer:  1,
	TitleAssociate: 2,
	TitleProfessor: 3,
}

type TrainingType string

const (
	TTTeaching TrainingType = "教学能力"
	TTScience  TrainingType = "科研能力"
	TTManage   TrainingType = "管理能力"
	TTEthics   TrainingType = "师德师风"
)

type TrainingForm string

const (
	TFLecture   TrainingForm = "讲座"
	TFWorkshop  TrainingForm = "工作坊"
	TFOnline    TrainingForm = "在线课程"
	TFPractice  TrainingForm = "实践研修"
)

type AttendanceStatus string

const (
	ASPresent AttendanceStatus = "出勤"
	ASLate    AttendanceStatus = "迟到"
	ASLeave   AttendanceStatus = "请假"
	ASAbsent  AttendanceStatus = "缺席"
)

type Training struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Type     TrainingType `json:"type"`
	Form     TrainingForm `json:"form"`
	Date     string       `json:"date"`
	Hours    int          `json:"hours"`
	Lecturer string       `json:"lecturer"`
	Capacity int          `json:"capacity"`
}

type Registration struct {
	ID           string           `json:"id"`
	TrainingID   string           `json:"training_id"`
	TeacherID    string           `json:"teacher_id"`
	Attendance   []AttendanceItem `json:"attendance,omitempty"`
	Score        *float64         `json:"score,omitempty"`
	StudyReport  *bool            `json:"study_report,omitempty"`
}

type AttendanceItem struct {
	Status AttendanceStatus `json:"status"`
}

type Teacher struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CurrentTitle Title `json:"current_title"`
}

type EvaluationDimension string

const (
	EDAttitude EvaluationDimension = "教学态度"
	EDContent  EvaluationDimension = "教学内容"
	EDMethod   EvaluationDimension = "教学方法"
	EDEffect   EvaluationDimension = "教学效果"
)

type TeachingEvaluation struct {
	ID         string  `json:"id"`
	TeacherID  string  `json:"teacher_id"`
	Semester   string  `json:"semester"`
	Attitude   *Score  `json:"attitude"`
	Content    *Score  `json:"content"`
	Method     *Score  `json:"method"`
	Effect     *Score  `json:"effect"`
	FinalScore float64 `json:"final_score"`
	Qualified  bool    `json:"qualified"`
}

type Score struct {
	Student *float64 `json:"student"`
	Supervisor *float64 `json:"supervisor"`
}

type ReviewStage string

const (
	RSInitial  ReviewStage = "学院初审"
	RSExternal ReviewStage = "校外专家盲审"
	RSFinal    ReviewStage = "校评审委员会终审"
)

type ReviewStatus string

const (
	RSPending ReviewStatus = "待评审"
	RSPassed  ReviewStatus = "通过"
	RSRejected ReviewStatus = "不通过"
)

type TitleApplication struct {
	ID                string        `json:"id"`
	TeacherID         string        `json:"teacher_id"`
	Year              int           `json:"year"`
	ApplyTitle        Title         `json:"apply_title"`
	Materials         string        `json:"materials"`
	AchievementSummary string       `json:"achievement_summary"`
	CurrentStage      ReviewStage   `json:"current_stage"`
	InitialReview     *ReviewResult `json:"initial_review"`
	ExternalReview    *ReviewResult `json:"external_review"`
	FinalReview       *ReviewResult `json:"final_review"`
	FinalStatus       *bool         `json:"final_status"`
}

type ReviewResult struct {
	Status  ReviewStatus `json:"status"`
	Comment string       `json:"comment,omitempty"`
}
