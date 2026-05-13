package models

import "time"

type Product struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	SafetyStock  int     `json:"safety_stock"`
}

type Location struct {
	Code      string `json:"code"`
	Capacity  int    `json:"capacity"`
	Used      int    `json:"used"`
}

type Batch struct {
	ID             string    `json:"id"`
	ProductID      string    `json:"product_id"`
	LocationCode   string    `json:"location_code"`
	ManufactureDate time.Time `json:"manufacture_date"`
	Quantity       int       `json:"quantity"`
	CreatedAt      time.Time `json:"created_at"`
}

type InboundOrder struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	BatchID     string    `json:"batch_id"`
	LocationCode string   `json:"location_code"`
	ManufactureDate time.Time `json:"manufacture_date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
}

type OutboundOrder struct {
	ID         string    `json:"id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	Details    []OutboundDetail `json:"details"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
}

type OutboundDetail struct {
	BatchID    string `json:"batch_id"`
	Quantity   int    `json:"quantity"`
	LocationCode string `json:"location_code"`
}

type StockRecord struct {
	ProductID    string `json:"product_id"`
	LocationCode string `json:"location_code"`
	Quantity     int    `json:"quantity"`
}

type StockCountReport struct {
	ID         string          `json:"id"`
	CreatedAt  time.Time       `json:"created_at"`
	Status     string          `json:"status"`
	Items      []StockCountItem `json:"items"`
}

type StockCountItem struct {
	ProductID       string `json:"product_id"`
	LocationCode    string `json:"location_code"`
	SystemQuantity  int    `json:"system_quantity"`
	ActualQuantity  int    `json:"actual_quantity"`
	Difference      int    `json:"difference"`
	DifferencePct   float64 `json:"difference_pct"`
	Highlight       bool   `json:"highlight"`
}

type Database struct {
	Products      []Product      `json:"products"`
	Locations     []Location     `json:"locations"`
	Batches       []Batch        `json:"batches"`
	InboundOrders []InboundOrder `json:"inbound_orders"`
	OutboundOrders []OutboundOrder `json:"outbound_orders"`
	StockRecords  []StockRecord  `json:"stock_records"`
	Reports       []StockCountReport `json:"reports"`
}
