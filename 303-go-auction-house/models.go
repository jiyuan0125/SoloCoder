package main

import (
	"errors"
	"time"
)

type AuctionStatus string

const (
	StatusActive   AuctionStatus = "active"
	StatusClosed   AuctionStatus = "closed"
	StatusSold     AuctionStatus = "sold"
	StatusUnsold   AuctionStatus = "unsold"
)

type Auction struct {
	ID             string        `json:"id"`
	SellerID       string        `json:"seller_id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	StartingPrice  float64       `json:"starting_price"`
	Increment      float64       `json:"increment"`
	EndTime        time.Time     `json:"end_time"`
	Status         AuctionStatus `json:"status"`
	CurrentPrice   float64       `json:"current_price"`
	WinningBidID   string        `json:"winning_bid_id,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

type Bid struct {
	ID        string    `json:"id"`
	AuctionID string    `json:"auction_id"`
	BidderID  string    `json:"bidder_id"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

type CreateAuctionRequest struct {
	SellerID      string  `json:"seller_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	StartingPrice float64 `json:"starting_price"`
	Increment     float64 `json:"increment"`
	EndTime       string  `json:"end_time"`
}

type CreateBidRequest struct {
	BidderID string  `json:"bidder_id"`
	Amount   float64 `json:"amount"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var (
	ErrInvalidName           = errors.New("拍品名称不能为空")
	ErrInvalidDescription    = errors.New("拍品描述不能超过500字")
	ErrInvalidStartingPrice  = errors.New("起拍价必须大于零")
	ErrInvalidIncrement      = errors.New("加价幅度必须大于零")
	ErrInvalidEndTime        = errors.New("截止时间必须在未来")
	ErrAuctionNotFound       = errors.New("拍品不存在")
	ErrAuctionEnded          = errors.New("该拍品竞拍已结束")
	ErrBidTooLow             = errors.New("出价已被超越")
	ErrInvalidBidAmount      = errors.New("出价金额无效")
)
