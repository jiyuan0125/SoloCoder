package api

type LockerSize string

const (
	SizeSmall  LockerSize = "small"
	SizeMedium LockerSize = "medium"
	SizeLarge  LockerSize = "large"
)

type CompartmentStatus string

const (
	StatusFree      CompartmentStatus = "free"
	StatusOccupied  CompartmentStatus = "occupied"
	StatusOverdue   CompartmentStatus = "overdue"
)

type CreateLockerRequest struct {
	ID            string `json:"id"`
	Address       string `json:"address"`
	SmallCount    int    `json:"small_count"`
	MediumCount   int    `json:"medium_count"`
	LargeCount    int    `json:"large_count"`
	BusinessStart string `json:"business_start"`
	BusinessEnd   string `json:"business_end"`
}

type StorePackageRequest struct {
	LockerID      string `json:"locker_id"`
	CourierID     string `json:"courier_id"`
	RecipientPhone string `json:"recipient_phone"`
	Size          LockerSize `json:"size"`
}

type StorePackageResponse struct {
	CompartmentID string `json:"compartment_id"`
	PickupCode    string `json:"pickup_code"`
}

type PickupPackageRequest struct {
	LockerID      string `json:"locker_id"`
	PickupCode    string `json:"pickup_code"`
}

type PickupPackageResponse struct {
	CompartmentID string `json:"compartment_id"`
	Success       bool   `json:"success"`
}

type ListCompartmentsRequest struct {
	LockerID string `json:"locker_id"`
}

type CompartmentInfo struct {
	ID           string          `json:"id"`
	Size         LockerSize      `json:"size"`
	Status       CompartmentStatus `json:"status"`
	CourierID    string          `json:"courier_id,omitempty"`
	RecipientPhone string        `json:"recipient_phone,omitempty"`
	PickupCode   string          `json:"pickup_code,omitempty"`
	StoreTime    string          `json:"store_time,omitempty"`
	ExpireTime   string          `json:"expire_time,omitempty"`
}

type ListCompartmentsResponse struct {
	Compartments []CompartmentInfo `json:"compartments"`
}

type CreateCourierRequest struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type CourierInfo struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type GetCourierRequest struct {
	ID string `json:"id"`
}

type RechargeCourierRequest struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

type ListOverdueRequest struct {
}

type OverdueInfo struct {
	LockerID      string `json:"locker_id"`
	CompartmentID string `json:"compartment_id"`
	CourierID     string `json:"courier_id"`
	RecipientPhone string `json:"recipient_phone"`
	StoreTime     string `json:"store_time"`
	ExpireTime    string `json:"expire_time"`
}

type ListOverdueResponse struct {
	OverduePackages []OverdueInfo `json:"overdue_packages"`
}

type HandleOverdueRequest struct {
	LockerID      string `json:"locker_id"`
	CompartmentID string `json:"compartment_id"`
	CourierID     string `json:"courier_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
	Message string `json:"message,omitempty"`
}
