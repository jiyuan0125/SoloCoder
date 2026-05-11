package core

import (
	"archivesystem/common"
	"fmt"
)

func (s *ArchiveService) ApplyBorrow(req common.ApplyBorrowRequest) (int64, error) {
	if req.ArchiveID == "" || req.Applicant == "" || req.Reason == "" || req.ExpectedReturn == "" {
		return 0, fmt.Errorf("参数不完整")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	arch, ok := s.archives[req.ArchiveID]
	if !ok || arch.IsDestroyed {
		return 0, fmt.Errorf("档案不存在")
	}

	for _, br := range s.borrowRecords {
		if br.ArchiveID == req.ArchiveID &&
			(br.Status == common.BorrowStatusApproved ||
				br.Status == common.BorrowStatusBorrowed ||
				br.Status == common.BorrowStatusOverdue ||
				br.Status == common.BorrowStatusPending) {
			return 0, fmt.Errorf("该档案已被借阅或申请中")
		}
	}

	s.borrowID++
	record := &common.BorrowRecord{
		ID:              s.borrowID,
		ArchiveID:       req.ArchiveID,
		Applicant:       req.Applicant,
		Reason:          req.Reason,
		ExpectedReturn:  req.ExpectedReturn,
		Status:          common.BorrowStatusPending,
		ManagerApproval: common.ApprovalStatusPending,
		AdminApproval:   common.ApprovalStatusPending,
		ApplyDate:       today(),
	}

	if arch.SecrecyLevel == common.ArchiveSecrecyLevelPublic {
		record.Status = common.BorrowStatusApproved
		record.ManagerApproval = common.ApprovalStatusApproved
		record.AdminApproval = common.ApprovalStatusApproved
	}

	s.borrowRecords = append(s.borrowRecords, record)
	return record.ID, nil
}

func (s *ArchiveService) ApproveBorrow(req common.ApproveBorrowRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var record *common.BorrowRecord
	for _, br := range s.borrowRecords {
		if br.ID == req.BorrowID {
			record = br
			break
		}
	}

	if record == nil {
		return fmt.Errorf("借阅记录不存在")
	}

	if record.Status != common.BorrowStatusPending {
		return fmt.Errorf("该借阅申请已处理")
	}

	arch, ok := s.archives[record.ArchiveID]
	if !ok {
		return fmt.Errorf("档案不存在")
	}

	action := "拒绝"
	if req.IsApproved {
		action = "通过"
	}

	switch arch.SecrecyLevel {
	case common.ArchiveSecrecyLevelPublic:
		return fmt.Errorf("公开档案无需审批")
	case common.ArchiveSecrecyLevelInternal:
		if req.UserRole != common.UserRoleManager {
			return fmt.Errorf("需部门经理审批")
		}
		if req.IsApproved {
			record.ManagerApproval = common.ApprovalStatusApproved
			record.AdminApproval = common.ApprovalStatusApproved
			record.Status = common.BorrowStatusApproved
		} else {
			record.ManagerApproval = common.ApprovalStatusRejected
			record.Status = common.BorrowStatusRejected
		}
	case common.ArchiveSecrecyLevelConf, common.ArchiveSecrecyLevelTopSecret:
		if req.UserRole == common.UserRoleManager {
			if req.IsApproved {
				record.ManagerApproval = common.ApprovalStatusApproved
			} else {
				record.ManagerApproval = common.ApprovalStatusRejected
				record.Status = common.BorrowStatusRejected
			}
		} else if req.UserRole == common.UserRoleAdmin {
			if req.IsApproved {
				record.AdminApproval = common.ApprovalStatusApproved
			} else {
				record.AdminApproval = common.ApprovalStatusRejected
				record.Status = common.BorrowStatusRejected
			}
		} else {
			return fmt.Errorf("无审批权限")
		}

		if record.ManagerApproval == common.ApprovalStatusApproved &&
			record.AdminApproval == common.ApprovalStatusApproved {
			record.Status = common.BorrowStatusApproved
		}
	}

	_ = action
	return nil
}

func (s *ArchiveService) ReturnArchive(req common.ReturnArchiveRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var record *common.BorrowRecord
	for _, br := range s.borrowRecords {
		if br.ID == req.BorrowID {
			record = br
			break
		}
	}

	if record == nil {
		return fmt.Errorf("借阅记录不存在")
	}

	if record.Status != common.BorrowStatusBorrowed &&
		record.Status != common.BorrowStatusOverdue {
		return fmt.Errorf("该档案不在借阅中")
	}

	record.ActualReturn = today()
	record.Status = common.BorrowStatusReturned
	return nil
}

func (s *ArchiveService) ConfirmBorrow(borrowID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, br := range s.borrowRecords {
		if br.ID == borrowID {
			if br.Status != common.BorrowStatusApproved {
				return fmt.Errorf("借阅申请未通过")
			}
			br.Status = common.BorrowStatusBorrowed
			return nil
		}
	}
	return fmt.Errorf("借阅记录不存在")
}

func (s *ArchiveService) ListBorrowRecords() []common.BorrowRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.BorrowRecord, 0, len(s.borrowRecords))
	for _, br := range s.borrowRecords {
		result = append(result, *br)
	}
	return result
}

func (s *ArchiveService) GetOverdueReminders(currentDate string) []common.OverdueReminder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if currentDate == "" {
		currentDate = today()
	}
	now, err := parseDate(currentDate)
	if err != nil {
		return nil
	}

	result := make([]common.OverdueReminder, 0)
	for _, br := range s.borrowRecords {
		if br.Status != common.BorrowStatusBorrowed && br.Status != common.BorrowStatusOverdue {
			continue
		}

		expected, err := parseDate(br.ExpectedReturn)
		if err != nil {
			continue
		}

		if now.After(expected) {
			arch, ok := s.archives[br.ArchiveID]
			title := ""
			if ok {
				title = arch.Title
			}

			days := int(now.Sub(expected).Hours() / 24)
			result = append(result, common.OverdueReminder{
				BorrowID:       br.ID,
				ArchiveID:      br.ArchiveID,
				ArchiveTitle:   title,
				Applicant:      br.Applicant,
				ExpectedReturn: br.ExpectedReturn,
				OverdueDays:    days,
			})

			br.Status = common.BorrowStatusOverdue
		}
	}
	return result
}
