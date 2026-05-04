package client

import (
	"auction-house/internal/common"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultServerURL = "http://localhost:8080"

type APIClient struct {
	serverURL string
}

func NewAPIClient(serverURL string) *APIClient {
	if serverURL == "" {
		serverURL = defaultServerURL
	}
	return &APIClient{serverURL: serverURL}
}

func (c *APIClient) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.serverURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *APIClient) CreateAuction(req *common.CreateAuctionRequest) (*common.CreateAuctionResponse, error) {
	respBytes, err := c.doRequest(http.MethodPost, "/auctions/create", req)
	if err != nil {
		return nil, err
	}

	var resp common.CreateAuctionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetAuction(id string) (*common.GetAuctionResponse, error) {
	respBytes, err := c.doRequest(http.MethodGet, "/auctions/"+id, nil)
	if err != nil {
		return nil, err
	}

	var resp common.GetAuctionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetSellerAuctions(sellerID string) (*common.GetSellerAuctionsResponse, error) {
	req := common.GetSellerAuctionsRequest{SellerID: sellerID}
	respBytes, err := c.doRequest(http.MethodPost, "/auctions/seller", req)
	if err != nil {
		return nil, err
	}

	var resp common.GetSellerAuctionsResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) PlaceBid(req *common.BidRequest) (*common.BidResponse, error) {
	respBytes, err := c.doRequest(http.MethodPost, "/bids/place", req)
	if err != nil {
		return nil, err
	}

	var resp common.BidResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetBidHistory(auctionID string) (*common.GetBidHistoryResponse, error) {
	req := common.GetBidHistoryRequest{AuctionID: auctionID}
	respBytes, err := c.doRequest(http.MethodPost, "/bids/history", req)
	if err != nil {
		return nil, err
	}

	var resp common.GetBidHistoryResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func PrintAuction(auction *common.Auction) {
	fmt.Println("==============================")
	fmt.Printf("拍品ID: %s\n", auction.ID)
	fmt.Printf("名称: %s\n", auction.Name)
	fmt.Printf("描述: %s\n", auction.Description)
	fmt.Printf("起拍价: %.2f\n", auction.StartPrice)
	fmt.Printf("加价幅度: %.2f\n", auction.BidIncrement)
	fmt.Printf("当前价格: %.2f\n", auction.CurrentPrice)
	fmt.Printf("截止时间: %s\n", auction.Deadline.Format("2006-01-02 15:04:05"))
	fmt.Printf("卖家ID: %s\n", auction.SellerID)
	fmt.Printf("状态: %s\n", auction.Status)
	if auction.Status == common.StatusClosed {
		fmt.Printf("中标者: %s\n", auction.WinnerID)
		fmt.Printf("成交价: %.2f\n", auction.WinningPrice)
	}
	fmt.Println("==============================")
}

func PrintBid(bid *common.Bid, index int) {
	fmt.Printf("%d. 出价人: %s, 价格: %.2f, 时间: %s\n",
		index, bid.BidderID, bid.Price, bid.BidTime.Format("2006-01-02 15:04:05"))
}
