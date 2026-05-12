package service

import (
	"fmt"
	"time"

	"drg-system/internal/model"
	"drg-system/internal/repository"
)

type TodoService struct {
	repo repository.Repository
}

func NewTodoService(repo repository.Repository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) CreateSubmitTodo(hospitalID string, period string) (*model.Todo, error) {
	now := time.Now()
	todo := model.Todo{
		ID:          fmt.Sprintf("todo_%d", now.UnixNano()),
		Type:        model.TodoTypeSubmitSettlement,
		Title:       "提交结算数据",
		Description: fmt.Sprintf("请提交%s月份的DRG结算数据", period),
		Status:      model.TodoStatusPending,
		ReferenceID: period,
		HospitalID:  hospitalID,
	}
	return s.repo.CreateTodo(todo)
}

func (s *TodoService) CreateModifyTodo(hospitalID string, settlementID string, comment string) (*model.Todo, error) {
	now := time.Now()
	todo := model.Todo{
		ID:          fmt.Sprintf("todo_%d", now.UnixNano()),
		Type:        model.TodoTypeModifySettlement,
		Title:       "修改结算数据",
		Description: fmt.Sprintf("结算审核不通过，请修改后重新提交。审核意见：%s", comment),
		Status:      model.TodoStatusPending,
		ReferenceID: settlementID,
		HospitalID:  hospitalID,
	}
	return s.repo.CreateTodo(todo)
}

func (s *TodoService) ListTodos() []model.Todo {
	return s.repo.ListTodos()
}

func (s *TodoService) CompleteTodo(id string) (*model.Todo, error) {
	todo, err := s.repo.GetTodo(id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	todo.Status = model.TodoStatusCompleted
	todo.CompletedAt = &now
	return s.repo.UpdateTodo(id, *todo)
}

func (s *TodoService) CheckAndCreateMonthlyTodos() {
	today := time.Now()
	if today.Day() == 1 {
		period := fmt.Sprintf("%d-%02d", today.Year(), today.Month())
		s.CreateSubmitTodo("H001", period)
	}
}
