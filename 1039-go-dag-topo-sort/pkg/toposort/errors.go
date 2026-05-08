package toposort

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTaskIDEmpty = errors.New("task ID cannot be empty")
)

type ErrSelfDependency struct {
	TaskID string
}

func (e ErrSelfDependency) Error() string {
	return fmt.Sprintf("task %q depends on itself (self-dependency)", e.TaskID)
}

func (e ErrSelfDependency) Is(target error) bool {
	_, ok := target.(ErrSelfDependency)
	return ok
}

type ErrTaskNotFound struct {
	TaskID string
}

func (e ErrTaskNotFound) Error() string {
	return fmt.Sprintf("task %q not found", e.TaskID)
}

func (e ErrTaskNotFound) Is(target error) bool {
	_, ok := target.(ErrTaskNotFound)
	return ok
}

type ErrCyclicDependency struct {
	Path []string
}

func (e ErrCyclicDependency) Error() string {
	if len(e.Path) == 0 {
		return "cyclic dependency detected"
	}
	return fmt.Sprintf("cyclic dependency detected: %s", strings.Join(e.Path, " -> "))
}

func (e ErrCyclicDependency) Is(target error) bool {
	_, ok := target.(ErrCyclicDependency)
	return ok
}

type ErrDependencyNotFound struct {
	TaskID       string
	DependencyID string
}

func (e ErrDependencyNotFound) Error() string {
	return fmt.Sprintf("task %q depends on non-existent task %q", e.TaskID, e.DependencyID)
}

func (e ErrDependencyNotFound) Is(target error) bool {
	_, ok := target.(ErrDependencyNotFound)
	return ok
}
