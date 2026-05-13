package models

import (
	"time"
)

type ContractStatus string

const (
	StatusDraft        ContractStatus = "draft"
	StatusLegalReview  ContractStatus = "legal_review"
	StatusModifying    ContractStatus = "modifying"
	StatusPendingSign  ContractStatus = "pending_sign"
	StatusSigned       ContractStatus = "signed"
	StatusPerforming   ContractStatus = "performing"
	StatusExpiryRemind ContractStatus = "expiry_remind"
	StatusExpired      ContractStatus = "expired"
	StatusTerminated   ContractStatus = "terminated"
)

type Contract struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Parties         []string       `json:"parties"`
	SignDate        time.Time      `json:"sign_date"`
	EffectiveDate   time.Time      `json:"effective_date"`
	ExpiryDate      time.Time      `json:"expiry_date"`
	Amount          float64        `json:"amount"`
	Status          ContractStatus `json:"status"`
	Content         string         `json:"content"`
	BreachRecords   []BreachRecord `json:"breach_records"`
	Reminded30      bool           `json:"reminded_30"`
	Reminded14      bool           `json:"reminded_14"`
	Reminded7       bool           `json:"reminded_7"`
	Archived        bool           `json:"archived"`
	RelatedContract string         `json:"related_contract,omitempty"`
	ResourceIDs     []string       `json:"resource_ids"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type BreachRecord struct {
	ID          string    `json:"id"`
	ContractID  string    `json:"contract_id"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
}

type OperationLog struct {
	ID         string         `json:"id"`
	ContractID string         `json:"contract_id"`
	Action     string         `json:"action"`
	FromStatus ContractStatus `json:"from_status"`
	ToStatus   ContractStatus `json:"to_status"`
	ResourceID string         `json:"resource_id,omitempty"`
	Details    string         `json:"details"`
	Timestamp  time.Time      `json:"timestamp"`
}

type ResourceType string

const (
	ResourceTypeProject  ResourceType = "project"
	ResourceTypeCustomer ResourceType = "customer"
	ResourceTypeVendor   ResourceType = "vendor"
	ResourceTypeAsset    ResourceType = "asset"
	ResourceTypeOther    ResourceType = "other"
)

type Resource struct {
	ID   string       `json:"id"`
	Name string       `json:"name"`
	Type ResourceType `json:"type"`
}
