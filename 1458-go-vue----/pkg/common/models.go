package common

import "time"

type CustomerStatus string

const (
	StatusNew       CustomerStatus = "NEW"
	StatusFollowing CustomerStatus = "FOLLOWING"
	StatusWarning   CustomerStatus = "WARNING"
	StatusSigned    CustomerStatus = "SIGNED"
)

type BusinessType string

const (
	BusinessTypeRestaurant BusinessType = "RESTAURANT"
	BusinessTypeRetail     BusinessType = "RETAIL"
	BusinessTypeOffice     BusinessType = "OFFICE"
	BusinessTypeWarehouse  BusinessType = "WAREHOUSE"
)

type FollowMethod string

const (
	FollowMethodPhone    FollowMethod = "PHONE"
	FollowMethodWeChat   FollowMethod = "WECHAT"
	FollowMethodMeeting  FollowMethod = "MEETING"
	FollowMethodEmail    FollowMethod = "EMAIL"
)

type FollowStatus string

const (
	FollowStatusOpen  FollowStatus = "OPEN"
	FollowStatusClose FollowStatus = "CLOSED"
)

type Customer struct {
	ID               string       `json:"id"`
	CustomerName     string       `json:"customer_name"`
	ContactPerson    string       `json:"contact_person"`
	ContactPhone     string       `json:"contact_phone"`
	IntentArea       float64      `json:"intent_area"`
	IntentBusiness   BusinessType `json:"intent_business"`
	ExpectedMoveIn   time.Time    `json:"expected_move_in"`
	Status           CustomerStatus `json:"status"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type FollowRecord struct {
	ID              string       `json:"id"`
	CustomerID      string       `json:"customer_id"`
	Method          FollowMethod `json:"method"`
	Content         string       `json:"content"`
	NextPlanTime    *time.Time   `json:"next_plan_time"`
	Status          FollowStatus `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type Contract struct {
	ID              string     `json:"id"`
	CustomerID      string     `json:"customer_id"`
	CustomerName    string     `json:"customer_name"`
	LeaseArea       float64    `json:"lease_area"`
	MonthlyRentRate float64    `json:"monthly_rent_rate"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	FreeRentDays    int        `json:"free_rent_days"`
	TotalRent       float64    `json:"total_rent"`
	CreatedAt       time.Time  `json:"created_at"`
}

type DashboardStats struct {
	MonthlyNewCustomers    int     `json:"monthly_new_customers"`
	MonthlySignedContracts int     `json:"monthly_signed_contracts"`
	SigningRate            float64 `json:"signing_rate"`
	AverageFollowCount     float64 `json:"average_follow_count"`
}

type CreateCustomerRequest struct {
	CustomerName   string       `json:"customer_name"`
	ContactPerson  string       `json:"contact_person"`
	ContactPhone   string       `json:"contact_phone"`
	IntentArea     float64      `json:"intent_area"`
	IntentBusiness BusinessType `json:"intent_business"`
	ExpectedMoveIn time.Time    `json:"expected_move_in"`
}

type CreateCustomerResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Customer *Customer `json:"customer,omitempty"`
}

type ListCustomersRequest struct {
	Status CustomerStatus `json:"status,omitempty"`
}

type ListCustomersResponse struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message"`
	Customers []*Customer `json:"customers"`
}

type GetCustomerRequest struct {
	ID string `json:"id"`
}

type GetCustomerResponse struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Customer *Customer `json:"customer,omitempty"`
}

type CreateFollowRequest struct {
	CustomerID   string       `json:"customer_id"`
	Method       FollowMethod `json:"method"`
	Content      string       `json:"content"`
	NextPlanTime *time.Time   `json:"next_plan_time"`
}

type CreateFollowResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Follow  *FollowRecord `json:"follow,omitempty"`
}

type ListFollowsRequest struct {
	CustomerID string `json:"customer_id"`
}

type ListFollowsResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Follows []*FollowRecord `json:"follows"`
}

type CloseFollowRequest struct {
	ID string `json:"id"`
}

type CloseFollowResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CreateContractRequest struct {
	CustomerID      string    `json:"customer_id"`
	LeaseArea       float64   `json:"lease_area"`
	MonthlyRentRate float64   `json:"monthly_rent_rate"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	FreeRentDays    int       `json:"free_rent_days"`
}

type CreateContractResponse struct {
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
	Contract *Contract `json:"contract,omitempty"`
}

type ListContractsRequest struct{}

type ListContractsResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Contracts []*Contract `json:"contracts"`
}

type GetDashboardRequest struct{}

type GetDashboardResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Stats   *DashboardStats `json:"stats,omitempty"`
}
