package common

import "time"

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type MealType string

const (
	MealBreakfast MealType = "breakfast"
	MealLunch     MealType = "lunch"
	MealDinner    MealType = "dinner"
	MealSupper    MealType = "supper"
)

type Employee struct {
	ID       string
	Name     string
	Gender   Gender
	Weight   int
	Balance  int
	Credit   int
	IsActive bool
}

type NutritionInfo struct {
	Calories  float64
	Protein   float64
	Carbs     float64
	Fat       float64
}

type Dish struct {
	ID         string
	Name       string
	Price      int
	Nutrition  NutritionInfo
	IsOnSale   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type DishSnapshot struct {
	ID        string
	Name      string
	Price     int
	Nutrition NutritionInfo
}

type RechargeRecord struct {
	ID         string
	EmployeeID string
	Amount     int
	Before     int
	After      int
	CreatedAt  time.Time
}

type ConsumptionItem struct {
	Dish     DishSnapshot
	Quantity int
}

type ConsumptionRecord struct {
	ID           string
	EmployeeID   string
	EmployeeName string
	Items        []ConsumptionItem
	TotalAmount  int
	MealType     MealType
	IsCredit     bool
	CreatedAt    time.Time
}

type NutritionSummary struct {
	StartDate  time.Time
	EndDate    time.Time
	Calories   float64
	Protein    float64
	Carbs      float64
	Fat        float64
	Recommended float64
}
