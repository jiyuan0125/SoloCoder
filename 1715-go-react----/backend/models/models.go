package models

import (
	"time"

	"gorm.io/gorm"
)

type DiseaseClass string

const (
	ClassA DiseaseClass = "甲类"
	ClassB DiseaseClass = "乙类"
	ClassC DiseaseClass = "丙类"
)

type Disease struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"uniqueIndex;not null" json:"name"`
	Class          DiseaseClass   `gorm:"not null" json:"class"`
	MinIncubation  int            `gorm:"not null;default:1" json:"minIncubation"`
	MaxIncubation  int            `gorm:"not null;default:14" json:"maxIncubation"`
	ReportDeadline int            `gorm:"not null;default:24" json:"reportDeadline"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type OutbreakReport struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	PatientName    string         `gorm:"not null" json:"patientName"`
	PatientID      string         `gorm:"not null" json:"patientId"`
	DiseaseName    string         `gorm:"not null" json:"diseaseName"`
	Region         string         `gorm:"not null" json:"region"`
	District       string         `gorm:"not null" json:"district"`
	OnsetTime      time.Time      `gorm:"not null" json:"onsetTime"`
	ReportTime     time.Time      `gorm:"not null" json:"reportTime"`
	IsDelayed      bool           `gorm:"default:false" json:"isDelayed"`
	Status         string         `gorm:"default:pending" json:"status"`
	HasClustering  bool           `gorm:"default:false" json:"hasClustering"`
	InvestigationID *uint         `json:"investigationId"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type Investigation struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ReportID      uint           `gorm:"not null" json:"reportId"`
	PatientName   string         `gorm:"not null" json:"patientName"`
	OnsetTime     time.Time      `json:"onsetTime"`
	VisitTime     time.Time      `json:"visitTime"`
	ConfirmTime   time.Time      `json:"confirmTime"`
	Clinical      string         `json:"clinical"`
	Status        string         `gorm:"default:in_progress" json:"status"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ContactType string

const (
	ContactLive   ContactType = "同住"
	ContactEat    ContactType = "同餐"
	ContactWork   ContactType = "同工作"
	ContactTravel ContactType = "同乘车"
	ContactOther  ContactType = "其他"
)

type ContactStatus string

const (
	StatusNormal    ContactStatus = "正常"
	StatusSymptoms  ContactStatus = "出现症状"
	StatusConfirmed ContactStatus = "确诊"
	StatusExcluded  ContactStatus = "排除"
)

type Contact struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	InvestigationID  uint           `gorm:"not null" json:"investigationId"`
	Name             string         `gorm:"not null" json:"name"`
	Phone            string         `json:"phone"`
	ContactType      ContactType    `gorm:"not null" json:"contactType"`
	FirstContactDate time.Time      `gorm:"not null" json:"firstContactDate"`
	LastContactDate  time.Time      `gorm:"not null" json:"lastContactDate"`
	ObservationStart time.Time      `json:"observationStart"`
	ObservationEnd   time.Time      `json:"observationEnd"`
	Status           ContactStatus  `gorm:"default:正常" json:"status"`
	DailyUpdates     string         `json:"dailyUpdates"`
	IsNewCase        bool           `gorm:"default:false" json:"isNewCase"`
	ParentContactID  *uint          `json:"parentContactId"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type Vaccine struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	Manufacturer  string         `json:"manufacturer"`
	BatchNumber   string         `json:"batchNumber"`
	ExpiryDate    time.Time      `json:"expiryDate"`
	Schedule      string         `json:"schedule"`
	Doses         int            `gorm:"default:1" json:"doses"`
	IntervalDays  string         `json:"intervalDays"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type VaccinationRecord struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	RecipientName  string         `gorm:"not null" json:"recipientName"`
	RecipientID    string         `gorm:"not null" json:"recipientId"`
	VaccineName    string         `gorm:"not null" json:"vaccineName"`
	DoseNumber     int            `gorm:"not null" json:"doseNumber"`
	VaccinationDate time.Time     `gorm:"not null" json:"vaccinationDate"`
	Unit           string         `json:"unit"`
	Doctor         string         `json:"doctor"`
	Contraindication string        `json:"contraindication"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type TodoPriority string

const (
	PriorityHighest TodoPriority = "最高"
	PriorityHigh    TodoPriority = "高"
	PriorityMedium  TodoPriority = "中"
	PriorityLow     TodoPriority = "低"
)

type TodoStatus string

const (
	TodoPending   TodoStatus = "待办"
	TodoProcessing TodoStatus = "处理中"
	TodoDone      TodoStatus = "已完成"
)

type Todo struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Priority    TodoPriority   `gorm:"default:中" json:"priority"`
	Status      TodoStatus     `gorm:"default:待办" json:"status"`
	RelatedType string         `json:"relatedType"`
	RelatedID   uint           `json:"relatedId"`
	Assignee    string         `json:"assignee"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Alert struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Type        string         `gorm:"not null" json:"type"`
	Message     string         `gorm:"not null" json:"message"`
	DiseaseName string         `json:"diseaseName"`
	District    string         `json:"district"`
	Count       int            `json:"count"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type WeeklyReport struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	WeekStart       time.Time      `gorm:"not null" json:"weekStart"`
	WeekEnd         time.Time      `gorm:"not null" json:"weekEnd"`
	NewCases        int            `json:"newCases"`
	DiseaseDist     string         `json:"diseaseDist"`
	RegionDist      string         `json:"regionDist"`
	ContactStats    string         `json:"contactStats"`
	ReportContent   string         `json:"reportContent"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
