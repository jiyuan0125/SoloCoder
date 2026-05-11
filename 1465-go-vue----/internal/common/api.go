package common

import "time"

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func Success(data interface{}) *Response {
	return NewResponse(200, "success", data)
}

func Error(code int, message string) *Response {
	return NewResponse(code, message, nil)
}

type RegisterGasOutletRequest struct {
	ID       string `json:"id"`
	Location string `json:"location"`
}

type RegisterWastewaterOutletRequest struct {
	ID string `json:"id"`
}

type GasReportRequest struct {
	OutletID     string             `json:"outlet_id"`
	ReportedAt   time.Time          `json:"reported_at"`
	Measurements map[string]float64 `json:"measurements"`
}

type WastewaterReportRequest struct {
	OutletID     string             `json:"outlet_id"`
	ReportedAt   time.Time          `json:"reported_at"`
	Measurements map[string]float64 `json:"measurements"`
}

type AddSolidWasteRequest struct {
	Name            string    `json:"name"`
	Category        int       `json:"category"`
	Amount          float64   `json:"amount"`
	StorageLocation string    `json:"storage_location"`
	GeneratedAt     time.Time `json:"generated_at"`
}

type DisposeSolidWasteRequest struct {
	ID             string `json:"id"`
	DisposalMethod string `json:"disposal_method"`
}

type QueryGasReportsRequest struct {
	OutletID string    `json:"outlet_id"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
}

type QueryWastewaterReportsRequest struct {
	OutletID string    `json:"outlet_id"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
}

type GetDailyReportRequest struct {
	OutletID string    `json:"outlet_id"`
	Date     time.Time `json:"date"`
}

type QueryAlarmsRequest struct {
	EntityType *int    `json:"entity_type"`
	EntityID   string  `json:"entity_id"`
	Level      *int    `json:"level"`
	Resolved   *bool   `json:"resolved"`
}

type ResolveAlarmRequest struct {
	ID string `json:"id"`
}

type GasOutletResponse struct {
	ID       string `json:"id"`
	Location string `json:"location"`
	Status   int    `json:"status"`
}

type MeasurementResponse struct {
	Value      float64 `json:"value"`
	Factor     string  `json:"factor"`
	Unit       string  `json:"unit"`
	IsValid    bool    `json:"is_valid"`
	IsAbnormal bool    `json:"is_abnormal"`
	IsExceeded bool    `json:"is_exceeded"`
}

type GasReportResponse struct {
	ID              string                 `json:"id"`
	OutletID        string                 `json:"outlet_id"`
	ReportedAt      time.Time              `json:"reported_at"`
	ReceivedAt      time.Time              `json:"received_at"`
	DataStatus      int                    `json:"data_status"`
	Measurements    map[string]MeasurementResponse `json:"measurements"`
	ExceededFactors []string               `json:"exceeded_factors"`
	AbnormalFactors []string               `json:"abnormal_factors"`
}

type WastewaterReportResponse struct {
	ID              string                 `json:"id"`
	OutletID        string                 `json:"outlet_id"`
	ReportedAt      time.Time              `json:"reported_at"`
	ReceivedAt      time.Time              `json:"received_at"`
	DataStatus      int                    `json:"data_status"`
	Measurements    map[string]MeasurementResponse `json:"measurements"`
	ExceededFactors []string               `json:"exceeded_factors"`
	AbnormalFactors []string               `json:"abnormal_factors"`
}

type DailyFactorDataResponse struct {
	Max      float64 `json:"max"`
	Min      float64 `json:"min"`
	Average  float64 `json:"average"`
	Count    int     `json:"count"`
	Unit     string  `json:"unit"`
	Exceeded bool    `json:"exceeded"`
}

type DailyReportResponse struct {
	Date     time.Time                      `json:"date"`
	OutletID string                         `json:"outlet_id"`
	Status   int                            `json:"status"`
	Data     map[string]DailyFactorDataResponse `json:"data"`
}

type SolidWasteRecordResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Category        int     `json:"category"`
	Amount          float64 `json:"amount"`
	StorageLocation string  `json:"storage_location"`
	DisposalMethod  string  `json:"disposal_method"`
	GeneratedAt     time.Time `json:"generated_at"`
	DisposedAt      *time.Time `json:"disposed_at,omitempty"`
	Status          int     `json:"status"`
}

type AlarmResponse struct {
	ID         string                 `json:"id"`
	Type       int                    `json:"type"`
	Level      int                    `json:"level"`
	EntityType int                    `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	RelatedData map[string]interface{} `json:"related_data"`
	Message    string                 `json:"message"`
	CreatedAt  time.Time              `json:"created_at"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	IsResolved bool                   `json:"is_resolved"`
}
