package main

import (
	"sync"
	"time"
)

type BidService struct {
	store          *Store
	auctionService *AuctionService
	auctionLocks   map[string]*sync.Mutex
	globalLock     sync.Mutex
}

func NewBidService(store *Store, auctionService *AuctionService) *BidService {
	return &BidService{
		store:          store,
		auctionService: auctionService,
		auctionLocks:   make(map[string]*sync.Mutex),
	}
}

func (s *BidService) getAuctionLock(auctionID string) *sync.Mutex {
	s.globalLock.Lock()
	defer s.globalLock.Unlock()

	lock, exists := s.auctionLocks[auctionID]
	if !exists {
		lock = &sync.Mutex{}
		s.auctionLocks[auctionID] = lock
	}
	return lock
}

func (s *BidService) PlaceBid(auctionID string, bidderID string, amount float64) (*Bid, error) {
	lock := s.getAuctionLock(auctionID)
	lock.Lock()
	defer lock.Unlock()

	auction, exists := s.store.GetAuction(auctionID)
	if !exists {
		return nil, ErrAuctionNotFound
	}

	if !s.auctionService.IsAuctionActive(auction) {
		return nil, ErrAuctionEnded
	}

	minimumBid := auction.CurrentPrice + auction.Increment
	if amount < minimumBid {
		return nil, ErrBidTooLow
	}

	bid := &Bid{
		ID:        generateRandomID(),
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    amount,
		Timestamp: time.Now(),
	}

	s.store.AddBid(bid)

	auction.CurrentPrice = amount
	s.store.UpdateAuction(auction)

	return bid, nil
}

func (s *BidService) GetBidHistory(auctionID string) ([]*Bid, error) {
	_, exists := s.store.GetAuction(auctionID)
	if !exists {
		return nil, ErrAuctionNotFound
	}

	bids := s.store.GetBids(auctionID)

	sortedBids := make([]*Bid, len(bids))
	copy(sortedBids, bids)

	for i := 0; i < len(sortedBids)-1; i++ {
		for j := i + 1; j < len(sortedBids); j++ {
			if sortedBids[i].Amount < sortedBids[j].Amount {
				sortedBids[i], sortedBids[j] = sortedBids[j], sortedBids[i]
			}
		}
	}

	return sortedBids, nil
}

func (s *BidService) GetWinningBid(auctionID string) (*Bid, error) {
	auction, exists := s.store.GetAuction(auctionID)
	if !exists {
		return nil, ErrAuctionNotFound
	}

	if auction.WinningBidID == "" {
		return nil, nil
	}

	bids := s.store.GetBids(auctionID)
	for _, bid := range bids {
		if bid.ID == auction.WinningBidID {
			return bid, nil
		}
	}

	return nil, nil
}


