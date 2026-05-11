package core

import (
	"cleaning-service/internal/common"
	"fmt"
	"time"
)

func (s *Service) CreateQualityCheck(req *common.CreateQualityCheckRequest) (*common.QualityCheck, error) {
	task := s.store.GetTask(req.TaskID)
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	if req.FloorScore < 1 || req.FloorScore > 5 {
		return nil, fmt.Errorf("floor score must be between 1 and 5")
	}
	if req.DeskScore < 1 || req.DeskScore > 5 {
		return nil, fmt.Errorf("desk score must be between 1 and 5")
	}
	if req.TrashScore < 1 || req.TrashScore > 5 {
		return nil, fmt.Errorf("trash score must be between 1 and 5")
	}

	checkCount := 1
	if task.QualityCheckID != "" {
		prevQC := s.store.GetQualityCheck(task.QualityCheckID)
		if prevQC != nil {
			checkCount = prevQC.CheckCount + 1
		}
	}

	qc := &common.QualityCheck{
		ID:          s.store.nextID("QC"),
		TaskID:      req.TaskID,
		InspectorID: req.InspectorID,
		Inspector:   req.Inspector,
		FloorScore:  req.FloorScore,
		DeskScore:   req.DeskScore,
		TrashScore:  req.TrashScore,
		TotalScore:  req.FloorScore + req.DeskScore + req.TrashScore,
		Notes:       req.Notes,
		CheckedAt:   time.Now(),
		CheckCount:  checkCount,
	}

	s.store.SaveQualityCheck(qc)
	task.QualityCheckID = qc.ID

	if qc.TotalScore >= 15 {
		task.Status = common.TaskStatusCompleted
		s.store.SaveTask(task)
		return qc, nil
	}

	if checkCount >= 3 {
		task.Status = common.TaskStatusFailed
		s.store.SaveTask(task)
		s.createManagementTodo(task)
		return qc, nil
	}

	todo := &common.TodoItem{
		ID:          s.store.nextID("TD"),
		TaskID:      task.ID,
		CleanerID:   task.CleanerID,
		Type:        "rework",
		Description: fmt.Sprintf("质检不合格（总分%d分），请于24小时内返工", qc.TotalScore),
		DueDate:     time.Now().Add(24 * time.Hour),
		Completed:   false,
		CreatedAt:   time.Now(),
	}

	s.store.SaveTodo(todo)
	task.Status = common.TaskStatusRecheck
	s.store.SaveTask(task)

	return qc, nil
}

func (s *Service) createManagementTodo(task *common.Task) {
	todo := &common.TodoItem{
		ID:          s.store.nextID("TD"),
		TaskID:      task.ID,
		CleanerID:   task.CleanerID,
		Type:        "management",
		Description: fmt.Sprintf("任务%s三次质检不合格，请班组管理员介入处理", task.TaskNo),
		DueDate:     time.Now().Add(48 * time.Hour),
		Completed:   false,
		CreatedAt:   time.Now(),
	}

	s.store.SaveTodo(todo)
}

func (s *Service) ListQualityChecks() []*common.QualityCheck {
	return s.store.ListQualityChecks()
}

func (s *Service) ListTodos() []*common.TodoItem {
	return s.store.ListTodos()
}

func (s *Service) CompleteTodo(todoID string) (*common.TodoItem, error) {
	todo := s.store.GetTodo(todoID)
	if todo == nil {
		return nil, fmt.Errorf("todo not found")
	}

	todo.Completed = true
	s.store.SaveTodo(todo)

	return todo, nil
}
