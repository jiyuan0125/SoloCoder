package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type TaskStatus string

const (
	StatusPending TaskStatus = "PENDING"
	StatusReady   TaskStatus = "READY"
	StatusRunning TaskStatus = "RUNNING"
	StatusSuccess TaskStatus = "SUCCESS"
	StatusFailed  TaskStatus = "FAILED"
)

type StateHistoryEntry struct {
	Status    TaskStatus `json:"status"`
	Timestamp time.Time  `json:"timestamp"`
	Reason    string     `json:"reason,omitempty"`
}

type Task struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Dependencies []string           `json:"dependencies"`
	Status       TaskStatus         `json:"status"`
	History      []StateHistoryEntry `json:"history"`
	Reason       string             `json:"reason,omitempty"`
	mu           sync.Mutex
}

type Scheduler struct {
	tasks     map[string]*Task
	dependents map[string][]string
	mu        sync.RWMutex
	cond      *sync.Cond
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		tasks:      make(map[string]*Task),
		dependents: make(map[string][]string),
	}
	s.cond = sync.NewCond(&s.mu)
	go s.run()
	return s
}

func (s *Scheduler) updateStatus(task *Task, status TaskStatus, reason string) {
	task.mu.Lock()
	task.Status = status
	task.Reason = reason
	task.History = append(task.History, StateHistoryEntry{
		Status:    status,
		Timestamp: time.Now(),
		Reason:    reason,
	})
	task.mu.Unlock()
}

func (s *Scheduler) checkDependencies(taskID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return false
	}
	for _, depID := range task.Dependencies {
		dep, ok := s.tasks[depID]
		if !ok {
			return false
		}
		if dep.Status != StatusSuccess {
			return false
		}
	}
	return true
}

func (s *Scheduler) executeTask(task *Task) {
	s.updateStatus(task, StatusRunning, "")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		delay := 1 + rand.Intn(5)
		time.Sleep(time.Duration(delay) * time.Second)
		if rand.Intn(10) == 0 {
			done <- fmt.Errorf("模拟执行失败")
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			s.handleFailure(task.ID, err.Error())
		} else {
			s.handleSuccess(task.ID)
		}
	case <-ctx.Done():
		s.handleFailure(task.ID, "任务执行超时")
	}
}

func (s *Scheduler) handleSuccess(taskID string) {
	s.mu.Lock()
	task, ok := s.tasks[taskID]
	if !ok {
		s.mu.Unlock()
		return
	}
	s.updateStatus(task, StatusSuccess, "")
	dependents := make([]string, len(s.dependents[taskID]))
	copy(dependents, s.dependents[taskID])
	s.mu.Unlock()

	for _, depID := range dependents {
		s.mu.Lock()
		depTask, ok := s.tasks[depID]
		if ok && depTask.Status == StatusPending && s.checkDependencies(depID) {
			s.updateStatus(depTask, StatusReady, "")
		}
		s.mu.Unlock()
	}

	s.cond.Broadcast()
}

func (s *Scheduler) handleFailure(taskID string, reason string) {
	s.mu.Lock()
	task, ok := s.tasks[taskID]
	if !ok {
		s.mu.Unlock()
		return
	}
	s.updateStatus(task, StatusFailed, reason)
	affected := s.collectAllDependents(taskID)
	s.mu.Unlock()

	for _, affectedID := range affected {
		s.mu.Lock()
		if affectedTask, ok := s.tasks[affectedID]; ok {
			if affectedTask.Status == StatusPending || affectedTask.Status == StatusReady {
				s.updateStatus(affectedTask, StatusFailed, fmt.Sprintf("上游任务 %s 执行失败", taskID))
			}
		}
		s.mu.Unlock()
	}

	s.cond.Broadcast()
}

func (s *Scheduler) collectAllDependents(taskID string) []string {
	visited := make(map[string]bool)
	result := []string{}
	var dfs func(id string)
	dfs = func(id string) {
		for _, depID := range s.dependents[id] {
			if !visited[depID] {
				visited[depID] = true
				result = append(result, depID)
				dfs(depID)
			}
		}
	}
	dfs(taskID)
	return result
}

func (s *Scheduler) run() {
	for {
		s.mu.Lock()
		var readyTask *Task
		for _, task := range s.tasks {
			if task.Status == StatusReady {
				readyTask = task
				break
			}
		}
		if readyTask == nil {
			s.cond.Wait()
			s.mu.Unlock()
			continue
		}
		s.mu.Unlock()
		go s.executeTask(readyTask)
	}
}

func (s *Scheduler) detectCycle(newTaskID string, deps []string) ([]string, bool) {
	adj := make(map[string][]string)
	for id, task := range s.tasks {
		adj[id] = append([]string{}, task.Dependencies...)
	}
	adj[newTaskID] = append([]string{}, deps...)

	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(id string) ([]string, bool)
	dfs = func(id string) ([]string, bool) {
		if recStack[id] {
			idx := -1
			for i, p := range path {
				if p == id {
					idx = i
					break
				}
			}
			if idx != -1 {
				cycle := append(path[idx:], id)
				return cycle, true
			}
			return []string{id, id}, true
		}
		if visited[id] {
			return nil, false
		}
		visited[id] = true
		recStack[id] = true
		path = append(path, id)

		for _, dep := range adj[id] {
			if cycle, found := dfs(dep); found {
				return cycle, true
			}
		}

		path = path[:len(path)-1]
		recStack[id] = false
		return nil, false
	}

	for id := range adj {
		if !visited[id] {
			if cycle, found := dfs(id); found {
				return cycle, true
			}
		}
	}
	return nil, false
}

func (s *Scheduler) addTask(id, name string, deps []string) (*Task, int, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; exists {
		return nil, 409, fmt.Sprintf("任务 %s 已存在", id)
	}

	for _, depID := range deps {
		if _, exists := s.tasks[depID]; !exists {
			return nil, 400, fmt.Sprintf("依赖的任务 %s 不存在", depID)
		}
	}

	if cycle, hasCycle := s.detectCycle(id, deps); hasCycle {
		cycleStr := ""
		for i, node := range cycle {
			if i > 0 {
				cycleStr += "→"
			}
			cycleStr += node
		}
		return nil, 409, fmt.Sprintf("检测到循环依赖: %s", cycleStr)
	}

	task := &Task{
		ID:           id,
		Name:         name,
		Dependencies: append([]string{}, deps...),
		Status:       StatusPending,
		History: []StateHistoryEntry{
			{Status: StatusPending, Timestamp: time.Now()},
		},
	}

	s.tasks[id] = task
	for _, depID := range deps {
		s.dependents[depID] = append(s.dependents[depID], id)
	}

	if s.checkDependencies(id) {
		s.updateStatus(task, StatusReady, "")
		s.cond.Signal()
	}

	return task, 0, ""
}

func main() {
	rand.Seed(time.Now().UnixNano())
	scheduler := NewScheduler()
	app := fiber.New()

	type CreateTaskRequest struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		Dependencies []string `json:"dependencies"`
	}

	app.Post("/tasks", func(c *fiber.Ctx) error {
		var req CreateTaskRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "请求体解析失败"})
		}
		if req.ID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "任务 ID 不能为空"})
		}
		if req.Name == "" {
			req.Name = req.ID
		}

		task, code, msg := scheduler.addTask(req.ID, req.Name, req.Dependencies)
		if code != 0 {
			return c.Status(code).JSON(fiber.Map{"error": msg})
		}
		return c.Status(201).JSON(task)
	})

	app.Get("/tasks/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		scheduler.mu.RLock()
		task, ok := scheduler.tasks[id]
		scheduler.mu.RUnlock()
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "任务不存在"})
		}
		return c.JSON(task)
	})

	type DependencyEdge struct {
		From string `json:"from"`
		To   string `json:"to"`
	}

	type GraphResponse struct {
		Tasks []*Task          `json:"tasks"`
		Edges []DependencyEdge `json:"edges"`
	}

	app.Get("/graph", func(c *fiber.Ctx) error {
		scheduler.mu.RLock()
		defer scheduler.mu.RUnlock()

		tasks := make([]*Task, 0, len(scheduler.tasks))
		for _, task := range scheduler.tasks {
			tasks = append(tasks, task)
		}

		edges := []DependencyEdge{}
		for id, task := range scheduler.tasks {
			for _, depID := range task.Dependencies {
				edges = append(edges, DependencyEdge{
					From: depID,
					To:   id,
				})
			}
		}

		return c.JSON(GraphResponse{
			Tasks: tasks,
			Edges: edges,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app.Listen(":" + port)
}
