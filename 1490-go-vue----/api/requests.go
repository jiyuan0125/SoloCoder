package api

type CreateRepairOrderRequest struct {
	Address          string           `json:"address"`
	Contact          string           `json:"contact"`
	ContactPhone     string           `json:"contact_phone"`
	BlockageType     BlockageType     `json:"blockage_type"`
	BlockageSeverity BlockageSeverity `json:"blockage_severity"`
	IsRecurring      bool             `json:"is_recurring"`
	AreaCode         string           `json:"area_code"`
}

type SubmitRepairRecordRequest struct {
	OrderID          string           `json:"order_id"`
	MasterID         string           `json:"master_id"`
	UncloggingMethod UncloggingMethod `json:"unclogging_method"`
	DurationMinutes  int              `json:"duration_minutes"`
	ReplacedParts    bool             `json:"replaced_parts"`
	PartName         string           `json:"part_name,omitempty"`
	PartCost         float64          `json:"part_cost,omitempty"`
}

type AcceptOrderRequest struct {
	OrderID  string `json:"order_id"`
	Accepted bool   `json:"accepted"`
	Comment  string `json:"comment,omitempty"`
}

type MasterLoginRequest struct {
	MasterID string `json:"master_id"`
}
