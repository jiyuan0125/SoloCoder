package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	auctionService *AuctionService
	bidService     *BidService
	rulesService   *RulesService
	store          *Store
}

func NewHandler(auctionService *AuctionService, bidService *BidService, rulesService *RulesService, store *Store) *Handler {
	return &Handler{
		auctionService: auctionService,
		bidService:     bidService,
		rulesService:   rulesService,
		store:          store,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		h.respondError(w, http.StatusNotFound, "路由不存在")
		return
	}

	switch r.Method {
	case http.MethodPost:
		if parts[0] == "auctions" {
			h.createAuction(w, r)
			return
		}
		if len(parts) == 3 && parts[0] == "auctions" && parts[2] == "bids" {
			h.placeBid(w, r, parts[1])
			return
		}

	case http.MethodGet:
		if len(parts) == 2 && parts[0] == "auctions" {
			h.getAuction(w, r, parts[1])
			return
		}
		if len(parts) == 3 && parts[0] == "auctions" && parts[2] == "bids" {
			h.getBidHistory(w, r, parts[1])
			return
		}
		if len(parts) == 2 && parts[0] == "sellers" {
			h.getSellerAuctions(w, r, parts[1])
			return
		}
	}

	h.respondError(w, http.StatusNotFound, "路由不存在")
}

func (h *Handler) createAuction(w http.ResponseWriter, r *http.Request) {
	var req CreateAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if err := h.rulesService.ValidateCreateAuction(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	auction, err := h.auctionService.CreateAuction(&req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.store.Save(); err != nil {
		h.respondError(w, http.StatusInternalServerError, "保存数据失败")
		return
	}

	h.respondJSON(w, http.StatusCreated, SuccessResponse{
		Message: "拍品创建成功",
		Data:    auction,
	})
}

func (h *Handler) getAuction(w http.ResponseWriter, r *http.Request, auctionID string) {
	h.auctionService.CheckAndCloseExpiredAuctions()

	auction, err := h.auctionService.GetAuction(auctionID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	type AuctionDetail struct {
		*Auction
		WinningBid *Bid `json:"winning_bid,omitempty"`
	}

	detail := AuctionDetail{Auction: auction}

	if auction.Status == StatusSold && auction.WinningBidID != "" {
		winningBid, _ := h.bidService.GetWinningBid(auctionID)
		detail.WinningBid = winningBid
	}

	h.respondJSON(w, http.StatusOK, detail)
}

func (h *Handler) getSellerAuctions(w http.ResponseWriter, r *http.Request, sellerID string) {
	h.auctionService.CheckAndCloseExpiredAuctions()

	auctions := h.auctionService.GetSellerAuctions(sellerID)
	h.respondJSON(w, http.StatusOK, auctions)
}

func (h *Handler) placeBid(w http.ResponseWriter, r *http.Request, auctionID string) {
	h.auctionService.CheckAndCloseExpiredAuctions()

	var req CreateBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	bid, err := h.bidService.PlaceBid(auctionID, req.BidderID, req.Amount)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrAuctionNotFound {
			status = http.StatusNotFound
		} else if err == ErrAuctionEnded {
			status = http.StatusForbidden
		}
		h.respondError(w, status, err.Error())
		return
	}

	if err := h.store.Save(); err != nil {
		h.respondError(w, http.StatusInternalServerError, "保存数据失败")
		return
	}

	h.respondJSON(w, http.StatusCreated, SuccessResponse{
		Message: "出价成功",
		Data:    bid,
	})
}

func (h *Handler) getBidHistory(w http.ResponseWriter, r *http.Request, auctionID string) {
	_, err := h.auctionService.GetAuction(auctionID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	bids, err := h.bidService.GetBidHistory(auctionID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, bids)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
