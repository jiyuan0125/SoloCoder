package common

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Item struct {
	ID          string     `json:"id"`
	SellerID    string     `json:"seller_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    Category   `json:"category"`
	Condition   Condition  `json:"condition"`
	OriginalPrice int64   `json:"original_price"`
	Price       int64      `json:"price"`
	Status      ItemStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type NegotiationRound struct {
	Round    int             `json:"round"`
	Role     NegotiationRole `json:"role"`
	Price    int64           `json:"price"`
	Time     time.Time       `json:"time"`
	Response string          `json:"response,omitempty"`
}

type Negotiation struct {
	ID        string              `json:"id"`
	ItemID    string              `json:"item_id"`
	BuyerID   string              `json:"buyer_id"`
	SellerID  string              `json:"seller_id"`
	Status    NegotiationStatus   `json:"status"`
	Rounds    []NegotiationRound  `json:"rounds"`
	FinalPrice int64              `json:"final_price,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	LastActiveAt time.Time        `json:"last_active_at"`
}

type Order struct {
	ID              string      `json:"id"`
	ItemID          string      `json:"item_id"`
	BuyerID         string      `json:"buyer_id"`
	SellerID        string      `json:"seller_id"`
	NegotiationID   string      `json:"negotiation_id"`
	Price           int64       `json:"price"`
	ServiceFee      int64       `json:"service_fee"`
	Status          OrderStatus `json:"status"`
	LogisticsNo     string      `json:"logistics_no,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	PaidAt          *time.Time  `json:"paid_at,omitempty"`
	ShippedAt       *time.Time  `json:"shipped_at,omitempty"`
	CompletedAt     *time.Time  `json:"completed_at,omitempty"`
	StatusUpdatedAt time.Time   `json:"status_updated_at"`
}
