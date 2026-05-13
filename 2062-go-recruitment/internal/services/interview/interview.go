package interview

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"recruitment/internal/models"
	"recruitment/internal/services/candidate"
	"recruitment/internal/store"
	"recruitment/pkg/utils"
)

type Service struct {
	store          *store.Store
	candidateSvc   *candidate.Service
	calendarDir    string
}

func NewService(s *store.Store, cs *candidate.Service, calendarDir string) *Service {
	if err := os.MkdirAll(calendarDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create calendar directory: %v\n", err)
	}
	return &Service{store: s, candidateSvc: cs, calendarDir: calendarDir}
}

type ScheduleRequest struct {
	CandidateID string
	Interviewer string
	StartTime   time.Time
	EndTime     time.Time
	IsRetry     bool
}

type ConflictInfo struct {
	Interviewer string
	StartTime   time.Time
	EndTime     time.Time
	Description string
	Source      string
}

type ParseError struct {
	LineNumber int
	LineContent string
	Message    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("第 %d 行解析错误 [%s]: %s", e.LineNumber, e.LineContent, e.Message)
}

func (s *Service) Schedule(req ScheduleRequest) (*models.Interview, error) {
	if req.CandidateID == "" {
		return nil, fmt.Errorf("候选人ID不能为空")
	}
	if req.Interviewer == "" {
		return nil, fmt.Errorf("面试官不能为空")
	}
	if req.StartTime.IsZero() {
		return nil, fmt.Errorf("开始时间不能为空")
	}
	if req.EndTime.IsZero() {
		return nil, fmt.Errorf("结束时间不能为空")
	}
	if req.StartTime.After(req.EndTime) || req.StartTime.Equal(req.EndTime) {
		return nil, fmt.Errorf("结束时间必须晚于开始时间")
	}

	cand, err := s.candidateSvc.GetByID(req.CandidateID)
	if err != nil {
		return nil, err
	}

	if cand.Status == models.CandidateStatusRejected ||
		cand.Status == models.CandidateStatusWithdrawn ||
		cand.Status == models.CandidateStatusOfferAbandoned {
		return nil, fmt.Errorf("候选人已淘汰，无法安排面试")
	}

	conflicts, err := s.checkConflicts(req.Interviewer, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		msg := fmt.Sprintf("面试官 %s 时间冲突:\n", req.Interviewer)
		for _, c := range conflicts {
			msg += fmt.Sprintf("  - %s 至 %s: %s (来源: %s)\n",
				c.StartTime.Format("2006-01-02 15:04"),
				c.EndTime.Format("2006-01-02 15:04"),
				c.Description, c.Source)
		}
		return nil, fmt.Errorf("%s", msg)
	}

	round, err := s.getNextRound(req.CandidateID)
	if err != nil {
		return nil, err
	}

	interview := &models.Interview{
		ID:          utils.GenerateID(),
		CandidateID: req.CandidateID,
		JobID:       cand.JobID,
		Interviewer: req.Interviewer,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      models.InterviewStatusScheduled,
		IsRetry:     req.IsRetry,
		Round:       round,
	}

	err = s.store.Transact(func(db *models.Database) error {
		db.Interviews = append(db.Interviews, *interview)
		for i := range db.Candidates {
			if db.Candidates[i].ID == req.CandidateID {
				db.Candidates[i].Status = models.CandidateStatusInterview
				break
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return interview, nil
}

func (s *Service) checkConflicts(interviewer string, startTime, endTime time.Time) ([]ConflictInfo, error) {
	var conflicts []ConflictInfo

	calendarEvents, err := s.loadCalendarEvents(interviewer)
	if err != nil {
		return nil, err
	}
	for _, e := range calendarEvents {
		if utils.TimeOverlaps(startTime, endTime, e.StartTime, e.EndTime) {
			conflicts = append(conflicts, ConflictInfo{
				Interviewer: interviewer,
				StartTime:   e.StartTime,
				EndTime:     e.EndTime,
				Description: e.Description,
				Source:      "日历文件",
			})
		}
	}

	scheduledInterviews, err := s.getScheduledInterviews(interviewer)
	if err != nil {
		return nil, err
	}
	for _, i := range scheduledInterviews {
		if utils.TimeOverlaps(startTime, endTime, i.StartTime, i.EndTime) {
			conflicts = append(conflicts, ConflictInfo{
				Interviewer: interviewer,
				StartTime:   i.StartTime,
				EndTime:     i.EndTime,
				Description: fmt.Sprintf("面试 - 候选人: %s (第%d轮)", i.CandidateID, i.Round),
				Source:      "已安排面试",
			})
		}
	}

	return conflicts, nil
}

func (s *Service) loadCalendarEvents(interviewer string) ([]models.CalendarEvent, error) {
	filename := strings.ReplaceAll(interviewer, " ", "_") + ".json"
	filepath := fmt.Sprintf("%s/%s", s.calendarDir, filename)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return []models.CalendarEvent{}, nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("读取日历文件失败: %s, 错误: %w", filename, err)
	}

	var events []models.CalendarEvent
	if err := json.Unmarshal(data, &events); err == nil {
		return events, nil
	}

	return s.parseTextCalendar(filepath)
}

func (s *Service) parseTextCalendar(filepath string) ([]models.CalendarEvent, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("打开日历文件失败: %w", err)
	}
	defer file.Close()

	var events []models.CalendarEvent
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 3 {
			return nil, &ParseError{
				LineNumber:  lineNum,
				LineContent: line,
				Message:     "格式错误，需要: 开始时间|结束时间|描述",
			}
		}

		startStr := strings.TrimSpace(parts[0])
		endStr := strings.TrimSpace(parts[1])
		desc := strings.TrimSpace(parts[2])

		startTime, err := time.Parse("2006-01-02 15:04", startStr)
		if err != nil {
			return nil, &ParseError{
				LineNumber:  lineNum,
				LineContent: line,
				Message:     fmt.Sprintf("开始时间格式错误: %v", err),
			}
		}

		endTime, err := time.Parse("2006-01-02 15:04", endStr)
		if err != nil {
			return nil, &ParseError{
				LineNumber:  lineNum,
				LineContent: line,
				Message:     fmt.Sprintf("结束时间格式错误: %v", err),
			}
		}

		events = append(events, models.CalendarEvent{
			StartTime:   startTime,
			EndTime:     endTime,
			Description: desc,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return events, nil
}

func (s *Service) getScheduledInterviews(interviewer string) ([]models.Interview, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	var result []models.Interview
	for _, i := range db.Interviews {
		if i.Interviewer == interviewer && i.Status == models.InterviewStatusScheduled {
			result = append(result, i)
		}
	}
	return result, nil
}

func (s *Service) getNextRound(candidateID string) (int, error) {
	db, err := s.store.Load()
	if err != nil {
		return 0, err
	}

	maxRound := 0
	for _, i := range db.Interviews {
		if i.CandidateID == candidateID && i.Round > maxRound {
			maxRound = i.Round
		}
	}
	return maxRound + 1, nil
}

type SubmitResultRequest struct {
	InterviewID string
	Result      models.InterviewResult
	Feedback    string
}

func (s *Service) SubmitResult(req SubmitResultRequest) error {
	if req.InterviewID == "" {
		return fmt.Errorf("面试ID不能为空")
	}
	if req.Result == "" {
		return fmt.Errorf("面试结果不能为空")
	}

	err := s.store.Transact(func(db *models.Database) error {
		var interview *models.Interview
		for i := range db.Interviews {
			if db.Interviews[i].ID == req.InterviewID {
				interview = &db.Interviews[i]
				break
			}
		}

		if interview == nil {
			return fmt.Errorf("面试不存在: %s", req.InterviewID)
		}

		if interview.Status == models.InterviewStatusCompleted {
			return fmt.Errorf("面试结果已提交")
		}

		interview.Status = models.InterviewStatusCompleted
		interview.Result = &req.Result
		interview.Feedback = req.Feedback
		submittedAt := time.Now()
		interview.SubmittedAt = &submittedAt

		if req.Result == models.InterviewResultFail {
			for i := range db.Candidates {
				if db.Candidates[i].ID == interview.CandidateID {
					db.Candidates[i].Status = models.CandidateStatusRejected
					db.Candidates[i].RejectionReason = fmt.Sprintf("第%d轮面试未通过", interview.Round)
					rejectedAt := time.Now()
					db.Candidates[i].RejectedAt = &rejectedAt
					break
				}
			}
		} else if req.Result == models.InterviewResultPending {
		} else if req.Result == models.InterviewResultPass {
			hasPendingRetry := false
			for i := range db.Interviews {
				if db.Interviews[i].CandidateID == interview.CandidateID &&
					db.Interviews[i].ID != interview.ID &&
					db.Interviews[i].Status == models.InterviewStatusScheduled &&
					db.Interviews[i].IsRetry {
					hasPendingRetry = true
					break
				}
			}
			if !hasPendingRetry {
				for i := range db.Candidates {
					if db.Candidates[i].ID == interview.CandidateID {
						db.Candidates[i].Status = models.CandidateStatusOffer
						break
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	interview, err := s.GetByID(req.InterviewID)
	if err != nil {
		return nil
	}
	return s.store.TriggerUpdateCheck(interview.JobID)
}

func (s *Service) GetByID(id string) (*models.Interview, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range db.Interviews {
		if db.Interviews[i].ID == id {
			return &db.Interviews[i], nil
		}
	}

	return nil, fmt.Errorf("面试不存在: %s", id)
}

func (s *Service) List() ([]models.Interview, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	return db.Interviews, nil
}

func (s *Service) ListByCandidate(candidateID string) ([]models.Interview, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	var result []models.Interview
	for _, i := range db.Interviews {
		if i.CandidateID == candidateID {
			result = append(result, i)
		}
	}
	return result, nil
}

func (s *Service) AllInterviewsPassed(candidateID string) (bool, error) {
	db, err := s.store.Load()
	if err != nil {
		return false, err
	}

	var interviews []models.Interview
	for _, i := range db.Interviews {
		if i.CandidateID == candidateID {
			interviews = append(interviews, i)
		}
	}

	if len(interviews) == 0 {
		return false, nil
	}

	for _, i := range interviews {
		if i.Status != models.InterviewStatusCompleted {
			return false, nil
		}
		if i.Result == nil || *i.Result != models.InterviewResultPass {
			return false, nil
		}
	}

	return true, nil
}
