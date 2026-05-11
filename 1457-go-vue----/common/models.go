package common

import "time"

type ComplaintChannel string

const (
	ChannelPhone  ComplaintChannel = "phone"
	ChannelWeChat ComplaintChannel = "wechat"
	ChannelAPP    ComplaintChannel = "app"
	ChannelOnsite ComplaintChannel = "onsite"
)

type ComplaintType string

const (
	TypeServiceQuality ComplaintType = "service_quality"
	TypeProductQuality ComplaintType = "product_quality"
	TypeLogistics      ComplaintType = "logistics"
	TypeAfterSales     ComplaintType = "after_sales"
	TypeOther          ComplaintType = "other"
)

type UrgencyLevel string

const (
	UrgencyNormal UrgencyLevel = "normal"
	UrgencyUrgent UrgencyLevel = "urgent"
	UrgencySuper  UrgencyLevel = "super"
)

type TicketStatus string

const (
	StatusPendingDispatch TicketStatus = "pending_dispatch"
	StatusDispatched      TicketStatus = "dispatched"
	StatusProcessing      TicketStatus = "processing"
	StatusPendingReview   TicketStatus = "pending_review"
	StatusTimeout         TicketStatus = "timeout"
	StatusClosed          TicketStatus = "closed"
)

type ReviewResult string

const (
	ReviewSatisfied       ReviewResult = "satisfied"
	ReviewBasiclySatisfied ReviewResult = "basicly_satisfied"
	ReviewDissatisfied    ReviewResult = "dissatisfied"
)

type CreateTicketRequest struct {
	ComplainerName    string           `json:"complainer_name"`
	ContactPhone      string           `json:"contact_phone"`
	Content           string           `json:"content"`
	ComplaintChannel  ComplaintChannel `json:"complaint_channel"`
	ComplaintType     ComplaintType    `json:"complaint_type"`
	Region            string           `json:"region"`
	ExternalSystemID  string           `json:"external_system_id,omitempty"`
}

type CreateTicketResponse struct {
	TicketNo string `json:"ticket_no"`
}

type DispatchTicketRequest struct {
	ResponsibleDepartment string       `json:"responsible_department"`
	ResponsiblePerson     string       `json:"responsible_person"`
	UrgencyLevel          UrgencyLevel `json:"urgency_level"`
	DispatcherID          string       `json:"dispatcher_id"`
}

type CompleteProcessingRequest struct {
	ProcessingResult string `json:"processing_result"`
}

type ReviewRequest struct {
	ReviewResult  ReviewResult `json:"review_result"`
	Remark        string       `json:"remark"`
	ReviewerID    string       `json:"reviewer_id"`
}

type TicketResponse struct {
	TicketNo              string         `json:"ticket_no"`
	ComplainerName        string         `json:"complainer_name"`
	ContactPhone          string         `json:"contact_phone"`
	Content               string         `json:"content"`
	ComplaintChannel      ComplaintChannel `json:"complaint_channel"`
	ComplaintType         ComplaintType  `json:"complaint_type"`
	Region                string         `json:"region"`
	Status                TicketStatus   `json:"status"`
	UrgencyLevel          *UrgencyLevel  `json:"urgency_level,omitempty"`
	ResponsibleDepartment string         `json:"responsible_department,omitempty"`
	ResponsiblePerson     string         `json:"responsible_person,omitempty"`
	ProcessingResult      string         `json:"processing_result,omitempty"`
	ReviewResult          *ReviewResult  `json:"review_result,omitempty"`
	ReviewRemark          string         `json:"review_remark,omitempty"`
	RetryCount            int            `json:"retry_count"`
	IsEscalated           bool           `json:"is_escalated"`
	ExternalSystemID      string         `json:"external_system_id,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	DispatchedAt          *time.Time     `json:"dispatched_at,omitempty"`
	ProcessedAt           *time.Time     `json:"processed_at,omitempty"`
	ClosedAt              *time.Time     `json:"closed_at,omitempty"`
	ResponseDeadline      time.Time      `json:"response_deadline"`
}

type StatisticsResponse struct {
	TodayNewCount        int     `json:"today_new_count"`
	ProcessingCount      int     `json:"processing_count"`
	ClosedCount          int     `json:"closed_count"`
	AvgProcessingHours   float64 `json:"avg_processing_hours"`
	ReviewSatisfaction   float64 `json:"review_satisfaction"`
}

type ListTicketsResponse struct {
	Tickets []TicketResponse `json:"tickets"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
