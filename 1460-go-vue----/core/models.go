package core

import (
	"time"
)

type Material struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Spec     string  `json:"spec"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type InquiryStatus string

const (
	InquiryStatusOpen    InquiryStatus = "open"
	InquiryStatusExpired InquiryStatus = "expired"
	InquiryStatusClosed  InquiryStatus = "closed"
)

type Inquiry struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	BuyerID         string         `json:"buyer_id"`
	Materials       []Material     `json:"materials"`
	SupplierIDs     []string       `json:"supplier_ids"`
	Deadline        time.Time      `json:"deadline"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Status          InquiryStatus  `json:"status"`
	MaterialsVersion int64         `json:"materials_version"`
}

type QuotationItem struct {
	MaterialID       string  `json:"material_id"`
	UnitPrice        float64 `json:"unit_price"`
	DeliveryDays     int     `json:"delivery_days"`
}

type QuotationStatus string

const (
	QuotationStatusValid   QuotationStatus = "valid"
	QuotationStatusInvalid QuotationStatus = "invalid"
	QuotationStatusExpired QuotationStatus = "expired"
)

type Quotation struct {
	ID               string           `json:"id"`
	InquiryID        string           `json:"inquiry_id"`
	SupplierID       string           `json:"supplier_id"`
	Items            []QuotationItem  `json:"items"`
	TotalPrice       float64          `json:"total_price"`
	Remark           string           `json:"remark"`
	SubmittedAt      time.Time        `json:"submitted_at"`
	Status           QuotationStatus  `json:"status"`
	MaterialsVersion int64            `json:"materials_version"`
}

type TodoReminder struct {
	ID            string    `json:"id"`
	SupplierID    string    `json:"supplier_id"`
	InquiryID     string    `json:"inquiry_id"`
	InquiryTitle  string    `json:"inquiry_title"`
	Deadline      time.Time `json:"deadline"`
	CreatedAt     time.Time `json:"created_at"`
}

type ComparisonItem struct {
	MaterialID        string             `json:"material_id"`
	MaterialName      string             `json:"material_name"`
	Spec              string             `json:"spec"`
	Quantity          float64            `json:"quantity"`
	Unit              string             `json:"unit"`
	SupplierPrices    map[string]float64 `json:"supplier_prices"`
	HasNoPrice        bool               `json:"has_no_price"`
}

type ComparisonResult struct {
	InquiryID      string           `json:"inquiry_id"`
	Items          []ComparisonItem `json:"items"`
	GeneratedAt    time.Time        `json:"generated_at"`
}

type AwardedMaterial struct {
	MaterialID string `json:"material_id"`
	SupplierID string `json:"supplier_id"`
}

type PurchaseOrderItem struct {
	MaterialID   string  `json:"material_id"`
	Name         string  `json:"name"`
	Spec         string  `json:"spec"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	UnitPrice    float64 `json:"unit_price"`
	Amount       float64 `json:"amount"`
	DeliveryDays int     `json:"delivery_days"`
}

type PurchaseOrder struct {
	ID            string             `json:"id"`
	InquiryID     string             `json:"inquiry_id"`
	SupplierID    string             `json:"supplier_id"`
	BuyerID       string             `json:"buyer_id"`
	Items         []PurchaseOrderItem `json:"items"`
	TotalAmount   float64            `json:"total_amount"`
	CreatedAt     time.Time          `json:"created_at"`
	Remark        string             `json:"remark"`
}
