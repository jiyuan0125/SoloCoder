package common

import (
	"errors"
	"strings"
)

var ValidHouseTypes = []string{
	"一室一厅",
	"两室一厅",
	"两室两厅",
	"三室一厅",
	"三室两厅",
	"四室及以上",
}

type FloorLevel string

const (
	FloorLow    FloorLevel = "low"
	FloorMiddle FloorLevel = "middle"
	FloorHigh   FloorLevel = "high"
)

var FloorLevelMap = map[FloorLevel]string{
	FloorLow:    "低层(1-3层)",
	FloorMiddle: "中层(4-10层)",
	FloorHigh:   "高层(11层以上)",
}

type PriceType string

const (
	PriceTypeRent  PriceType = "rent"
	PriceTypeSell  PriceType = "sell"
	PriceTypeNegotiable PriceType = "negotiable"
)

type HouseStatus string

const (
	StatusActive   HouseStatus = "active"
	StatusOffline  HouseStatus = "offline"
	StatusSold     HouseStatus = "sold"
)

type Property struct {
	ID           string      `json:"id"`
	LandlordID   string      `json:"landlord_id"`
	Community    string      `json:"community"`
	HouseType    string      `json:"house_type"`
	Area         float64     `json:"area"`
	Floor        int         `json:"floor"`
	Orientation  string      `json:"orientation"`
	PriceType    PriceType   `json:"price_type"`
	Price        float64     `json:"price"`
	Contact      string      `json:"contact"`
	Status       HouseStatus `json:"status"`
	CreatedAt    int64       `json:"created_at"`
	UpdatedAt    int64       `json:"updated_at"`
}

type Favorite struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	PropertyID  string `json:"property_id"`
	CreatedAt   int64  `json:"created_at"`
}

type FilterRequest struct {
	MinPrice    *float64   `json:"min_price"`
	MaxPrice    *float64   `json:"max_price"`
	MinArea     *float64   `json:"min_area"`
	MaxArea     *float64   `json:"max_area"`
	HouseTypes  []string   `json:"house_types"`
	FloorLevels []FloorLevel `json:"floor_levels"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func ValidateHouseType(houseType string) error {
	for _, t := range ValidHouseTypes {
		if t == houseType {
			return nil
		}
	}
	return errors.New("户型必须从预设列表中选择: " + strings.Join(ValidHouseTypes, ", "))
}

func GetFloorLevel(floor int) FloorLevel {
	if floor <= 3 {
		return FloorLow
	} else if floor <= 10 {
		return FloorMiddle
	} else {
		return FloorHigh
	}
}

func MatchFloorLevel(floor int, targetLevels []FloorLevel) bool {
	if len(targetLevels) == 0 {
		return true
	}
	propertyLevel := GetFloorLevel(floor)
	for _, level := range targetLevels {
		if level == propertyLevel {
			return true
		}
	}
	return false
}
