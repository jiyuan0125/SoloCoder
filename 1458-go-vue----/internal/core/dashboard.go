package core

import (
	"merchant-mgmt-system/pkg/common"
	"time"
)

func (s *Service) GetDashboardStats() (*common.DashboardStats, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	monthlyNewCustomers := 0
	customers := s.storage.GetAllCustomers()
	for _, c := range customers {
		if !c.CreatedAt.Before(startOfMonth) {
			monthlyNewCustomers++
		}
	}

	monthlySignedContracts := 0
	contracts := s.storage.GetAllContracts()
	for _, c := range contracts {
		if !c.CreatedAt.Before(startOfMonth) {
			monthlySignedContracts++
		}
	}

	var signingRate float64
	if monthlyNewCustomers > 0 {
		signingRate = float64(monthlySignedContracts) / float64(monthlyNewCustomers) * 100
	}

	totalFollows := 0
	followMap := make(map[string]int)
	follows := s.storage.GetAllFollows()
	for _, f := range follows {
		followMap[f.CustomerID]++
		totalFollows++
	}

	var avgFollowCount float64
	if len(customers) > 0 {
		avgFollowCount = float64(totalFollows) / float64(len(customers))
	}

	return &common.DashboardStats{
		MonthlyNewCustomers:    monthlyNewCustomers,
		MonthlySignedContracts: monthlySignedContracts,
		SigningRate:            signingRate,
		AverageFollowCount:     avgFollowCount,
	}, nil
}
