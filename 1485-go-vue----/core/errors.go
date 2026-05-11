package core

import "fmt"

type TaskNotFoundError struct {
	TaskID string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("task not found: %s", e.TaskID)
}

type TaskExistsError struct {
	TaskID string
}

func (e *TaskExistsError) Error() string {
	return fmt.Sprintf("task already exists: %s", e.TaskID)
}

type TaskAlreadyEndedError struct {
	TaskID string
}

func (e *TaskAlreadyEndedError) Error() string {
	return fmt.Sprintf("task already ended: %s", e.TaskID)
}

type InvalidCargoTypeError struct {
	CargoType string
}

func (e *InvalidCargoTypeError) Error() string {
	return fmt.Sprintf("invalid cargo type: %s", e.CargoType)
}
