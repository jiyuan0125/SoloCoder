package service

import (
	"errors"
	"net/http"
	"time"

	"medical-quality-system/internal/middleware"
	"medical-quality-system/internal/model"
	"medical-quality-system/pkg/db"

	"gorm.io/gorm"
)

type TodoService struct{}

func NewTodoService() *TodoService {
	return &TodoService{}
}

func (s *TodoService) List(status model.TodoStatus, todoType model.TodoType) ([]model.Todo, error) {
	var todos []model.Todo
	query := db.DB.Preload("Indicator").Preload("PDCA")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if todoType != "" {
		query = query.Where("type = ?", todoType)
	}

	err := query.Order("created_at DESC").Find(&todos).Error
	return todos, err
}

func (s *TodoService) Get(id uint) (*model.Todo, error) {
	var todo model.Todo
	err := db.DB.Preload("Indicator").Preload("PDCA").First(&todo, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, middleware.NewAppError(http.StatusNotFound, "待办不存在")
	}
	return &todo, err
}

func (s *TodoService) Update(id uint, todo *model.Todo) error {
	existing, err := s.Get(id)
	if err != nil {
		return err
	}
	return db.DB.Model(existing).Updates(todo).Error
}

func (s *TodoService) UpdateStatus(id uint, status model.TodoStatus) error {
	todo, err := s.Get(id)
	if err != nil {
		return err
	}
	todo.Status = status
	return db.DB.Save(todo).Error
}

func (s *TodoService) Delete(id uint) error {
	_, err := s.Get(id)
	if err != nil {
		return err
	}
	return db.DB.Delete(&model.Todo{}, id).Error
}

func (s *TodoService) CheckOverdue() error {
	now := time.Now().Format("2006-01-02")
	var todos []model.Todo
	err := db.DB.Where("status IN ? AND due_date < ? AND deleted_at IS NULL",
		[]model.TodoStatus{model.TodoStatusPending, model.TodoStatusProgress},
		now,
	).Find(&todos).Error
	if err != nil {
		return err
	}

	for _, todo := range todos {
		todo.Status = model.TodoStatusOverdue
		db.DB.Save(&todo)
	}

	return nil
}
