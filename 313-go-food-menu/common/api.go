package common

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateDishRequest struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

type UpdateDishRequest struct {
	Name        string  `json:"name,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Description string  `json:"description,omitempty"`
	ImageURL    string  `json:"image_url,omitempty"`
	Category    string  `json:"category,omitempty"`
}

type SetRecommendRequest struct {
	DishIDs []string `json:"dish_ids"`
}

type SearchRequest struct {
	Category string `json:"category"`
	Keyword  string `json:"keyword"`
}

type DishListResponse struct {
	Dishes []Dish `json:"dishes"`
	Total  int    `json:"total"`
}

type CategorySummaryResponse struct {
	Summaries []CategorySummary `json:"summaries"`
}

type RecommendResponse struct {
	RecommendInfo RecommendInfo `json:"recommend_info"`
	Dishes        []Dish        `json:"dishes"`
}
