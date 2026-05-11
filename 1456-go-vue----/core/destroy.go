package core

import (
	"archivesystem/common"
	"fmt"
)

func getRetentionYears(category common.ArchiveCategory) int {
	switch category {
	case common.ArchiveCategoryPersonnel:
		return -1
	case common.ArchiveCategoryFinance:
		return 30
	case common.ArchiveCategoryContract:
		return 10
	case common.ArchiveCategoryTechDoc:
		return 5
	}
	return 0
}

func (s *ArchiveService) ListPendingDestroy(req common.ListPendingDestroyRequest) []common.Archive {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currentDate := req.CurrentDate
	if currentDate == "" {
		currentDate = today()
	}
	now, err := parseDate(currentDate)
	if err != nil {
		return nil
	}

	result := make([]common.Archive, 0)
	for _, arch := range s.archives {
		if arch.IsDestroyed {
			continue
		}

		retentionYears := getRetentionYears(arch.Category)
		if retentionYears <= 0 {
			continue
		}

		archiveDate, err := parseDate(arch.ArchiveDate)
		if err != nil {
			continue
		}

		retentionEnd := archiveDate.AddDate(retentionYears, 0, 0)
		if now.After(retentionEnd) || now.Equal(retentionEnd) {
			result = append(result, *arch)
		}
	}
	return result
}

func (s *ArchiveService) DestroyArchive(req common.DestroyArchiveRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	arch, ok := s.archives[req.ArchiveID]
	if !ok || arch.IsDestroyed {
		return fmt.Errorf("档案不存在或已销毁")
	}

	for _, br := range s.borrowRecords {
		if br.ArchiveID == req.ArchiveID &&
			(br.Status == common.BorrowStatusBorrowed ||
				br.Status == common.BorrowStatusOverdue ||
				br.Status == common.BorrowStatusApproved) {
			return fmt.Errorf("档案借阅中，无法销毁")
		}
	}

	arch.IsDestroyed = true

	s.destroyID++
	record := &common.DestroyRecord{
		ID:          s.destroyID,
		ArchiveID:   req.ArchiveID,
		Title:       arch.Title,
		Destroyer:   req.Destroyer,
		DestroyDate: today(),
	}
	s.destroyRecords = append(s.destroyRecords, record)

	return nil
}

func (s *ArchiveService) ListDestroyRecords() []common.DestroyRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.DestroyRecord, 0, len(s.destroyRecords))
	for _, dr := range s.destroyRecords {
		result = append(result, *dr)
	}
	return result
}
