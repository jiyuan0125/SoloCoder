package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type AuctionService struct {
	store *Store
}

func NewAuctionService(store *Store) *AuctionService {
	return &AuctionService{store: store}
}

func (s *AuctionService) CreateAuction(req *CreateAuctionRequest) (*Auction, error) {
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, ErrInvalidEndTime
	}

	auction := &Auction{
		ID:            generateRandomID(),
		SellerID:      req.SellerID,
		Name:          req.Name,
		Description:   req.Description,
		StartingPrice: req.StartingPrice,
		Increment:     req.Increment,
		EndTime:       endTime,
		Status:        StatusActive,
		CurrentPrice:  req.StartingPrice,
		CreatedAt:     time.Now(),
	}

	s.store.CreateAuction(auction)
	return auction, nil
}

func (s *AuctionService) GetAuction(auctionID string) (*Auction, error) {
	auction, exists := s.store.GetAuction(auctionID)
	if !exists {
		return nil, ErrAuctionNotFound
	}
	return auction, nil
}

func (s *AuctionService) GetSellerAuctions(sellerID string) []*Auction {
	return s.store.GetAuctionsBySeller(sellerID)
}

func (s *AuctionService) CloseAuction(auctionID string) error {
	auction, exists := s.store.GetAuction(auctionID)
	if !exists {
		return ErrAuctionNotFound
	}

	bids := s.store.GetBids(auctionID)
	if len(bids) == 0 {
		auction.Status = StatusUnsold
	} else {
		var highestBid *Bid
		for _, bid := range bids {
			if highestBid == nil || bid.Amount > highestBid.Amount {
				highestBid = bid
			}
		}
		auction.Status = StatusSold
		auction.WinningBidID = highestBid.ID
		auction.CurrentPrice = highestBid.Amount
	}

	s.store.UpdateAuction(auction)
	return nil
}

func (s *AuctionService) IsAuctionActive(auction *Auction) bool {
	if auction.Status != StatusActive {
		return false
	}
	return time.Now().Before(auction.EndTime)
}

func (s *AuctionService) CheckAndCloseExpiredAuctions() {
	now := time.Now()
	s.store.Lock()
	defer s.store.Unlock()

	for id, auction := range s.store.Auctions {
		if auction.Status == StatusActive && now.After(auction.EndTime) {
			bids := s.store.Bids[id]
			if len(bids) == 0 {
				auction.Status = StatusUnsold
			} else {
				var highestBid *Bid
				for _, bid := range bids {
					if highestBid == nil || bid.Amount > highestBid.Amount {
						highestBid = bid
					}
				}
				auction.Status = StatusSold
				auction.WinningBidID = highestBid.ID
				auction.CurrentPrice = highestBid.Amount
			}
		}
	}
}

func generateRandomID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
