package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "PENDING"
	StatusRunning    TaskStatus = "RUNNING"
	StatusSucceeded  TaskStatus = "SUCCEEDED"
	StatusFailed     TaskStatus = "FAILED"
	StatusCancelled  TaskStatus = "CANCELLED"
	StatusTimeout    TaskStatus = "TIMEOUT"
)

type ExecutionRecord struct {
	StartTime   time.Time  `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	Duration    int64      `json:"duration"`
	Status      TaskStatus `json:"status"`
	Error       string     `json:"error"`
	FailedCause string     `json:"failedCause"`
}

type Task struct {
	Name         string            `json:"name"`
	Dependencies []string          `json:"dependencies"`
	Timeout      int               `json:"timeout"`
	Status       TaskStatus        `json:"status"`
	LastRun      *ExecutionRecord  `json:"lastRun"`
	CreateTime   time.Time         `json:"createTime"`
	mu           sync.RWMutex
}

type TaskRequest struct {
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
	Timeout      int      `json:"timeout"`
}

type GraphResponse struct {
	Tasks []TaskNode `json:"tasks"`
	Edges []Edge     `json:"edges"`
}

type TaskNode struct {
	Name         string       `json:"name"`
	Dependencies []string     `json:"dependencies"`
	Timeout      int          `json:"timeout"`
	Status       TaskStatus   `json:"status"`
	LastRun      *ExecutionRecord `json:"lastRun"`
	CreateTime   time.Time    `json:"createTime"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Scheduler struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make(map[string]*Task),
	}
}

func (s *Scheduler) CreateTask(req TaskRequest) (*Task, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("task name is required")
	}

	for _, dep := range req.Dependencies {
		if dep == req.Name {
			return nil, fmt.Errorf("task cannot depend on itself")
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[req.Name]; exists {
		return nil, fmt.Errorf("task '%s' already exists", req.Name)
	}

	for _, dep := range req.Dependencies {
		if _, exists := s.tasks[dep]; !exists {
			return nil, fmt.Errorf("dependency task '%s' does not exist", dep)
		}
	}

	cycle := s.detectCycle(req.Name, req.Dependencies)
	if cycle != nil {
		cycleStr := ""
		for i, t := range cycle {
			if i > 0 {
				cycleStr += "→"
			}
			cycleStr += t
		}
		return nil, fmt.Errorf("cycle detected: %s", cycleStr)
	}

	if req.Timeout <= 0 {
		req.Timeout = 60
	}

	task := &Task{
		Name:         req.Name,
		Dependencies: append([]string{}, req.Dependencies...),
		Timeout:      req.Timeout,
		Status:       StatusPending,
		CreateTime:   time.Now(),
	}

	s.tasks[req.Name] = task
	return task, nil
}

func (s *Scheduler) detectCycle(newTaskName string, newDeps []string) []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	tempTasks := make(map[string][]string)
	for name, task := range s.tasks {
		tempTasks[name] = append([]string{}, task.Dependencies...)
	}
	tempTasks[newTaskName] = newDeps

	var dfs func(string) ([]string, bool)
	dfs = func(name string) ([]string, bool) {
		visited[name] = true
		recStack[name] = true
		path = append(path, name)

		for _, dep := range tempTasks[name] {
			if !visited[dep] {
				if cycle, found := dfs(dep); found {
					return cycle, true
				}
			} else if recStack[dep] {
				cycleStart := -1
				for i, t := range path {
					if t == dep {
						cycleStart = i
						break
					}
				}
				cycle := append(path[cycleStart:], dep)
				return cycle, true
			}
		}

		recStack[name] = false
		path = path[:len(path)-1]
		return nil, false
	}

	cycle, _ := dfs(newTaskName)
	if cycle != nil {
		return cycle
	}

	for name := range tempTasks {
		if !visited[name] {
			if cycle, _ := dfs(name); cycle != nil {
				return cycle
			}
		}
	}

	return nil
}

func (s *Scheduler) GetTask(name string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[name]
	return task, ok
}

func (s *Scheduler) GetAllTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *Scheduler) GetGraph() GraphResponse {
	tasks := s.GetAllTasks()
	nodes := make([]TaskNode, 0, len(tasks))
	edges := make([]Edge, 0)

	for _, task := range tasks {
		nodes = append(nodes, TaskNode{
			Name:         task.Name,
			Dependencies: append([]string{}, task.Dependencies...),
			Timeout:      task.Timeout,
			Status:       task.Status,
			LastRun:      task.LastRun,
			CreateTime:   task.CreateTime,
		})
		for _, dep := range task.Dependencies {
			edges = append(edges, Edge{
				From: dep,
				To:   task.Name,
			})
		}
	}

	return GraphResponse{Tasks: nodes, Edges: edges}
}

func (s *Scheduler) getDownstreamTasks(taskName string) []string {
	downstream := make([]string, 0)
	visited := make(map[string]bool)

	var dfs func(string)
	dfs = func(name string) {
		for _, t := range s.GetAllTasks() {
			for _, dep := range t.Dependencies {
				if dep == name && !visited[t.Name] {
					visited[t.Name] = true
					downstream = append(downstream, t.Name)
					dfs(t.Name)
				}
			}
		}
	}

	dfs(taskName)
	return downstream
}

func (s *Scheduler) markTaskFailed(taskName string, errorMsg string, failedCause string) {
	if task, ok := s.GetTask(taskName); ok {
		task.mu.Lock()
		task.Status = StatusFailed
		if task.LastRun != nil {
			now := time.Now()
			task.LastRun.EndTime = &now
			task.LastRun.Duration = now.Sub(task.LastRun.StartTime).Milliseconds()
			task.LastRun.Status = StatusFailed
			task.LastRun.Error = errorMsg
			task.LastRun.FailedCause = failedCause
		}
		task.mu.Unlock()
	}
}

func (s *Scheduler) runTask(task *Task) {
	task.mu.Lock()
	if task.Status == StatusRunning {
		task.mu.Unlock()
		return
	}
	task.Status = StatusRunning
	task.LastRun = &ExecutionRecord{
		StartTime: time.Now(),
		Status:    StatusRunning,
	}
	task.mu.Unlock()

	done := make(chan struct{}, 1)
	var execError error

	go func() {
		defer close(done)
		execError = s.executeTaskLogic(task)
	}()

	select {
	case <-done:
		task.mu.Lock()
		now := time.Now()
		task.LastRun.EndTime = &now
		task.LastRun.Duration = now.Sub(task.LastRun.StartTime).Milliseconds()
		if execError != nil {
			task.Status = StatusFailed
			task.LastRun.Status = StatusFailed
			task.LastRun.Error = execError.Error()
			task.LastRun.FailedCause = fmt.Sprintf("任务 '%s' 执行失败", task.Name)
			task.mu.Unlock()
			s.propagateFailure(task.Name, fmt.Sprintf("依赖任务 %s 失败", task.Name))
		} else {
			task.Status = StatusSucceeded
			task.LastRun.Status = StatusSucceeded
			task.mu.Unlock()
			s.triggerDependents(task.Name)
		}
	case <-time.After(time.Duration(task.Timeout) * time.Second):
		task.mu.Lock()
		now := time.Now()
		task.LastRun.EndTime = &now
		task.LastRun.Duration = now.Sub(task.LastRun.StartTime).Milliseconds()
		task.Status = StatusTimeout
		task.LastRun.Status = StatusTimeout
		task.LastRun.Error = fmt.Sprintf("任务超时（%d 秒）", task.Timeout)
		task.LastRun.FailedCause = fmt.Sprintf("任务 '%s' 超时", task.Name)
		task.mu.Unlock()
		s.propagateFailure(task.Name, fmt.Sprintf("依赖任务 %s 超时", task.Name))
	}
}

func (s *Scheduler) executeTaskLogic(task *Task) error {
	time.Sleep(100 * time.Millisecond)
	
	if len(task.Name) >= 4 && task.Name[len(task.Name)-4:] == "FAIL" {
		return fmt.Errorf("task execution failed as requested")
	}
	
	return nil
}

func (s *Scheduler) triggerDependents(taskName string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.tasks {
		if t.Status == StatusPending {
			allDepsSucceeded := true
			for _, dep := range t.Dependencies {
				if depTask, ok := s.tasks[dep]; ok {
					if depTask.Status != StatusSucceeded {
						allDepsSucceeded = false
						break
					}
				}
			}
			if allDepsSucceeded && len(t.Dependencies) > 0 {
				go s.runTask(t)
			}
		}
	}
}

func (s *Scheduler) propagateFailure(failedTaskName string, cause string) {
	downstream := s.getDownstreamTasks(failedTaskName)
	for _, name := range downstream {
		if task, ok := s.GetTask(name); ok {
			task.mu.Lock()
			if task.Status == StatusRunning || task.Status == StatusPending {
				if task.LastRun == nil {
					now := time.Now()
					task.LastRun = &ExecutionRecord{
						StartTime: now,
					}
				}
				now := time.Now()
				task.LastRun.EndTime = &now
				task.LastRun.Duration = now.Sub(task.LastRun.StartTime).Milliseconds()
				task.Status = StatusFailed
				task.LastRun.Status = StatusFailed
				task.LastRun.Error = "级联失败"
				task.LastRun.FailedCause = cause
			}
			task.mu.Unlock()
		}
	}
}

func (s *Scheduler) StartTask(taskName string) error {
	task, ok := s.GetTask(taskName)
	if !ok {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	if len(task.Dependencies) == 0 {
		go s.runTask(task)
		return nil
	}

	for _, dep := range task.Dependencies {
		depTask, ok := s.GetTask(dep)
		if !ok || depTask.Status != StatusSucceeded {
			return fmt.Errorf("dependency task '%s' is not completed successfully", dep)
		}
	}

	go s.runTask(task)
	return nil
}

func main() {
	scheduler := NewScheduler()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(cors.New())

	app.Post("/tasks", func(c *fiber.Ctx) error {
		var req TaskRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		task, err := scheduler.CreateTask(req)
		if err != nil {
			errMsg := err.Error()
			if len(errMsg) >= 14 && errMsg[:14] == "cycle detected" {
				cyclePath := errMsg[16:]
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Cycle detected in task dependencies",
					"cycle": cyclePath,
				})
			}
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(task)
	})

	app.Get("/tasks", func(c *fiber.Ctx) error {
		return c.JSON(scheduler.GetAllTasks())
	})

	app.Get("/tasks/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		task, ok := scheduler.GetTask(name)
		if !ok {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("task '%s' not found", name))
		}
		return c.JSON(task)
	})

	app.Post("/tasks/:name/start", func(c *fiber.Ctx) error {
		name := c.Params("name")
		err := scheduler.StartTask(name)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("task '%s' started", name),
		})
	})

	app.Get("/graph", func(c *fiber.Ctx) error {
		return c.JSON(scheduler.GetGraph())
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Event Scheduler starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
