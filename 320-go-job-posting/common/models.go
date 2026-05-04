package common

import "time"

type EducationLevel string

const (
	EducationAny       EducationLevel = "不限"
	EducationCollege   EducationLevel = "大专"
	EducationBachelor  EducationLevel = "本科"
	EducationMaster    EducationLevel = "硕士"
	EducationDoctor    EducationLevel = "博士"
)

type ApplicationStatus string

const (
	StatusUnderReview ApplicationStatus = "查看中"
	StatusInterview   ApplicationStatus = "面试邀约"
	StatusNotSuitable ApplicationStatus = "不合适"
)

type Job struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	MinSalary    int            `json:"min_salary"`
	MaxSalary    int            `json:"max_salary"`
	City         string         `json:"city"`
	Education    EducationLevel `json:"education"`
	Experience   string         `json:"experience"`
	Description  string         `json:"description"`
	IsOffline    bool           `json:"is_offline"`
	CreatedAt    time.Time      `json:"created_at"`
}

type Resume struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Summary      string    `json:"summary"`
	CreatedAt    time.Time `json:"created_at"`
}

type Application struct {
	ID              string            `json:"id"`
	JobID           string            `json:"job_id"`
	ResumeID        string            `json:"resume_id"`
	Status          ApplicationStatus `json:"status"`
	InterviewTime   *string           `json:"interview_time,omitempty"`
	InterviewLocation *string         `json:"interview_location,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type ApplicationDetail struct {
	Application
	JobTitle     string `json:"job_title"`
	ResumeName   string `json:"resume_name"`
	ResumePhone  string `json:"resume_phone"`
	ResumeSummary string `json:"resume_summary"`
}
