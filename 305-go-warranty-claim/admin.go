package main

type AdminService struct {
	storage *Storage
}

func NewAdminService(storage *Storage) *AdminService {
	return &AdminService{storage: storage}
}

func (s *AdminService) GetPendingClaims() []*WarrantyClaim {
	return s.storage.GetClaimsByStatus(StatusPending)
}

func (s *AdminService) GetAllClaims() []*WarrantyClaim {
	return s.storage.GetAllClaims()
}

func (s *AdminService) GetStatistics() *Statistics {
	allClaims := s.storage.GetAllClaims()
	
	stats := &Statistics{
		TotalClaims:    len(allClaims),
		ClaimsByStatus: make(map[string]int),
	}

	for _, claim := range allClaims {
		switch claim.Status {
		case StatusPending:
			stats.PendingClaims++
		case StatusApproved:
			stats.ApprovedClaims++
		case StatusRejected:
			stats.RejectedClaims++
		case StatusProcessing:
			stats.ProcessingClaims++
		}
		stats.ClaimsByStatus[string(claim.Status)]++
	}

	return stats
}
