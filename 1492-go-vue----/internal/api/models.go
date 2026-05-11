package api

import "time"

type HouseInfo struct {
	ID          string    `json:"id"`
	Area        float64   `json:"area"`
	Layout      string    `json:"layout"`
	Floor       int       `json:"floor"`
	Orientation string    `json:"orientation"`
	CreatedAt   time.Time `json:"created_at"`
}

type SpaceDesign struct {
	SpaceName string `json:"space_name"`
	Style     string `json:"style"`
	Materials string `json:"materials"`
}

type DesignScheme struct {
	ID          string        `json:"id"`
	HouseID     string        `json:"house_id"`
	Version     int           `json:"version"`
	Spaces      []SpaceDesign `json:"spaces"`
	Description string        `json:"description"`
	Confirmed   bool          `json:"confirmed"`
	CreatedAt   time.Time     `json:"created_at"`
}

type WorkItem struct {
	ID            string  `json:"id"`
	Category      string  `json:"category"`
	SubCategory   string  `json:"sub_category"`
	Unit          string  `json:"unit"`
	Quantity      float64 `json:"quantity"`
	UnitPriceFen  int64   `json:"unit_price_fen"`
	ActualQuantity float64 `json:"actual_quantity"`
	NeedsReconfirm bool   `json:"needs_reconfirm"`
}

type Material struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Spec         string  `json:"spec"`
	Brand        string  `json:"brand"`
	UnitPriceFen int64   `json:"unit_price_fen"`
	Quantity     float64 `json:"quantity"`
}

type Quotation struct {
	ID             string     `json:"id"`
	SchemeID       string     `json:"scheme_id"`
	WorkItems      []WorkItem `json:"work_items"`
	Materials      []Material `json:"materials"`
	WorkTotalFen   int64      `json:"work_total_fen"`
	MaterialTotalFen int64    `json:"material_total_fen"`
	TotalFen       int64      `json:"total_fen"`
	CreatedAt      time.Time  `json:"created_at"`
}

type ChangeOrder struct {
	ID            string    `json:"id"`
	QuotationID   string    `json:"quotation_id"`
	Content       string    `json:"content"`
	Reason        string    `json:"reason"`
	AmountDiffFen int64     `json:"amount_diff_fen"`
	Confirmed     bool      `json:"confirmed"`
	CreatedAt     time.Time `json:"created_at"`
}

type ConstructionPhase struct {
	ID           string     `json:"id"`
	QuotationID  string     `json:"quotation_id"`
	Category     string     `json:"category"`
	OrderIndex   int        `json:"order_index"`
	ExpectedStart time.Time `json:"expected_start"`
	ExpectedEnd   time.Time `json:"expected_end"`
	ActualStart   *time.Time `json:"actual_start"`
	ActualEnd     *time.Time `json:"actual_end"`
	Progress      int        `json:"progress"`
	DelayAlerted  bool       `json:"delay_alerted"`
	LastUpdatedBy string     `json:"last_updated_by"`
}

type Settlement struct {
	ID                string    `json:"id"`
	QuotationID       string    `json:"quotation_id"`
	OriginalTotalFen  int64     `json:"original_total_fen"`
	ChangeTotalFen    int64     `json:"change_total_fen"`
	FinalTotalFen     int64     `json:"final_total_fen"`
	CreatedAt         time.Time `json:"created_at"`
}
