package handler

import (
	"auction-house/internal/common"
	"auction-house/internal/server/auction"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func writeJSONResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := common.APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func CreateAuctionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "方法不允许", nil)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "读取请求体失败", nil)
		return
	}
	defer r.Body.Close()

	var req common.CreateAuctionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "请求参数解析失败", nil)
		return
	}

	service := auction.GetService()
	auctionItem, code := service.CreateAuction(&req)

	if code != common.ErrCodeSuccess {
		writeJSONResponse(w, code, common.GetErrorMessage(code), nil)
		return
	}

	writeJSONResponse(w, common.ErrCodeSuccess, common.GetErrorMessage(common.ErrCodeSuccess), auctionItem)
}

func GetAuctionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "方法不允许", nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/auctions/")
	if id == "" {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "缺少拍品ID", nil)
		return
	}

	service := auction.GetService()
	auctionItem, code := service.GetAuction(id)

	if code != common.ErrCodeSuccess {
		writeJSONResponse(w, code, common.GetErrorMessage(code), nil)
		return
	}

	writeJSONResponse(w, common.ErrCodeSuccess, common.GetErrorMessage(common.ErrCodeSuccess), auctionItem)
}

func GetSellerAuctionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "方法不允许", nil)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "读取请求体失败", nil)
		return
	}
	defer r.Body.Close()

	var req common.GetSellerAuctionsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "请求参数解析失败", nil)
		return
	}

	service := auction.GetService()
	auctions := service.GetSellerAuctions(req.SellerID)

	writeJSONResponse(w, common.ErrCodeSuccess, common.GetErrorMessage(common.ErrCodeSuccess), auctions)
}

func PlaceBidHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "方法不允许", nil)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "读取请求体失败", nil)
		return
	}
	defer r.Body.Close()

	var req common.BidRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "请求参数解析失败", nil)
		return
	}

	service := auction.GetService()
	bid, code := service.PlaceBid(&req)

	if code != common.ErrCodeSuccess {
		writeJSONResponse(w, code, common.GetErrorMessage(code), nil)
		return
	}

	writeJSONResponse(w, common.ErrCodeSuccess, common.GetErrorMessage(common.ErrCodeSuccess), bid)
}

func GetBidHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "方法不允许", nil)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "读取请求体失败", nil)
		return
	}
	defer r.Body.Close()

	var req common.GetBidHistoryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONResponse(w, common.ErrCodeInvalidRequest, "请求参数解析失败", nil)
		return
	}

	service := auction.GetService()
	bids := service.GetBidHistory(req.AuctionID)

	writeJSONResponse(w, common.ErrCodeSuccess, common.GetErrorMessage(common.ErrCodeSuccess), bids)
}

func SetupRoutes() {
	http.HandleFunc("/auctions/create", CreateAuctionHandler)
	http.HandleFunc("/auctions/", GetAuctionHandler)
	http.HandleFunc("/auctions/seller", GetSellerAuctionsHandler)
	http.HandleFunc("/bids/place", PlaceBidHandler)
	http.HandleFunc("/bids/history", GetBidHistoryHandler)
}
