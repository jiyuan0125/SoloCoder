package services

import (
	"epidemic-management/database"
	"epidemic-management/models"
)

func CreateTodo(todo *models.Todo) error {
	return database.DB.Create(todo).Error
}

func ListTodos() ([]models.Todo, error) {
	var todos []models.Todo
	err := database.DB.Order("created_at desc").Find(&todos).Error
	return todos, err
}

func UpdateTodoStatus(id uint, status models.TodoStatus) error {
	var todo models.Todo
	if err := database.DB.First(&todo, id).Error; err != nil {
		return err
	}
	todo.Status = status
	return database.DB.Save(&todo).Error
}

func ListTodosByStatus(status models.TodoStatus) ([]models.Todo, error) {
	var todos []models.Todo
	err := database.DB.Where("status = ?", status).Order("created_at desc").Find(&todos).Error
	return todos, err
}
