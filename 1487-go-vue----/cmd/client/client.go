package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"moving-platform/pkg/common"
	"moving-platform/pkg/core"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("server error: %s", resp.Status)
		}
		return fmt.Errorf("%s", errResp.Error)
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Estimate(items core.ItemList, distanceKM float64) (int64, error) {
	req := common.EstimateRequest{
		Items:      items,
		DistanceKM: distanceKM,
	}

	var resp common.EstimateResponse
	if err := c.doRequest(http.MethodPost, "/api/estimate", req, &resp); err != nil {
		return 0, err
	}

	return resp.Estimate, nil
}

func (c *Client) CreateBooking(req common.CreateBookingRequest) (string, int64, error) {
	var resp common.CreateBookingResponse
	if err := c.doRequest(http.MethodPost, "/api/bookings", req, &resp); err != nil {
		return "", 0, err
	}

	return resp.BookingID, resp.Estimate, nil
}

func (c *Client) GetBooking(id string) (*core.Booking, error) {
	var resp common.GetBookingResponse
	if err := c.doRequest(http.MethodGet, "/api/bookings/"+url.PathEscape(id), nil, &resp); err != nil {
		return nil, err
	}

	return resp.Booking, nil
}

func (c *Client) ListBookings() ([]*core.Booking, error) {
	var resp common.ListBookingsResponse
	if err := c.doRequest(http.MethodGet, "/api/bookings", nil, &resp); err != nil {
		return nil, err
	}

	return resp.Bookings, nil
}

func (c *Client) CancelBooking(id string) (int64, error) {
	var resp common.CancelBookingResponse
	if err := c.doRequest(http.MethodDelete, "/api/bookings/"+url.PathEscape(id), nil, &resp); err != nil {
		return 0, err
	}

	return resp.CancellationFee, nil
}

func (c *Client) CompleteBooking(id string) error {
	path := fmt.Sprintf("/api/bookings/%s?action=complete", url.PathEscape(id))
	return c.doRequest(http.MethodPatch, path, nil, nil)
}

func (c *Client) SettleBooking(id string, finalItems *core.ItemList, finalDistanceKM float64) (int64, error) {
	req := common.SettleBookingRequest{
		FinalItems:       finalItems,
		FinalDistanceKM:  finalDistanceKM,
	}

	var resp common.SettleBookingResponse
	path := fmt.Sprintf("/api/bookings/%s?action=settle", url.PathEscape(id))
	if err := c.doRequest(http.MethodPatch, path, req, &resp); err != nil {
		return 0, err
	}

	return resp.FinalAmount, nil
}

func (c *Client) AddReview(id string, rating int, comment string) error {
	req := common.AddReviewRequest{
		Rating:  rating,
		Comment: comment,
	}

	path := fmt.Sprintf("/api/bookings/%s?action=review", url.PathEscape(id))
	return c.doRequest(http.MethodPatch, path, req, nil)
}

func (c *Client) CheckAvailability(date time.Time) ([]core.TimeSlot, error) {
	path := fmt.Sprintf("/api/availability?date=%s", url.QueryEscape(date.Format("2006-01-02")))
	var resp common.CheckAvailabilityResponse
	if err := c.doRequest(http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	return resp.AvailableSlots, nil
}
