package common

const (
	ApplicationStatusPending      = "pending"
	ApplicationStatusAutoApproved = "auto_approved"
	ApplicationStatusApproved     = "approved"
	ApplicationStatusRejected     = "rejected"
	ApplicationStatusAppealed     = "appealed"

	AppealStatusPending  = "pending"
	AppealStatusResolved = "resolved"
	AppealStatusDismissed = "dismissed"

	SerialNumberLength = 12
	MaxDescriptionLength = 200
)

type WarrantyApplication struct {
	ID              string
	UserID          string
	SerialNumber    string
	PurchaseDate    string
	Description     string
	Status          string
	RejectReason    string
	SubmittedAt     int64
	AutoApproved    bool
	ApprovedAt      int64
	WarrantyExpiry  string
}

type Appeal struct {
	ID               string
	ApplicationID    string
	Reason           string
	Status           string
	SubmittedAt      int64
	AdminNote        string
	ResolvedAt       int64
}

type SubmitApplicationRequest struct {
	UserID       string
	SerialNumber string
	PurchaseDate string
	Description  string
}

type SubmitApplicationResponse struct {
	Success       bool
	ApplicationID string
	Status        string
	Message       string
}

type ReviewApplicationRequest struct {
	ApplicationID string
	Approved      bool
	Reason        string
}

type SubmitAppealRequest struct {
	ApplicationID string
	Reason        string
}

type GetApplicationsResponse struct {
	Success      bool
	Applications []WarrantyApplication
}

type GetAppealsResponse struct {
	Success bool
	Appeals []AppealWithApplication
}

type AppealWithApplication struct {
	Appeal      Appeal
	Application WarrantyApplication
}

type StatisticsResponse struct {
	Success                    bool
	TotalApplications          int
	PendingApplications        int
	ApprovedApplications       int
	RejectedApplications       int
	AutoApprovedApplications   int
	AppealedApplications       int
}
