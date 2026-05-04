package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"realestate/pkg/common"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}) (*common.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("响应解析失败: %v", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("%s", result.Message)
	}

	return &result, nil
}

type CreatePropertyRequest struct {
	LandlordID  string             `json:"landlord_id"`
	Community   string             `json:"community"`
	HouseType   string             `json:"house_type"`
	Area        float64            `json:"area"`
	Floor       int                `json:"floor"`
	Orientation string             `json:"orientation"`
	PriceType   common.PriceType   `json:"price_type"`
	Price       float64            `json:"price"`
	Contact     string             `json:"contact"`
}

func (c *Client) CreateProperty(req CreatePropertyRequest) (*common.Property, error) {
	resp, err := c.doRequest(http.MethodPost, "/properties", req)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var property common.Property
	if err := json.Unmarshal(data, &property); err != nil {
		return nil, err
	}

	return &property, nil
}

type UpdatePropertyRequest struct {
	Community   string           `json:"community,omitempty"`
	HouseType   string           `json:"house_type,omitempty"`
	Area        float64          `json:"area,omitempty"`
	Floor       int              `json:"floor,omitempty"`
	Orientation string           `json:"orientation,omitempty"`
	PriceType   common.PriceType `json:"price_type,omitempty"`
	Price       float64          `json:"price,omitempty"`
	Contact     string           `json:"contact,omitempty"`
}

func (c *Client) UpdateProperty(id string, req UpdatePropertyRequest) (*common.Property, error) {
	resp, err := c.doRequest(http.MethodPut, "/properties/"+id, req)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var property common.Property
	if err := json.Unmarshal(data, &property); err != nil {
		return nil, err
	}

	return &property, nil
}

func (c *Client) OfflineProperty(id string) error {
	_, err := c.doRequest(http.MethodPost, "/properties/"+id+"/offline", nil)
	return err
}

func (c *Client) SellProperty(id string) error {
	_, err := c.doRequest(http.MethodPost, "/properties/"+id+"/sold", nil)
	return err
}

func (c *Client) GetLandlordProperties(landlordID string) ([]*common.Property, error) {
	query := url.Values{}
	query.Set("landlord_id", landlordID)

	resp, err := c.doRequest(http.MethodGet, "/properties?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var properties []*common.Property
	if err := json.Unmarshal(data, &properties); err != nil {
		return nil, err
	}

	return properties, nil
}

func (c *Client) GetProperty(id string) (*common.Property, error) {
	resp, err := c.doRequest(http.MethodGet, "/properties/"+id, nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var property common.Property
	if err := json.Unmarshal(data, &property); err != nil {
		return nil, err
	}

	return &property, nil
}

func (c *Client) FilterProperties(filter common.FilterRequest) ([]*common.Property, error) {
	resp, err := c.doRequest(http.MethodPost, "/properties/filter", filter)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var properties []*common.Property
	if err := json.Unmarshal(data, &properties); err != nil {
		return nil, err
	}

	return properties, nil
}

func (c *Client) AddFavorite(userID, propertyID string) error {
	req := map[string]string{
		"user_id":     userID,
		"property_id": propertyID,
	}
	_, err := c.doRequest(http.MethodPost, "/favorites", req)
	return err
}

func (c *Client) RemoveFavorite(userID, propertyID string) error {
	query := url.Values{}
	query.Set("user_id", userID)
	query.Set("property_id", propertyID)

	_, err := c.doRequest(http.MethodDelete, "/favorites?"+query.Encode(), nil)
	return err
}

type FavoriteItem struct {
	FavoriteID     string           `json:"favorite_id"`
	PropertyID     string           `json:"property_id"`
	CreatedAt      int64            `json:"created_at"`
	PropertyRemoved bool           `json:"property_removed,omitempty"`
	Property       *common.Property `json:"property,omitempty"`
	Status         string           `json:"status,omitempty"`
	StatusText     string           `json:"status_text,omitempty"`
}

func (c *Client) GetFavorites(userID string) ([]*FavoriteItem, error) {
	query := url.Values{}
	query.Set("user_id", userID)

	resp, err := c.doRequest(http.MethodGet, "/favorites?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var items []*FavoriteItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Client) CheckFavorite(userID, propertyID string) (bool, error) {
	query := url.Values{}
	query.Set("user_id", userID)
	query.Set("property_id", propertyID)

	resp, err := c.doRequest(http.MethodGet, "/favorites/check?"+query.Encode(), nil)
	if err != nil {
		return false, err
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("响应格式错误")
	}

	isFavorited, ok := data["is_favorited"].(bool)
	if !ok {
		return false, fmt.Errorf("响应格式错误")
	}

	return isFavorited, nil
}
