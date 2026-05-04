package main

import (
	"encoding/json"
	"os"
	"sync"
)

const dataFile = "auction_data.json"

type Store struct {
	mu       sync.RWMutex
	Auctions map[string]*Auction `json:"auctions"`
	Bids     map[string][]*Bid    `json:"bids"`
}

func NewStore() *Store {
	return &Store{
		Auctions: make(map[string]*Auction),
		Bids:     make(map[string][]*Bid),
	}
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(dataFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var loaded Store
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}

	s.Auctions = loaded.Auctions
	s.Bids = loaded.Bids

	return nil
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dataFile, data, 0644)
}

func (s *Store) GetAuction(auctionID string) (*Auction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	auction, exists := s.Auctions[auctionID]
	return auction, exists
}

func (s *Store) CreateAuction(auction *Auction) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Auctions[auction.ID] = auction
}

func (s *Store) UpdateAuction(auction *Auction) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Auctions[auction.ID] = auction
}

func (s *Store) GetAuctionsBySeller(sellerID string) []*Auction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var auctions []*Auction
	for _, auction := range s.Auctions {
		if auction.SellerID == sellerID {
			auctions = append(auctions, auction)
		}
	}
	return auctions
}

func (s *Store) GetBids(auctionID string) []*Bid {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Bids[auctionID]
}

func (s *Store) AddBid(bid *Bid) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Bids[bid.AuctionID] = append(s.Bids[bid.AuctionID], bid)
}

func (s *Store) Lock() {
	s.mu.Lock()
}

func (s *Store) Unlock() {
	s.mu.Unlock()
}

func (s *Store) RLock() {
	s.mu.RLock()
}

func (s *Store) RUnlock() {
	s.mu.RUnlock()
}
