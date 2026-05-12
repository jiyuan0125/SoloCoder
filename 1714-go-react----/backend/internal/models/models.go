package models

import (
	"time"

	"gorm.io/gorm"
)

type UnitType string

const (
	UnitTypeHospital       UnitType = "医院"
	UnitTypeClinic         UnitType = "诊所"
	UnitTypeBeautyShop     UnitType = "美容美发店"
	UnitTypeHotel          UnitType = "宾馆酒店"
	UnitTypeSwimmingPool   UnitType = "游泳馆"
	UnitTypeSchool         UnitType = "学校"
	UnitTypeWaterSupply    UnitType = "集中式供水单位"
)

type UnitStatus string

const (
	StatusNormal        UnitStatus = "正常"
	StatusRectifying    UnitStatus = "整改中"
	StatusSuspended     UnitStatus = "停业整顿"
	StatusCancelled     UnitStatus = "注销"
)

type InspectionType string

const (
	InspectionTypeDaily     InspectionType = "日常监督"
	InspectionTypeSpecial   InspectionType = "专项检查"
	InspectionTypeComplaint InspectionType = "投诉举报核查"
)

type InspectionResult string

const (
	ResultPass     InspectionResult = "合格"
	ResultFail     InspectionResult = "不合格"
	ResultNotApply InspectionResult = "不适用"
)

type RectificationStatus string

const (
	RectificationPending  RectificationStatus = "待整改"
	RectificationInReview RectificationStatus = "待复查"
	RectificationPassed   RectificationStatus = "整改合格"
	RectificationFailed   RectificationStatus = "整改失败"
)

type PublicNoticeStatus string

const (
	NoticeStatusActive PublicNoticeStatus = "公示中"
	NoticeStatusHistory PublicNoticeStatus = "历史公示"
)

type SupervisedUnit struct {
	ID                     uint           `gorm:"primaryKey" json:"id"`
	CreditCode             string         `gorm:"uniqueIndex;size:18" json:"credit_code"`
	Name                   string         `gorm:"not null" json:"name"`
	Type                   UnitType       `gorm:"not null" json:"type"`
	Address                string         `json:"address"`
	LegalRepresentative     string         `json:"legal_representative"`
	Phone                  string         `json:"phone"`
	HealthLicenseNumber    string         `gorm:"uniqueIndex;size:50" json:"health_license_number"`
	HealthLicenseValidity  time.Time      `json:"health_license_validity"`
	Status                 UnitStatus     `gorm:"default:正常" json:"status"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}

type InspectionItemTemplate struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	UnitType   UnitType `gorm:"not null;index" json:"unit_type"`
	ItemName   string   `gorm:"not null" json:"item_name"`
	CreatedAt  time.Time `json:"created_at"`
}

type InspectionRecord struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	UnitID            uint           `gorm:"not null;index" json:"unit_id"`
	Unit              SupervisedUnit `gorm:"foreignKey:UnitID" json:"unit"`
	InspectionDate    time.Time      `json:"inspection_date"`
	Inspectors        string         `gorm:"not null" json:"inspectors"`
	InspectionType    InspectionType `gorm:"not null" json:"inspection_type"`
	PassRate          float64        `json:"pass_rate"`
	IsQualified       bool           `json:"is_qualified"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	Items             []InspectionItem `gorm:"foreignKey:InspectionID" json:"items"`
}

type InspectionItem struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	InspectionID uint             `gorm:"not null;index" json:"inspection_id"`
	ItemName     string           `gorm:"not null" json:"item_name"`
	Result       InspectionResult `gorm:"not null" json:"result"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type HealthOpinion struct {
	ID                uint                   `gorm:"primaryKey" json:"id"`
	InspectionID      uint                   `gorm:"not null;uniqueIndex" json:"inspection_id"`
	InspectionRecord  InspectionRecord       `gorm:"foreignKey:InspectionID" json:"inspection_record"`
	IssueDate         time.Time              `json:"issue_date"`
	RectificationDays int                    `gorm:"not null" json:"rectification_days"`
	Deadline          time.Time              `json:"deadline"`
	Status            RectificationStatus    `gorm:"default:待整改" json:"status"`
	Remarks           string                 `json:"remarks"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type Penalty struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	OpinionID       uint             `gorm:"not null;index" json:"opinion_id"`
	Opinion         HealthOpinion    `gorm:"foreignKey:OpinionID" json:"opinion"`
	UnitID          uint             `gorm:"not null;index" json:"unit_id"`
	PenaltyTypes    string           `gorm:"not null" json:"penalty_types"`
	FineAmount      float64          `json:"fine_amount"`
	FineReason      string           `json:"fine_reason"`
	IssueDate       time.Time        `json:"issue_date"`
	Remarks         string           `json:"remarks"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type Complaint struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UnitID          uint           `gorm:"not null;index" json:"unit_id"`
	Unit            SupervisedUnit `gorm:"foreignKey:UnitID" json:"unit"`
	ComplaintDate   time.Time      `json:"complaint_date"`
	Content         string         `gorm:"not null" json:"content"`
	Status          string         `gorm:"default:待处理" json:"status"`
	HandlingResult  string         `json:"handling_result"`
	InspectionID    *uint          `gorm:"index" json:"inspection_id,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type PublicNotice struct {
	ID              uint               `gorm:"primaryKey" json:"id"`
	UnitID          uint               `gorm:"not null;index" json:"unit_id"`
	Unit            SupervisedUnit     `gorm:"foreignKey:UnitID" json:"unit"`
	Title           string             `gorm:"not null" json:"title"`
	Content         string             `gorm:"not null" json:"content"`
	NoticeType      string             `gorm:"not null" json:"notice_type"`
	PublishDate     time.Time          `json:"publish_date"`
	ExpiryDate      time.Time          `json:"expiry_date"`
	Status          PublicNoticeStatus `gorm:"default:公示中" json:"status"`
	RelatedID       *uint              `gorm:"index" json:"related_id,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type AuditLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	Action      string    `gorm:"not null" json:"action"`
	Resource    string    `json:"resource"`
	ResourceID  *uint     `gorm:"index" json:"resource_id,omitempty"`
	Description string    `gorm:"not null" json:"description"`
	IPAddress   string    `json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
}

type PenaltyType string

const (
	PenaltyWarning      PenaltyType = "警告"
	PenaltyFine         PenaltyType = "罚款"
	PenaltySuspend      PenaltyType = "停业整顿"
	PenaltyRevokeLicense PenaltyType = "吊销许可证"
)

type FineRange struct {
	Reason string  `json:"reason"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

var PenaltyFineRanges = []FineRange{
	{Reason: "无证行医", Min: 5000, Max: 30000},
	{Reason: "消毒不合格", Min: 1000, Max: 10000},
	{Reason: "水质不达标", Min: 2000, Max: 20000},
	{Reason: "医疗废物处置不当", Min: 3000, Max: 15000},
	{Reason: "其他违法", Min: 1000, Max: 50000},
}

func GetFineRange(reason string) (float64, float64, bool) {
	for _, r := range PenaltyFineRanges {
		if r.Reason == reason {
			return r.Min, r.Max, true
		}
	}
	return 0, 0, false
}
