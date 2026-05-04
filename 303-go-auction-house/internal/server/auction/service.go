package auction

import (
	"auction-house/internal/common"
	"auction-house/internal/server/store"
	"sync"
	"time"
)

type AuctionService struct {
	store *store.DataStore
	ticker *time.Ticker
	stopCh chan struct{}
	once sync.Once
}

var service *AuctionService
var serviceOnce sync.Once

func GetService() *AuctionService {
	serviceOnce.Do(func() {
		service = &AuctionService{
			store:  store.GetStore(),
			stopCh: make(chan struct{}),
		}
		service.startAuctionChecker()
	})
	return service
}

func (s *AuctionService) startAuctionChecker() {
	s.once.Do(func() {
		s.ticker = time.NewTicker(1 * time.Second)
		go func() {
			for {
				select {
				case <-s.ticker.C:
					s.checkAndCloseExpiredAuctions()
				case <-s.stopCh:
					s.ticker.Stop()
					return
				}
			}
		}()
	})
}

func (s *AuctionService) Stop() {
	close(s.stopCh)
}

func (s *AuctionService) checkAndCloseExpiredAuctions() {
	activeAuctions := s.store.GetAllActiveAuctions()
	now := time.Now()

	for _, auction := range activeAuctions {
		if now.After(auction.Deadline) {
			s.store.CloseAuction(auction.ID)
		}
	}
}

func (s *AuctionService) CreateAuction(req *common.CreateAuctionRequest) (*common.Auction, int) {
	return s.store.CreateAuction(req)
}

func (s *AuctionService) GetAuction(id string) (*common.Auction, int) {
	return s.store.GetAuction(id)
}

func (s *AuctionService) GetSellerAuctions(sellerID string) []*common.Auction {
	return s.store.GetSellerAuctions(sellerID)
}

func (s *AuctionService) PlaceBid(req *common.BidRequest) (*common.Bid, int) {
	return s.store.PlaceBid(req)
}

func (s *AuctionService) GetBidHistory(auctionID string) []*common.Bid {
	return s.store.GetBidHistory(auctionID)
}
