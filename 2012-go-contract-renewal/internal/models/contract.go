package models

import "time"

type Contract struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Party         string    `json:"party"`
	SignDate      time.Time `json:"sign_date"`
	ExpiryDate    time.Time `json:"expiry_date"`
	AutoRenew     bool      `json:"auto_renew"`
	Terminated    bool      `json:"terminated"`
	TerminatedAt  time.Time `json:"terminated_at,omitempty"`
}

type RenewalHistory struct {
	ContractID      string    `json:"contract_id"`
	OriginalID      string    `json:"original_id"`
	PreviousExpiry  time.Time `json:"previous_expiry"`
	NewExpiry       time.Time `json:"new_expiry"`
	IsAutoRenew     bool      `json:"is_auto_renew"`
	RenewedAt       time.Time `json:"renewed_at"`
}

type CleanupRecord struct {
	ContractID  string    `json:"contract_id"`
	Action      string    `json:"action"`
	Details     string    `json:"details"`
	ExecutedAt  time.Time `json:"executed_at"`
}
