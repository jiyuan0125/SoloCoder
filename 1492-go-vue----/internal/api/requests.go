package api

import "time"

type CreateHouseRequest struct {
	Area        float64 `json:"area"`
	Layout      string  `json:"layout"`
	Floor       int     `json:"floor"`
	Orientation string  `json:"orientation"`
}

type CreateDesignRequest struct {
	HouseID     string        `json:"house_id"`
	Version     int           `json:"version"`
	Spaces      []SpaceDesign `json:"spaces"`
	Description string        `json:"description"`
}

type ConfirmDesignRequest struct {
	SchemeID string `json:"scheme_id"`
}

type WorkItemInput struct {
	Category     string  `json:"category"`
	SubCategory  string  `json:"sub_category"`
	Unit         string  `json:"unit"`
	Quantity     float64 `json:"quantity"`
	UnitPriceFen int64   `json:"unit_price_fen"`
}

type MaterialInput struct {
	Name         string  `json:"name"`
	Spec         string  `json:"spec"`
	Brand        string  `json:"brand"`
	UnitPriceFen int64   `json:"unit_price_fen"`
	Quantity     float64 `json:"quantity"`
}

type CreateQuotationRequest struct {
	SchemeID  string           `json:"scheme_id"`
	WorkItems []WorkItemInput  `json:"work_items"`
	Materials []MaterialInput  `json:"materials"`
}

type CreateChangeOrderRequest struct {
	QuotationID   string `json:"quotation_id"`
	Content       string `json:"content"`
	Reason        string `json:"reason"`
	AmountDiffFen int64  `json:"amount_diff_fen"`
}

type ConfirmChangeOrderRequest struct {
	ChangeOrderID string `json:"change_order_id"`
}

type UpdateActualQuantityRequest struct {
	WorkItemID     string  `json:"work_item_id"`
	ActualQuantity float64 `json:"actual_quantity"`
}

type CreatePhaseRequest struct {
	QuotationID   string    `json:"quotation_id"`
	Category      string    `json:"category"`
	OrderIndex    int       `json:"order_index"`
	ExpectedStart time.Time `json:"expected_start"`
	ExpectedEnd   time.Time `json:"expected_end"`
}

type UpdateProgressRequest struct {
	PhaseID     string `json:"phase_id"`
	Progress    int    `json:"progress"`
	UpdatedBy   string `json:"updated_by"`
	ActualStart *time.Time `json:"actual_start,omitempty"`
	ActualEnd   *time.Time `json:"actual_end,omitempty"`
}
