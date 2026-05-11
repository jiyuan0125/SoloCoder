package common

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateEmployeeRequest struct {
	Name   string `json:"name"`
	Gender Gender `json:"gender"`
	Weight int    `json:"weight"`
}

type RechargeRequest struct {
	EmployeeID string `json:"employee_id"`
	Amount     int    `json:"amount"`
}

type CreateDishRequest struct {
	Name      string        `json:"name"`
	Price     int           `json:"price"`
	Nutrition NutritionInfo `json:"nutrition"`
}

type UpdateDishRequest struct {
	Name      *string        `json:"name,omitempty"`
	Price     *int           `json:"price,omitempty"`
	Nutrition *NutritionInfo `json:"nutrition,omitempty"`
}

type AddConsumptionItemRequest struct {
	DishID   string `json:"dish_id"`
	Quantity int    `json:"quantity"`
}

type CheckoutRequest struct {
	EmployeeID string       `json:"employee_id"`
	Items      []DishItem   `json:"items"`
	MealType   MealType     `json:"meal_type"`
}

type DishItem struct {
	DishID   string `json:"dish_id"`
	Quantity int    `json:"quantity"`
}

type NutritionSummaryRequest struct {
	EmployeeID string `json:"employee_id"`
	Days       int    `json:"days"`
}
