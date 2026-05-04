package common

import "time"

type AuctionStatus string

const (
	StatusActive   AuctionStatus = "active"
	StatusClosed   AuctionStatus = "closed"
	StatusUnsold   AuctionStatus = "unsold"
)

type Auction struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	StartPrice      float64       `json:"start_price"`
	BidIncrement    float64       `json:"bid_increment"`
	CurrentPrice    float64       `json:"current_price"`
	Deadline        time.Time     `json:"deadline"`
	SellerID        string        `json:"seller_id"`
	Status          AuctionStatus `json:"status"`
	WinnerID        string        `json:"winner_id,omitempty"`
	WinningPrice    float64       `json:"winning_price,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

type Bid struct {
	ID         string    `json:"id"`
	AuctionID  string    `json:"auction_id"`
	BidderID   string    `json:"bidder_id"`
	Price      float64   `json:"price"`
	BidTime    time.Time `json:"bid_time"`
}

type CreateAuctionRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	StartPrice    float64 `json:"start_price"`
	BidIncrement  float64 `json:"bid_increment"`
	Deadline      string  `json:"deadline"`
	SellerID      string  `json:"seller_id"`
}

type CreateAuctionResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    *Auction `json:"data,omitempty"`
}

type BidRequest struct {
	AuctionID string  `json:"auction_id"`
	BidderID  string  `json:"bidder_id"`
	Price     float64 `json:"price"`
}

type BidResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    *Bid    `json:"data,omitempty"`
}

type GetAuctionRequest struct {
	ID string `json:"id"`
}

type GetAuctionResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    *Auction `json:"data,omitempty"`
}

type GetSellerAuctionsRequest struct {
	SellerID string `json:"seller_id"`
}

type GetSellerAuctionsResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []*Auction `json:"data,omitempty"`
}

type GetBidHistoryRequest struct {
	AuctionID string `json:"auction_id"`
}

type GetBidHistoryResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    []*Bid  `json:"data,omitempty"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
