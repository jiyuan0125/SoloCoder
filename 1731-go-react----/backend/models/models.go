package models

import "time"

type IndustryCategory string

const (
	IndustryManufacturing IndustryCategory = "制造"
	IndustryService       IndustryCategory = "服务"
	IndustryIT            IndustryCategory = "IT"
	IndustryConstruction  IndustryCategory = "建筑"
	IndustryMedical       IndustryCategory = "医疗"
	IndustryOther         IndustryCategory = "其他"
)

type SkillLevel string

const (
	LevelPrimary    SkillLevel = "初级"
	LevelIntermediate SkillLevel = "中级"
	LevelAdvanced   SkillLevel = "高级"
	LevelTechnician SkillLevel = "技师"
	LevelSeniorTech SkillLevel = "高级技师"
)

type ExamSubject string

const (
	SubjectTheory  ExamSubject = "理论知识"
	SubjectPractical ExamSubject = "实操技能"
	SubjectReview  ExamSubject = "综合评审"
)

type BatchStatus string

const (
	BatchNotStarted BatchStatus = "未开始"
	BatchInProgress BatchStatus = "进行中"
	BatchCompleted  BatchStatus = "已结束"
)

type CertificateStatus string

const (
	CertValid   CertificateStatus = "有效"
	CertCancelled CertificateStatus = "注销"
	CertExpired  CertificateStatus = "过期"
)

type LevelConfig struct {
	Level    SkillLevel    `json:"level"`
	Subjects []ExamSubject `json:"subjects"`
}

type Occupation struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Code       string         `json:"code"`
	Industry   IndustryCategory `json:"industry"`
	Levels     []LevelConfig  `json:"levels"`
}

type ExamCandidate struct {
	ID              string            `json:"id"`
	BatchID         string            `json:"batchId"`
	Name            string            `json:"name"`
	IDCard          string            `json:"idCard"`
	Phone           string            `json:"phone"`
	AppliedLevel    SkillLevel        `json:"appliedLevel"`
	Scores          map[ExamSubject]float64 `json:"scores"`
	HasTakenExam    bool              `json:"hasTakenExam"`
	IsPassed        bool              `json:"isPassed"`
}

type ExamBatch struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	OccupationID string        `json:"occupationId"`
	OccupationName string       `json:"occupationName"`
	Level        SkillLevel    `json:"level"`
	ExamDate     time.Time     `json:"examDate"`
	ExamRoom     string        `json:"examRoom"`
	Status       BatchStatus   `json:"status"`
	Candidates   []ExamCandidate `json:"candidates"`
	CreatedAt    time.Time     `json:"createdAt"`
}

type Certificate struct {
	ID           string            `json:"id"`
	CertificateNo string           `json:"certificateNo"`
	CandidateID  string            `json:"candidateId"`
	Name         string            `json:"name"`
	IDCard       string            `json:"idCard"`
	OccupationID string            `json:"occupationId"`
	OccupationName string           `json:"occupationName"`
	Level        SkillLevel        `json:"level"`
	Status       CertificateStatus `json:"status"`
	IssueDate    time.Time         `json:"issueDate"`
	ExpiryDate   time.Time         `json:"expiryDate"`
}

type PassRateStats struct {
	OccupationID   string  `json:"occupationId"`
	OccupationName string  `json:"occupationName"`
	Level          SkillLevel `json:"level"`
	TotalTaken     int     `json:"totalTaken"`
	TotalPassed    int     `json:"totalPassed"`
	PassRate       float64 `json:"passRate"`
}
