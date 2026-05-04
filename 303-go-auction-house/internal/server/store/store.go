package store

import (
	"auction-house/internal/common"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	dataDir     = "./data"
	auctionFile = "auctions.json"
	bidFile     = "bids.json"
)

type DataStore struct {
	auctions map[string]*common.Auction
	bids     map[string][]*common.Bid
	mu       sync.RWMutex
}

var instance *DataStore
var once sync.Once

func GetStore() *DataStore {
	once.Do(func() {
		instance = &DataStore{
			auctions: make(map[string]*common.Auction),
			bids:     make(map[string][]*common.Bid),
		}
		instance.loadFromDisk()
	})
	return instance
}

func (s *DataStore) loadFromDisk() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	auctionPath := filepath.Join(dataDir, auctionFile)
	bidPath := filepath.Join(dataDir, bidFile)

	if _, err := os.Stat(auctionPath); err == nil {
		data, err := os.ReadFile(auctionPath)
		if err == nil {
			json.Unmarshal(data, &s.auctions)
		}
	}

	if _, err := os.Stat(bidPath); err == nil {
		data, err := os.ReadFile(bidPath)
		if err == nil {
			var allBids []*common.Bid
			json.Unmarshal(data, &allBids)
			for _, bid := range allBids {
				s.bids[bid.AuctionID] = append(s.bids[bid.AuctionID], bid)
			}
		}
	}

	return nil
}

func (s *DataStore) saveToDisk() error {
	os.MkdirAll(dataDir, 0755)

	auctionPath := filepath.Join(dataDir, auctionFile)
	bidPath := filepath.Join(dataDir, bidFile)

	auctionData, err := json.MarshalIndent(s.auctions, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(auctionPath, auctionData, 0644); err != nil {
		return err
	}

	var allBids []*common.Bid
	for _, bids := range s.bids {
		allBids = append(allBids, bids...)
	}
	bidData, err := json.MarshalIndent(allBids, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(bidPath, bidData, 0644); err != nil {
		return err
	}

	return nil
}

func (s *DataStore) CreateAuction(req *common.CreateAuctionRequest) (*common.Auction, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		return nil, common.ErrCodeInvalidDeadline
	}

	now := time.Now()
	if deadline.Before(now) || deadline.Equal(now) {
		return nil, common.ErrCodeInvalidDeadline
	}

	if req.Name == "" {
		return nil, common.ErrCodeInvalidName
	}

	if len(req.Description) > 500 {
		return nil, common.ErrCodeDescriptionTooLong
	}

	if req.StartPrice <= 0 {
		return nil, common.ErrCodeInvalidStartPrice
	}

	if req.BidIncrement <= 0 {
		return nil, common.ErrCodeInvalidBidIncrement
	}

	auction := &common.Auction{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		StartPrice:   req.StartPrice,
		BidIncrement: req.BidIncrement,
		CurrentPrice: req.StartPrice,
		Deadline:     deadline,
		SellerID:     req.SellerID,
		Status:       common.StatusActive,
		CreatedAt:    now,
	}

	s.auctions[auction.ID] = auction
	s.saveToDisk()

	return auction, common.ErrCodeSuccess
}

func (s *DataStore) GetAuction(id string) (*common.Auction, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	auction, exists := s.auctions[id]
	if !exists {
		return nil, common.ErrCodeAuctionNotFound
	}

	return auction, common.ErrCodeSuccess
}

func (s *DataStore) GetSellerAuctions(sellerID string) []*common.Auction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var auctions []*common.Auction
	for _, auction := range s.auctions {
		if auction.SellerID == sellerID {
			auctions = append(auctions, auction)
		}
	}

	return auctions
}

func (s *DataStore) GetAllActiveAuctions() []*common.Auction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var auctions []*common.Auction
	for _, auction := range s.auctions {
		if auction.Status == common.StatusActive {
			auctions = append(auctions, auction)
		}
	}

	return auctions
}

func (s *DataStore) PlaceBid(req *common.BidRequest) (*common.Bid, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	auction, exists := s.auctions[req.AuctionID]
	if !exists {
		return nil, common.ErrCodeAuctionNotFound
	}

	if auction.Status != common.StatusActive {
		return nil, common.ErrCodeAuctionAlreadyEnded
	}

	now := time.Now()
	if now.After(auction.Deadline) {
		return nil, common.ErrCodeAuctionAlreadyEnded
	}

	requiredPrice := auction.CurrentPrice + auction.BidIncrement
	if req.Price < requiredPrice {
		return nil, common.ErrCodeBidTooLow
	}

	if req.Price == auction.CurrentPrice {
		return nil, common.ErrCodeBidAlreadyExceeded
	}

	bid := &common.Bid{
		ID:        uuid.New().String(),
		AuctionID: req.AuctionID,
		BidderID:  req.BidderID,
		Price:     req.Price,
		BidTime:   now,
	}

	s.bids[req.AuctionID] = append(s.bids[req.AuctionID], bid)
	auction.CurrentPrice = req.Price

	s.saveToDisk()

	return bid, common.ErrCodeSuccess
}

func (s *DataStore) GetBidHistory(auctionID string) []*common.Bid {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bids := s.bids[auctionID]
	result := make([]*common.Bid, len(bids))
	copy(result, bids)

	for i := len(result) - 1; i > 0; i-- {
		for j := 0; j < i; j++ {
			if result[j].Price < result[j+1].Price {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return result
}

func (s *DataStore) CloseAuction(auctionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	auction, exists := s.auctions[auctionID]
	if !exists || auction.Status != common.StatusActive {
		return
	}

	bids := s.bids[auctionID]
	if len(bids) == 0 {
		auction.Status = common.StatusUnsold
	} else {
		var highestBid *common.Bid
		for _, bid := range bids {
			if highestBid == nil || bid.Price > highestBid.Price {
				highestBid = bid
			}
		}

		auction.Status = common.StatusClosed
		auction.WinnerID = highestBid.BidderID
		auction.WinningPrice = highestBid.Price
	}

	s.saveToDisk()
}
