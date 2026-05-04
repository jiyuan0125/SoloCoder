package main

import (
	"time"
)

type RulesService struct{}

func NewRulesService() *RulesService {
	return &RulesService{}
}

func (s *RulesService) ValidateCreateAuction(req *CreateAuctionRequest) error {
	if req.Name == "" {
		return ErrInvalidName
	}

	if len([]rune(req.Description)) > 500 {
		return ErrInvalidDescription
	}

	if req.StartingPrice <= 0 {
		return ErrInvalidStartingPrice
	}

	if req.Increment <= 0 {
		return ErrInvalidIncrement
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return ErrInvalidEndTime
	}

	if endTime.Before(time.Now()) {
		return ErrInvalidEndTime
	}

	return nil
}

func (s *RulesService) ValidateBid(auction *Auction, bidAmount float64) error {
	if bidAmount <= 0 {
		return ErrInvalidBidAmount
	}

	requiredPrice := auction.CurrentPrice + auction.Increment
	if bidAmount < requiredPrice {
		return ErrBidTooLow
	}

	return nil
}

func (s *RulesService) GetMinimumBidAmount(auction *Auction) float64 {
	return auction.CurrentPrice + auction.Increment
}
