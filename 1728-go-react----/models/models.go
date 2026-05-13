package models

import "time"

type Lab struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	BuildingNo   string `json:"building_no"`
	RoomNo       string `json:"room_no"`
	Manager      string `json:"manager"`
	DangerLevel  string `json:"danger_level"`
}

type Chemical struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	CAS           string     `json:"cas"`
	HazardCategory string    `json:"hazard_category"`
	Quantity      float64    `json:"quantity"`
	StorageLabID  string     `json:"storage_lab_id"`
	StorageCabinet string    `json:"storage_cabinet"`
	InboundDate   time.Time  `json:"inbound_date"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"`
	Status        string     `json:"status"`
}

type UsageRecord struct {
	ID         string    `json:"id"`
	ChemicalID string    `json:"chemical_id"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	Quantity   float64   `json:"quantity"`
	UsedAt     time.Time `json:"used_at"`
}

type TrainingParticipant struct {
	UserID      string  `json:"user_id"`
	UserName    string  `json:"user_name"`
	Score       float64 `json:"score"`
	Passed      bool    `json:"passed"`
}

type Training struct {
	ID            string               `json:"id"`
	Topic         string               `json:"topic"`
	Date          time.Time            `json:"date"`
	DurationHours float64              `json:"duration_hours"`
	Instructor    string               `json:"instructor"`
	Participants  []TrainingParticipant `json:"participants"`
}

type SafetyCheckItem struct {
	Item       string     `json:"item"`
	Result     string     `json:"result"`
	Responsible string    `json:"responsible,omitempty"`
	Deadline   *time.Time `json:"deadline,omitempty"`
}

type SafetyCheck struct {
	ID         string            `json:"id"`
	LabID      string            `json:"lab_id"`
	Inspector  string            `json:"inspector"`
	CheckDate  time.Time         `json:"check_date"`
	Items      []SafetyCheckItem `json:"items"`
}

type Todo struct {
	ID          string     `json:"id"`
	RelatedID   string     `json:"related_id"`
	RelatedType string     `json:"related_type"`
	Description string     `json:"description"`
	Responsible string     `json:"responsible"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Status      string     `json:"status"`
	Overdue     bool       `json:"overdue"`
	CreatedAt   time.Time  `json:"created_at"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Operator  string    `json:"operator"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details"`
}

type Department struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	MonthlyBudget float64 `json:"monthly_budget"`
}

type RequestItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type Request struct {
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	DepartmentID     string        `json:"department_id"`
	Applicant        string        `json:"applicant"`
	TotalAmount      float64       `json:"total_amount"`
	Status           string        `json:"status"`
	Items            []RequestItem `json:"items"`
	CreatedAt        time.Time     `json:"created_at"`
	AssignedAt       *time.Time    `json:"assigned_at,omitempty"`
	ProcessedAt      *time.Time    `json:"processed_at,omitempty"`
	CompletedAt      *time.Time    `json:"completed_at,omitempty"`
	ClosedAt         *time.Time    `json:"closed_at,omitempty"`
}

type Alert struct {
	ID           string    `json:"id"`
	DepartmentID string    `json:"department_id"`
	Message      string    `json:"message"`
	Month        string    `json:"month"`
	CreatedAt    time.Time `json:"created_at"`
}
