package core

import (
	"errors"
	"fmt"
	"laboratory/common"
	"sync"
	"time"
)

var claimMu sync.Mutex

func (s *Store) CreateTestItem(req *common.CreateTestItemRequest) (*common.TestItemInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	itemID := fmt.Sprintf("ITEM%d", len(s.testItems)+1)
	item := &TestItem{
		ID:            itemID,
		MethodID:      req.MethodID,
		Name:          req.Name,
		Steps:         req.Steps,
		JudgmentStd:   req.JudgmentStd,
		Type:          req.Type,
		Unit:          req.Unit,
		PassThreshold: req.PassThreshold,
		AllowRetest:   req.AllowRetest,
	}

	s.testItems[itemID] = item

	return &common.TestItemInfo{
		ID:            item.ID,
		MethodID:      item.MethodID,
		Name:          item.Name,
		Steps:         item.Steps,
		JudgmentStd:   item.JudgmentStd,
		Type:          item.Type,
		Unit:          item.Unit,
		PassThreshold: item.PassThreshold,
		AllowRetest:   item.AllowRetest,
	}, nil
}

func (s *Store) ListTestItems() []*common.TestItemInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.TestItemInfo, 0)
	for _, item := range s.testItems {
		result = append(result, &common.TestItemInfo{
			ID:            item.ID,
			MethodID:      item.MethodID,
			Name:          item.Name,
			Steps:         item.Steps,
			JudgmentStd:   item.JudgmentStd,
			Type:          item.Type,
			Unit:          item.Unit,
			PassThreshold: item.PassThreshold,
			AllowRetest:   item.AllowRetest,
		})
	}
	return result
}

func (s *Store) AssignTestItems(req *common.AssignTestItemsRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.samples[req.SampleID]; !ok {
		return errors.New("样品不存在")
	}

	for _, itemID := range req.ItemIDs {
		if _, ok := s.testItems[itemID]; !ok {
			return fmt.Errorf("检测项目 %s 不存在", itemID)
		}
	}

	st, ok := s.sampleTests[req.SampleID]
	if !ok {
		st = &SampleTests{
			SampleID: req.SampleID,
			ItemIDs:  []string{},
			Records:  make(map[string]*TestRecord),
		}
		s.sampleTests[req.SampleID] = st
	}

	for _, itemID := range req.ItemIDs {
		found := false
		for _, existing := range st.ItemIDs {
			if existing == itemID {
				found = true
				break
			}
		}
		if !found {
			st.ItemIDs = append(st.ItemIDs, itemID)
			st.Records[itemID] = &TestRecord{
				ItemID:   itemID,
				ItemName: s.testItems[itemID].Name,
				Status:   common.TestResultPending,
				IsRetest: false,
			}
		}
	}

	return nil
}

func (s *Store) ClaimSample(req *common.ClaimSampleRequest) error {
	claimMu.Lock()
	defer claimMu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	sample, ok := s.samples[req.SampleID]
	if !ok {
		return errors.New("样品不存在")
	}

	if sample.Status != common.SampleStatusPending {
		return errors.New("样品不在待检测状态，无法领取")
	}

	st, ok := s.sampleTests[req.SampleID]
	if !ok || len(st.ItemIDs) == 0 {
		return errors.New("样品未分配检测项目")
	}

	sample.Status = common.SampleStatusTesting
	sample.TesterName = req.TesterName
	sample.StatusLogs = append(sample.StatusLogs, common.StatusLogEntry{
		Time:     time.Now(),
		Status:   common.SampleStatusTesting,
		Operator: req.TesterName,
		Remark:   "领取样品开始检测",
	})

	return nil
}

func (s *Store) RecordTestResult(req *common.RecordTestResultRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sample, ok := s.samples[req.SampleID]
	if !ok {
		return errors.New("样品不存在")
	}

	if sample.Status != common.SampleStatusTesting {
		return errors.New("样品不在检测中状态")
	}

	st, ok := s.sampleTests[req.SampleID]
	if !ok {
		return errors.New("样品无检测记录")
	}

	item, ok := s.testItems[req.ItemID]
	if !ok {
		return errors.New("检测项目不存在")
	}

	record, ok := st.Records[req.ItemID]
	if !ok {
		return errors.New("该样品未分配此检测项目")
	}

	if req.IsRetest {
		if !item.AllowRetest {
			return errors.New("该检测项目不允许复检")
		}
		approved := false
		for _, r := range s.retestReqs {
			if r.SampleID == req.SampleID && r.ItemID == req.ItemID && r.Status == common.RetestStatusApproved {
				approved = true
				break
			}
		}
		if !approved {
			return errors.New("复检未获得审批")
		}
	} else {
		if record.Status != common.TestResultPending {
			return errors.New("该项目已检测，如需重新检测请申请复检")
		}
	}

	record.IsRetest = req.IsRetest
	record.Operator = req.Operator
	record.TestTime = time.Now()
	record.Unit = req.Unit

	if item.Type == common.TestTypeNumeric {
		record.NumericValue = req.NumericValue
		record.Unit = item.Unit
		if req.NumericValue >= item.PassThreshold {
			record.Status = common.TestResultPass
			record.JudgmentResult = common.TestResultPass
		} else {
			record.Status = common.TestResultFail
			record.JudgmentResult = common.TestResultFail
		}
	} else {
		record.JudgmentResult = req.JudgmentResult
		record.Status = req.JudgmentResult
	}

	allDone := true
	hasFail := false
	for _, itemID := range st.ItemIDs {
		r := st.Records[itemID]
		if r.Status == common.TestResultPending {
			allDone = false
			break
		}
		if r.Status == common.TestResultFail {
			hasFail = true
		}
	}

	if allDone {
		sample.Status = common.SampleStatusCompleted
		sample.StatusLogs = append(sample.StatusLogs, common.StatusLogEntry{
			Time:     time.Now(),
			Status:   common.SampleStatusCompleted,
			Operator: req.Operator,
			Remark:   "检测完成",
		})

		if hasFail {
		}
	}

	return nil
}

func (s *Store) GetSampleTests(sampleID string) ([]*common.TestRecordInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st, ok := s.sampleTests[sampleID]
	if !ok {
		return nil, errors.New("样品无检测记录")
	}

	result := make([]*common.TestRecordInfo, 0)
	for _, itemID := range st.ItemIDs {
		record := st.Records[itemID]
		result = append(result, &common.TestRecordInfo{
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
	}
	return result, nil
}

func (s *Store) ApplyRetest(req *common.ApplyRetestRequest) (*common.RetestRequestInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.testItems[req.ItemID]
	if !ok {
		return nil, errors.New("检测项目不存在")
	}

	if !item.AllowRetest {
		return nil, errors.New("该检测项目不允许复检")
	}

	requestID := fmt.Sprintf("RET%06d", s.retestSeq)
	s.retestSeq++

	r := &RetestRequest{
		ID:        requestID,
		SampleID:  req.SampleID,
		ItemID:    req.ItemID,
		Applicant: req.Applicant,
		Reason:    req.Reason,
		Status:    common.RetestStatusPending,
		ApplyTime: time.Now(),
	}

	s.retestReqs[requestID] = r

	return &common.RetestRequestInfo{
		ID:        r.ID,
		SampleID:  r.SampleID,
		ItemID:    r.ItemID,
		Applicant: r.Applicant,
		Reason:    r.Reason,
		Status:    r.Status,
		ApplyTime: r.ApplyTime,
	}, nil
}

func (s *Store) ApproveRetest(req *common.ApproveRetestRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.retestReqs[req.RequestID]
	if !ok {
		return errors.New("复检申请不存在")
	}

	if r.Status != common.RetestStatusPending {
		return errors.New("复检申请已处理")
	}

	if req.Approved {
		r.Status = common.RetestStatusApproved
		r.Approver = req.Approver
		r.Remark = req.Remark

		st, ok := s.sampleTests[r.SampleID]
		if ok {
			if record, ok := st.Records[r.ItemID]; ok {
				record.Status = common.TestResultPending
			}
		}

		sample, ok := s.samples[r.SampleID]
		if ok && sample.Status == common.SampleStatusCompleted {
			sample.Status = common.SampleStatusTesting
			sample.StatusLogs = append(sample.StatusLogs, common.StatusLogEntry{
				Time:     time.Now(),
				Status:   common.SampleStatusTesting,
				Operator: req.Approver,
				Remark:   "复检批准，重新进入检测",
			})
		}
	} else {
		r.Status = common.RetestStatusRejected
		r.Approver = req.Approver
		r.Remark = req.Remark
	}

	return nil
}

func (s *Store) ListRetestRequests() []*common.RetestRequestInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.RetestRequestInfo, 0)
	for _, r := range s.retestReqs {
		result = append(result, &common.RetestRequestInfo{
			ID:        r.ID,
			SampleID:  r.SampleID,
			ItemID:    r.ItemID,
			Applicant: r.Applicant,
			Reason:    r.Reason,
			Status:    r.Status,
			ApplyTime: r.ApplyTime,
			Approver:  r.Approver,
			Remark:    r.Remark,
		})
	}
	return result
}
