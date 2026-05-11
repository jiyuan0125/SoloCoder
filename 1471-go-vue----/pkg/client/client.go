package client

import (
	"bus-station/internal/api"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, endpoint string, body interface{}) (*api.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp api.Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func (c *Client) CreateSchedule(req *api.CreateScheduleRequest) (*api.Response, error) {
	return c.doRequest("POST", "/api/schedules/create", req)
}

func (c *Client) GetSchedule(scheduleNo, date string) (*api.Response, error) {
	query := url.Values{}
	query.Set("schedule_no", scheduleNo)
	query.Set("date", date)
	return c.doRequest("GET", "/api/schedules/get?"+query.Encode(), nil)
}

func (c *Client) ListSchedules(date string) (*api.Response, error) {
	query := url.Values{}
	query.Set("date", date)
	return c.doRequest("GET", "/api/schedules/list?"+query.Encode(), nil)
}

func (c *Client) SearchSchedules(departure, arrival, date string) (*api.Response, error) {
	req := &api.SearchSchedulesRequest{
		Departure: departure,
		Arrival:   arrival,
		Date:      date,
	}
	return c.doRequest("POST", "/api/schedules/search", req)
}

func (c *Client) CancelSchedule(scheduleNo, date string) (*api.Response, error) {
	req := &api.CancelScheduleRequest{
		ScheduleNo: scheduleNo,
		Date:       date,
	}
	return c.doRequest("POST", "/api/schedules/cancel", req)
}

func (c *Client) GetSeats(scheduleNo, date string) (*api.Response, error) {
	query := url.Values{}
	query.Set("schedule_no", scheduleNo)
	query.Set("date", date)
	return c.doRequest("GET", "/api/schedules/seats?"+query.Encode(), nil)
}

func (c *Client) PurchaseTickets(scheduleNo, date, passengerName string, seatNos []int) (*api.Response, error) {
	req := &api.PurchaseTicketsRequest{
		ScheduleNo:    scheduleNo,
		Date:          date,
		SeatNos:       seatNos,
		PassengerName: passengerName,
	}
	return c.doRequest("POST", "/api/tickets/purchase", req)
}

func (c *Client) GetTicket(ticketNo string) (*api.Response, error) {
	query := url.Values{}
	query.Set("ticket_no", ticketNo)
	return c.doRequest("GET", "/api/tickets/get?"+query.Encode(), nil)
}

func (c *Client) CheckIn(ticketNo string) (*api.Response, error) {
	req := &api.CheckInRequest{
		TicketNo: ticketNo,
	}
	return c.doRequest("POST", "/api/checkin", req)
}

func (c *Client) RequestRefund(ticketNo string) (*api.Response, error) {
	req := &api.RefundRequest{
		TicketNo: ticketNo,
	}
	return c.doRequest("POST", "/api/refunds/request", req)
}

func (c *Client) ProcessRefund(requestID string, approved bool, reason string) (*api.Response, error) {
	req := &api.ProcessRefundRequest{
		RequestID: requestID,
		Approved:  approved,
		Reason:    reason,
	}
	return c.doRequest("POST", "/api/refunds/process", req)
}

func (c *Client) ListPendingRefunds() (*api.Response, error) {
	return c.doRequest("GET", "/api/refunds/pending", nil)
}

func HandleResponse(resp *api.Response, err error) error {
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}

	if resp.Code != 0 {
		return fmt.Errorf("错误: %s (代码: %d)", resp.Message, resp.Code)
	}

	return nil
}
