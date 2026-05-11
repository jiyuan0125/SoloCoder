package common

type RegisterUserRequest struct {
	Username string   `json:"username"`
	Role     UserRole `json:"role,omitempty"`
}

type CreateItemRequest struct {
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      Category  `json:"category"`
	Condition     Condition `json:"condition"`
	OriginalPrice int64     `json:"original_price"`
	Price         int64     `json:"price"`
}

type ApproveItemRequest struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
}

type SearchItemsRequest struct {
	Keyword     string    `json:"keyword,omitempty"`
	Category    Category  `json:"category,omitempty"`
	MinPrice    int64     `json:"min_price,omitempty"`
	MaxPrice    int64     `json:"max_price,omitempty"`
	Condition   Condition `json:"condition,omitempty"`
}

type MakeOfferRequest struct {
	Price int64 `json:"price"`
}

type RespondOfferRequest struct {
	Action    string `json:"action"`
	CounterPrice int64 `json:"counter_price,omitempty"`
}

type ShipOrderRequest struct {
	LogisticsNo string `json:"logistics_no"`
}

type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type UserResponse struct {
	BaseResponse
	Data *User `json:"data,omitempty"`
}

type ItemResponse struct {
	BaseResponse
	Data *Item `json:"data,omitempty"`
}

type ItemsResponse struct {
	BaseResponse
	Data []Item `json:"data,omitempty"`
}

type NegotiationResponse struct {
	BaseResponse
	Data *Negotiation `json:"data,omitempty"`
}

type NegotiationsResponse struct {
	BaseResponse
	Data []Negotiation `json:"data,omitempty"`
}

type OrderResponse struct {
	BaseResponse
	Data *Order `json:"data,omitempty"`
}

type OrdersResponse struct {
	BaseResponse
	Data []Order `json:"data,omitempty"`
}
