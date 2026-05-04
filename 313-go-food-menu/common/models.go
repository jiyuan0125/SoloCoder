package common

import "time"

const (
	MaxNameLength        = 30
	MaxDescriptionLength = 200
)

var ValidCategories = []string{"凉菜", "热菜", "主食", "饮品", "甜品"}

type DishStatus string

const (
	StatusOnSale  DishStatus = "on_sale"
	StatusOffSale DishStatus = "off_sale"
)

type Dish struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Price       float64    `json:"price"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	Category    string     `json:"category"`
	Status      DishStatus `json:"status"`
	IsRecommend bool       `json:"is_recommend"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CategorySummary struct {
	Category     string `json:"category"`
	TotalCount   int    `json:"total_count"`
	OnSaleCount  int    `json:"on_sale_count"`
}

type RecommendInfo struct {
	Date       string   `json:"date"`
	DishIDs    []string `json:"dish_ids"`
	UpdatedAt  time.Time `json:"updated_at"`
}
