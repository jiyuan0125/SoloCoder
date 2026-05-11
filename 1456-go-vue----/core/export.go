package core

import (
	"archivesystem/common"
	"encoding/csv"
	"fmt"
	"io"
)

func maskTitle(title string) string {
	runes := []rune(title)
	if len(runes) <= 4 {
		return title
	}
	return string(runes[:4]) + "****"
}

func (s *ArchiveService) ExportArchives(w io.Writer, userRole common.UserRole) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"档案编号", "档案标题", "分类", "密级", "归档日期", "归档人"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, arch := range s.archives {
		if arch.IsDestroyed {
			continue
		}

		title := arch.Title
		if userRole == common.UserRoleEmployee &&
			(arch.SecrecyLevel == common.ArchiveSecrecyLevelConf ||
				arch.SecrecyLevel == common.ArchiveSecrecyLevelTopSecret) {
			title = maskTitle(title)
		}

		row := []string{
			arch.ArchiveID,
			title,
			string(arch.Category),
			string(arch.SecrecyLevel),
			arch.ArchiveDate,
			arch.Archiver,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (s *ArchiveService) ExportBorrowRecords(w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"记录ID", "档案编号", "申请人", "借阅事由", "申请日期",
		"预计归还日期", "实际归还日期", "状态", "部门经理审批", "管理员审批"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, br := range s.borrowRecords {
		row := []string{
			fmt.Sprintf("%d", br.ID),
			br.ArchiveID,
			br.Applicant,
			br.Reason,
			br.ApplyDate,
			br.ExpectedReturn,
			br.ActualReturn,
			string(br.Status),
			string(br.ManagerApproval),
			string(br.AdminApproval),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}
