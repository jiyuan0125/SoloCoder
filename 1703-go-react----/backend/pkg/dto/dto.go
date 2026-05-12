package dto

import (
	"medical-exam-system/internal/model"
	"time"
)

type CreatePackageRequest struct {
	Name        string   `json:"name" binding:"required"`
	Price       int      `json:"price" binding:"required,min=0"`
	Description string   `json:"description"`
	ItemIDs     []string `json:"item_ids" binding:"required"`
}

type UpdatePackageRequest struct {
	Name        string   `json:"name"`
	Price       int      `json:"price"`
	Description string   `json:"description"`
	ItemIDs     []string `json:"item_ids"`
}

type PackageResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Price       int                  `json:"price"`
	Description string               `json:"description"`
	Items       []ExamItemResponse   `json:"items"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type ExamItemResponse struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Department model.Department  `json:"department"`
	RefRange   string            `json:"ref_range"`
	Unit       string            `json:"unit"`
	Price      int               `json:"price"`
}

type PriceCalculationRequest struct {
	PackageID  string   `json:"package_id" binding:"required"`
	AddItemIDs []string `json:"add_item_ids"`
}

type PriceCalculationResponse struct {
	PackagePrice      int `json:"package_price"`
	AddItemsPrice     int `json:"add_items_price"`
	AddItemsDiscount  int `json:"add_items_discount"`
	AddItemsFinalPrice int `json:"add_items_final_price"`
	TotalPrice        int `json:"total_price"`
}

type CreateAppointmentRequest struct {
	CustomerName string          `json:"customer_name" binding:"required"`
	IDCard       string          `json:"id_card" binding:"required"`
	Phone        string          `json:"phone" binding:"required"`
	Gender       string          `json:"gender"`
	Age          int             `json:"age"`
	PackageID    string          `json:"package_id" binding:"required"`
	AddItemIDs   []string        `json:"add_item_ids"`
	ExamDate     string          `json:"exam_date" binding:"required"`
	TimeSlot     model.TimeSlot  `json:"time_slot" binding:"required"`
}

type AppointmentResponse struct {
	ID           string               `json:"id"`
	ExamNumber   string               `json:"exam_number"`
	CustomerName string               `json:"customer_name"`
	IDCard       string               `json:"id_card"`
	Phone        string               `json:"phone"`
	Gender       string               `json:"gender"`
	Age          int                  `json:"age"`
	PackageID    string               `json:"package_id"`
	PackageName  string               `json:"package_name"`
	AddItemIDs   []string             `json:"add_item_ids"`
	TotalPrice   int                  `json:"total_price"`
	ExamDate     string               `json:"exam_date"`
	TimeSlot     model.TimeSlot       `json:"time_slot"`
	Status       model.AppointmentStatus `json:"status"`
	CreatedAt    time.Time            `json:"created_at"`
}

type DayAvailability struct {
	Date     string `json:"date"`
	Morning  Availability `json:"morning"`
	Afternoon Availability `json:"afternoon"`
}

type Availability struct {
	Total     int `json:"total"`
	Used      int `json:"used"`
	Available int `json:"available"`
	IsFull    bool `json:"is_full"`
}

type SubmitExamResultRequest struct {
	ResultValue float64 `json:"result_value" binding:"required"`
}

type ExamResultResponse struct {
	ID            string                   `json:"id"`
	AppointmentID string                   `json:"appointment_id"`
	ItemID        string                   `json:"item_id"`
	ItemName      string                   `json:"item_name"`
	Department    model.Department         `json:"department"`
	ResultValue   *float64                 `json:"result_value"`
	Unit          string                   `json:"unit"`
	RefRange      string                   `json:"ref_range"`
	IsAbnormal    bool                     `json:"is_abnormal"`
	AbnormalType  string                   `json:"abnormal_type"`
	IsCritical    bool                     `json:"is_critical"`
	Status        model.ExamResultStatus   `json:"status"`
	ModifyCount   int                      `json:"modify_count"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

type CriticalAlertResponse struct {
	ID            string    `json:"id"`
	ExamNumber    string    `json:"exam_number"`
	CustomerName  string    `json:"customer_name"`
	ItemName      string    `json:"item_name"`
	ResultValue   float64   `json:"result_value"`
	RefRange      string    `json:"ref_range"`
	CreatedAt     time.Time `json:"created_at"`
	Resolved      bool      `json:"resolved"`
}

type UpdateReportRequest struct {
	GeneralAdvice  string `json:"general_advice"`
	FollowUpAdvice string `json:"follow_up_advice"`
}

type ReportResponse struct {
	ID              string               `json:"id"`
	ExamNumber      string               `json:"exam_number"`
	CustomerName    string               `json:"customer_name"`
	Gender          string               `json:"gender"`
	Age             int                  `json:"age"`
	PackageName     string               `json:"package_name"`
	Items           []ReportItemResponse `json:"items"`
	HasAbnormal     bool                 `json:"has_abnormal"`
	GeneralAdvice   string               `json:"general_advice"`
	FollowUpAdvice  string               `json:"follow_up_advice"`
	Status          model.ReportStatus   `json:"status"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	PublishedAt     *time.Time           `json:"published_at"`
}

type ReportItemResponse struct {
	ItemID       string           `json:"item_id"`
	ItemName     string           `json:"item_name"`
	Department   model.Department `json:"department"`
	ResultValue  *float64         `json:"result_value"`
	Unit         string           `json:"unit"`
	RefRange     string           `json:"ref_range"`
	IsAbnormal   bool             `json:"is_abnormal"`
	AbnormalType string           `json:"abnormal_type"`
	IsCritical   bool             `json:"is_critical"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Size       int         `json:"size"`
	Total      int64       `json:"total"`
	TotalPages int64       `json:"total_pages"`
}

type ExportRecordsRequest struct {
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
