package common

import "time"

type CreateOrderReq struct {
	UserID       string `json:"user_id"`
	Address      string `json:"address"`
	LockType     string `json:"lock_type"`
	Urgency      string `json:"urgency"`
	TimeSlot     string `json:"time_slot"`
	TimeSlotDate string `json:"time_slot_date"`
}

type CreateOrderResp struct {
	Success     bool               `json:"success"`
	OrderID     string             `json:"order_id,omitempty"`
	Order       *OrderView         `json:"order,omitempty"`
	Alternative *AlternativeView   `json:"alternative,omitempty"`
	Message     string             `json:"message,omitempty"`
}

type OrderView struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Address        string     `json:"address"`
	LockType       string     `json:"lock_type"`
	Urgency        string     `json:"urgency"`
	TimeSlot       string     `json:"time_slot"`
	TimeSlotDate   string     `json:"time_slot_date"`
	MasterID       string     `json:"master_id"`
	Status         string     `json:"status"`
	Detail         DetailView `json:"detail"`
	BaseCost       float64    `json:"base_cost"`
	UrgentFee      float64    `json:"urgent_fee"`
	TotalCost      float64    `json:"total_cost"`
	CreatedAt      string     `json:"created_at"`
	Rating         int        `json:"rating,omitempty"`
	Comment        string     `json:"comment,omitempty"`
	HasComplaint   bool       `json:"has_complaint,omitempty"`
}

type DetailView struct {
	NeedReplaceLock bool    `json:"need_replace_lock"`
	LockBrand       string  `json:"lock_brand,omitempty"`
	LockModel       string  `json:"lock_model,omitempty"`
	LockLevel       string  `json:"lock_level,omitempty"`
	PartsCost       float64 `json:"parts_cost,omitempty"`
	LaborCost       float64 `json:"labor_cost,omitempty"`
}

type AlternativeView struct {
	OtherSlots   []string        `json:"other_slots,omitempty"`
	OtherMasters []MasterSummary `json:"other_masters,omitempty"`
}

type MasterSummary struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	ServiceArea  string   `json:"service_area"`
	Status       string   `json:"status"`
	AvailableSlots []string `json:"available_slots,omitempty"`
}

type MasterView struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Phone          string   `json:"phone"`
	ServiceArea    string   `json:"service_area"`
	SupportedLocks []string `json:"supported_locks"`
	Status         string   `json:"status"`
	Location       string   `json:"location"`
}

type AcceptOrderReq struct {
	OrderID string `json:"order_id"`
}

type AcceptOrderResp struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StartServiceReq struct {
	OrderID string `json:"order_id"`
}

type StartServiceResp struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type CompleteServiceReq struct {
	OrderID         string  `json:"order_id"`
	NeedReplaceLock bool    `json:"need_replace_lock"`
	LockBrand       string  `json:"lock_brand,omitempty"`
	LockModel       string  `json:"lock_model,omitempty"`
	LockLevel       string  `json:"lock_level,omitempty"`
	PartsCost       float64 `json:"parts_cost,omitempty"`
}

type CompleteServiceResp struct {
	Success bool       `json:"success"`
	Order   *OrderView `json:"order,omitempty"`
	Message string     `json:"message,omitempty"`
}

type PayOrderReq struct {
	OrderID string `json:"order_id"`
}

type PayOrderResp struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RateOrderReq struct {
	OrderID string `json:"order_id"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment,omitempty"`
}

type RateOrderResp struct {
	Success      bool   `json:"success"`
	HasComplaint bool   `json:"has_complaint,omitempty"`
	Message      string `json:"message,omitempty"`
}

type GetOrderResp struct {
	Success bool       `json:"success"`
	Order   *OrderView `json:"order,omitempty"`
	Message string     `json:"message,omitempty"`
}

type ListOrdersResp struct {
	Success bool         `json:"success"`
	Orders  []*OrderView `json:"orders,omitempty"`
	Message string       `json:"message,omitempty"`
}

type ListMastersResp struct {
	Success bool         `json:"success"`
	Masters []*MasterView `json:"masters,omitempty"`
	Message string       `json:"message,omitempty"`
}

type AddMasterReq struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Phone          string   `json:"phone"`
	ServiceArea    string   `json:"service_area"`
	SupportedLocks []string `json:"supported_locks"`
	Location       string   `json:"location"`
}

type AddMasterResp struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
