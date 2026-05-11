package core

import (
	"canteen/internal/common"
	"time"
)

func (s *CanteenService) CalculateRecommendedCalories(employee *common.Employee) float64 {
	switch employee.Gender {
	case common.GenderMale:
		return float64(employee.Weight) * 24
	case common.GenderFemale:
		return float64(employee.Weight) * 22
	default:
		return float64(employee.Weight) * 24
	}
}

func (s *CanteenService) GetNutritionSummary(employeeID string, days int) (*common.NutritionSummary, error) {
	s.employeesMu.RLock()
	employee, exists := s.employees[employeeID]
	s.employeesMu.RUnlock()

	if !exists {
		return nil, ErrEmployeeNotFound
	}

	records := s.GetConsumptionRecordsByDays(employeeID, days)

	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	var totalCalories, totalProtein, totalCarbs, totalFat float64

	for _, record := range records {
		for _, item := range record.Items {
			quantity := float64(item.Quantity)
			totalCalories += item.Dish.Nutrition.Calories * quantity
			totalProtein += item.Dish.Nutrition.Protein * quantity
			totalCarbs += item.Dish.Nutrition.Carbs * quantity
			totalFat += item.Dish.Nutrition.Fat * quantity
		}
	}

	recommended := s.CalculateRecommendedCalories(employee) * float64(days)

	return &common.NutritionSummary{
		StartDate:   startDate,
		EndDate:     now,
		Calories:    totalCalories,
		Protein:     totalProtein,
		Carbs:       totalCarbs,
		Fat:         totalFat,
		Recommended: recommended,
	}, nil
}
