package core

import (
	"archivesystem/common"
	"fmt"
	"strings"
)

func (s *ArchiveService) CreateArchive(req common.CreateArchiveRequest) (string, error) {
	if req.Title == "" {
		return "", fmt.Errorf("档案标题不能为空")
	}
	if req.Category == "" {
		return "", fmt.Errorf("档案分类不能为空")
	}
	if req.SecrecyLevel == "" {
		return "", fmt.Errorf("密级不能为空")
	}
	if req.Archiver == "" {
		return "", fmt.Errorf("归档人不能为空")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	year := currentYear()
	if req.ArchiveDate != "" {
		if len(req.ArchiveDate) >= 4 {
			year = req.ArchiveDate[:4]
		}
	}

	count := s.idCounter[year] + 1
	s.idCounter[year] = count

	archiveID := fmt.Sprintf("%s%04d", year, count)

	if req.ArchiveDate == "" {
		req.ArchiveDate = today()
	}

	archive := &common.Archive{
		ArchiveID:    archiveID,
		Title:        req.Title,
		Category:     req.Category,
		SecrecyLevel: req.SecrecyLevel,
		ArchiveDate:  req.ArchiveDate,
		Archiver:     req.Archiver,
		IsDestroyed:  false,
	}

	s.archives[archiveID] = archive
	return archiveID, nil
}

func (s *ArchiveService) ListArchives(req common.ListArchivesRequest) []common.Archive {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.Archive, 0)
	for _, arch := range s.archives {
		if arch.IsDestroyed {
			continue
		}

		if req.Category != "" && arch.Category != req.Category {
			continue
		}

		if req.Keyword != "" {
			if !strings.Contains(arch.ArchiveID, req.Keyword) &&
				!strings.Contains(arch.Title, req.Keyword) {
				continue
			}
		}

		if req.UserRole == common.UserRoleEmployee &&
			(arch.SecrecyLevel == common.ArchiveSecrecyLevelConf ||
				arch.SecrecyLevel == common.ArchiveSecrecyLevelTopSecret) {
			result = append(result, common.Archive{
				ArchiveID:    arch.ArchiveID,
				Title:        arch.Title,
				Category:     "",
				SecrecyLevel: "",
				ArchiveDate:  "",
				Archiver:     "",
				IsDestroyed:  arch.IsDestroyed,
			})
			continue
		}

		result = append(result, *arch)
	}
	return result
}

func (s *ArchiveService) GetArchive(req common.GetArchiveRequest) (*common.Archive, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	arch, ok := s.archives[req.ArchiveID]
	if !ok || arch.IsDestroyed {
		return nil, fmt.Errorf("档案不存在")
	}

	if req.UserRole == common.UserRoleEmployee &&
		(arch.SecrecyLevel == common.ArchiveSecrecyLevelConf ||
			arch.SecrecyLevel == common.ArchiveSecrecyLevelTopSecret) {
		return &common.Archive{
			ArchiveID:    arch.ArchiveID,
			Title:        arch.Title,
			Category:     "",
			SecrecyLevel: "",
			ArchiveDate:  "",
			Archiver:     "",
			IsDestroyed:  arch.IsDestroyed,
		}, nil
	}

	return arch, nil
}
