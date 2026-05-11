package api

import (
	"time"
)

type Category struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID string  `json:"parent_id"`
	Price    float64 `json:"price"`
}

type CreateRecordRequest struct {
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	CustomerType string    `json:"customer_type"`
	Settlement   string    `json:"settlement"`
	Items        []RecordItem `json:"items"`
}

type RecordItem struct {
	CategoryID string  `json:"category_id"`
	Weight     float64 `json:"weight"`
}

type Record struct {
	ID           string       `json:"id"`
	DateTime     time.Time    `json:"date_time"`
	CustomerID   string       `json:"customer_id"`
	CustomerName string       `json:"customer_name"`
	CustomerType string       `json:"customer_type"`
	Settlement   string       `json:"settlement"`
	Items        []RecordItemDetail `json:"items"`
	TotalAmount  float64      `json:"total_amount"`
}

type RecordItemDetail struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Weight       float64 `json:"weight"`
	Price        float64 `json:"price"`
	Amount       float64 `json:"amount"`
}

type UpdatePriceRequest struct {
	CategoryID string  `json:"category_id"`
	NewPrice   float64 `json:"new_price"`
}

type ExportRequest struct {
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	CategoryID string `json:"category_id"`
}

type SummaryRequest struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

type SummaryItem struct {
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	TotalWeight   float64 `json:"total_weight"`
	TotalAmount   float64 `json:"total_amount"`
	WeightChange  float64 `json:"weight_change"`
	AmountChange  float64 `json:"amount_change"`
}

type SummaryResponse struct {
	Items []SummaryItem `json:"items"`
}
