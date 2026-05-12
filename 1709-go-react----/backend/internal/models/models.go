package models

import (
	"time"
)

type BloodType string

const (
	BloodTypeA  BloodType = "A"
	BloodTypeB  BloodType = "B"
	BloodTypeAB BloodType = "AB"
	BloodTypeO  BloodType = "O"
)

type OrganType string

const (
	OrganTypeHeart      OrganType = "heart"
	OrganTypeLiver      OrganType = "liver"
	OrganTypeKidneyLeft OrganType = "kidney_left"
	OrganTypeKidneyRight OrganType = "kidney_right"
	OrganTypeLungLeft   OrganType = "lung_left"
	OrganTypeLungRight  OrganType = "lung_right"
	OrganTypePancreas   OrganType = "pancreas"
	OrganTypeCornea     OrganType = "cornea"
	OrganTypeIntestine  OrganType = "intestine"
)

type OrganStatus string

const (
	OrganStatusPending     OrganStatus = "待匹配"
	OrganStatusMatched     OrganStatus = "已匹配"
	OrganStatusAcquired    OrganStatus = "已获取"
	OrganStatusTransplanted OrganStatus = "已移植"
	OrganStatusDiscarded   OrganStatus = "已超时废弃"
)

type UrgencyLevel string

const (
	UrgencyNormal    UrgencyLevel = "普通"
	UrgencyModerate  UrgencyLevel = "较急"
	UrgencyEmergency UrgencyLevel = "紧急"
)

type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "待处理"
	TodoStatusInProgress TodoStatus = "处理中"
	TodoStatusCompleted  TodoStatus = "已完成"
	TodoStatusOverdue    TodoStatus = "已逾期"
)

type TodoType string

const (
	TodoTypeOrganAssessment TodoType = "器官评估"
	TodoTypeWaitingMatching TodoType = "等待匹配"
	TodoTypeNotifyRecipient TodoType = "通知受体"
	TodoTypeScheduleSurgery TodoType = "安排手术"
	TodoTypePostOpFollowUp  TodoType = "术后随访"
)

type SurgeryResult string

const (
	SurgerySuccess       SurgeryResult = "成功"
	SurgeryFailure       SurgeryResult = "失败"
	SurgeryComplication  SurgeryResult = "并发症"
)

type PostOpStatus string

const (
	PostOpInSurgery   PostOpStatus = "手术中"
	PostOpObservation PostOpStatus = "术后观察"
	PostOpDischarged  PostOpStatus = "已出院"
)

type Gender string

const (
	GenderMale   Gender = "男"
	GenderFemale Gender = "女"
)

type OrganAssessment struct {
	FunctionScore int  `json:"function_score"`
	HasVesselAbnormality bool `json:"has_vessel_abnormality"`
	HasOtherLesions bool `json:"has_other_lesions"`
	AssessmentTime time.Time `json:"assessment_time"`
}

type Organ struct {
	ID             string         `json:"id"`
	OrganType      OrganType      `json:"organ_type"`
	DonorID        string         `json:"donor_id"`
	Status         OrganStatus    `json:"status"`
	Assessment     *OrganAssessment `json:"assessment,omitempty"`
	ColdIschemiaDeadline time.Time `json:"cold_ischemia_deadline,omitempty"`
	AcquiredTime   time.Time      `json:"acquired_time,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Donor struct {
	ID             string    `json:"id"`
	DonorNo        string    `json:"donor_no"`
	Name           string    `json:"name"`
	Gender         Gender    `json:"gender"`
	Age            int       `json:"age"`
	BloodType      BloodType `json:"blood_type"`
	Height         float64   `json:"height"`
	Weight         float64   `json:"weight"`
	DeathDate      time.Time `json:"death_date"`
	CauseOfDeath   string    `json:"cause_of_death"`
	Organs         []*Organ  `json:"organs,omitempty"`
	LocationID     string    `json:"location_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Recipient struct {
	ID              string       `json:"id"`
	RecipientNo     string       `json:"recipient_no"`
	Name            string       `json:"name"`
	BloodType       BloodType    `json:"blood_type"`
	OrganNeeded     OrganType    `json:"organ_needed"`
	RegistrationDate time.Time   `json:"registration_date"`
	UrgencyLevel    UrgencyLevel `json:"urgency_level"`
	HLA             string       `json:"hla"`
	PRA             int          `json:"pra"`
	Age             int          `json:"age"`
	IsMatching      bool         `json:"is_matching"`
	MatchedOrganID  *string      `json:"matched_organ_id,omitempty"`
	LocationID      string       `json:"location_id"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type TransplantRecord struct {
	ID              string         `json:"id"`
	OrganID         string         `json:"organ_id"`
	RecipientID     string         `json:"recipient_id"`
	SurgeryDate     time.Time      `json:"surgery_date"`
	Surgeon         string         `json:"surgeon"`
	Result          SurgeryResult  `json:"result"`
	PostOpStatus    PostOpStatus   `json:"post_op_status"`
	FollowUps       []*FollowUp    `json:"follow_ups,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type FollowUp struct {
	ID               string    `json:"id"`
	TransplantID     string    `json:"transplant_id"`
	FollowUpDate     time.Time `json:"follow_up_date"`
	FollowUpType     string    `json:"follow_up_type"`
	OrganFunction    string    `json:"organ_function"`
	RecoveryStatus   string    `json:"recovery_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type Todo struct {
	ID            string     `json:"id"`
	Type          TodoType   `json:"type"`
	RelatedID     string     `json:"related_id"`
	RelatedType   string     `json:"related_type"`
	Assignee      string     `json:"assignee"`
	DueDate       time.Time  `json:"due_date"`
	Status        TodoStatus `json:"status"`
	Description   string     `json:"description"`
	LocationID    string     `json:"location_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Location struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Personnel struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	LocationID string    `json:"location_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type StatisticsReport struct {
	ID               string            `json:"id"`
	ReportDate       time.Time         `json:"report_date"`
	NewDonors        int               `json:"new_donors"`
	SuccessfulMatches int              `json:"successful_matches"`
	AverageWaitTime  float64           `json:"average_wait_time_days"`
	OrganSupplyDemand map[OrganType]float64 `json:"organ_supply_demand_ratio"`
	CreatedAt        time.Time         `json:"created_at"`
}

type MatchNotification struct {
	ID            string    `json:"id"`
	OrganID       string    `json:"organ_id"`
	RecipientID   string    `json:"recipient_id"`
	NotifiedAt    time.Time `json:"notified_at"`
	Confirmed     bool      `json:"confirmed"`
	ExpiresAt     time.Time `json:"expires_at"`
	Priority      int       `json:"priority"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Size       int         `json:"size"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}
