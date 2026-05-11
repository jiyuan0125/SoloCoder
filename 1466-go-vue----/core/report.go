package core

import (
	"errors"
	"fmt"
	"laboratory/common"
	"time"
)

func (s *Store) GenerateReport(sampleID string) (*common.ReportInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sample, ok := s.samples[sampleID]
	if !ok {
		return nil, errors.New("样品不存在")
	}

	if sample.Status != common.SampleStatusCompleted {
		return nil, errors.New("样品检测未完成，无法生成报告")
	}

	st, ok := s.sampleTests[sampleID]
	if !ok {
		return nil, errors.New("样品无检测记录")
	}

	now := time.Now()
	reportID := fmt.Sprintf("RPT%s%04d", now.Format("200601"), s.reportSeq)
	s.reportSeq++

	tests := make([]common.TestRecordInfo, 0)
	allPass := true
	for _, itemID := range st.ItemIDs {
		record := st.Records[itemID]
		tests = append(tests, common.TestRecordInfo{
			ItemID:         record.ItemID,
			ItemName:       record.ItemName,
			Status:         record.Status,
			NumericValue:   record.NumericValue,
			Unit:           record.Unit,
			JudgmentResult: record.JudgmentResult,
			IsRetest:       record.IsRetest,
			Operator:       record.Operator,
			TestTime:       record.TestTime,
		})
		if record.Status == common.TestResultFail {
			allPass = false
		}
	}

	conclusion := common.TestResultFail
	if allPass {
		conclusion = common.TestResultPass
	}

	report := &Report{
		ID:         reportID,
		SampleID:   sampleID,
		CreateTime: now,
		Status:     common.ReportStatusDraft,
		Conclusion: conclusion,
		Tests:      tests,
	}

	s.reports[reportID] = report

	return reportToInfo(report), nil
}

func (s *Store) IssueReport(req *common.IssueReportRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	report, ok := s.reports[req.ReportID]
	if !ok {
		return errors.New("报告不存在")
	}

	if report.Status != common.ReportStatusDraft {
		return errors.New("只有草稿状态的报告可以签发")
	}

	report.Status = common.ReportStatusIssued
	report.Issuer = req.Issuer
	report.IssueTime = time.Now()

	return nil
}

func (s *Store) VoidReport(req *common.VoidReportRequest) (*common.ReportInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	report, ok := s.reports[req.ReportID]
	if !ok {
		return nil, errors.New("报告不存在")
	}

	if report.Status != common.ReportStatusIssued {
		return nil, errors.New("只能作废已签发的报告")
	}

	report.Status = common.ReportStatusVoid

	newReportID := fmt.Sprintf("RPT%s%04d", time.Now().Format("200601"), s.reportSeq)
	s.reportSeq++

	newReport := &Report{
		ID:         newReportID,
		SampleID:   report.SampleID,
		CreateTime: time.Now(),
		Status:     common.ReportStatusDraft,
		Conclusion: report.Conclusion,
		Tests:      append([]common.TestRecordInfo(nil), report.Tests...),
	}

	s.reports[newReportID] = newReport

	return reportToInfo(newReport), nil
}

func (s *Store) GetReport(reportID string) (*common.ReportInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, ok := s.reports[reportID]
	if !ok {
		return nil, errors.New("报告不存在")
	}

	return reportToInfo(report), nil
}

func (s *Store) ListReports() []*common.ReportInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.ReportInfo, 0)
	for _, report := range s.reports {
		result = append(result, reportToInfo(report))
	}
	return result
}

func (s *Store) GetOverdueSamples() []*common.SampleInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.SampleInfo, 0)
	cutoff := time.Now().AddDate(0, 0, -7)

	for _, sample := range s.samples {
		if sample.Status == common.SampleStatusPending && sample.DeliveryDate.Before(cutoff) {
			result = append(result, sampleToInfo(sample))
		}
	}

	return result
}

func reportToInfo(r *Report) *common.ReportInfo {
	return &common.ReportInfo{
		ID:         r.ID,
		SampleID:   r.SampleID,
		CreateTime: r.CreateTime,
		Status:     r.Status,
		Conclusion: r.Conclusion,
		Issuer:     r.Issuer,
		IssueTime:  r.IssueTime,
		Tests:      append([]common.TestRecordInfo(nil), r.Tests...),
	}
}
