package api

import (
	"property-management/pkg/core/models"
	"time"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateRepairRequest struct {
	OwnerID      string            `json:"owner_id"`
	RepairType   models.RepairType `json:"repair_type"`
	Description  string            `json:"description"`
	ExpectedTime time.Time         `json:"expected_time"`
}

type AssignRepairRequest struct {
	SupervisorID string `json:"supervisor_id"`
	StaffID      string `json:"staff_id"`
}

type CompleteRepairRequest struct {
	StaffID       string   `json:"staff_id"`
	Result        string   `json:"result"`
	ResultImageURLs []string `json:"result_image_urls"`
}

type ConfirmRepairRequest struct {
	OwnerID string `json:"owner_id"`
}

type CreateBillRequest struct {
	UserID   string          `json:"user_id"`
	BillType models.BillType `json:"bill_type"`
	Amount   int64           `json:"amount"`
	DueDate  time.Time       `json:"due_date"`
	Month    int             `json:"month"`
	Year     int             `json:"year"`
}

type GenerateMonthlyBillsRequest struct {
	BillType models.BillType `json:"bill_type"`
	Amount   int64           `json:"amount"`
	DueDate  time.Time       `json:"due_date"`
	Month    int             `json:"month"`
	Year     int             `json:"year"`
}

type PayBillRequest struct {
	UserID        string `json:"user_id"`
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

type CreateAnnouncementRequest struct {
	Title         string                  `json:"title"`
	Content       string                  `json:"content"`
	Scope         models.AnnouncementScope `json:"scope"`
	TargetBuilding string                 `json:"target_building"`
	EffectiveTime time.Time               `json:"effective_time"`
	ExpiryTime    time.Time               `json:"expiry_time"`
	CreatedBy     string                  `json:"created_by"`
}

type UpdateAnnouncementRequest struct {
	Title         string                  `json:"title"`
	Content       string                  `json:"content"`
	Scope         models.AnnouncementScope `json:"scope"`
	TargetBuilding string                 `json:"target_building"`
	EffectiveTime time.Time               `json:"effective_time"`
	ExpiryTime    time.Time               `json:"expiry_time"`
}

type ListRepairsResponse struct {
	Repairs []*models.Repair `json:"repairs"`
}

type ListBillsResponse struct {
	Bills []*models.Bill `json:"bills"`
}

type ListAnnouncementsResponse struct {
	Announcements []*models.Announcement `json:"announcements"`
}

type CalculatePenaltyResponse struct {
	PenaltyAmount int64 `json:"penalty_amount"`
	OverdueDays   int   `json:"overdue_days"`
}
