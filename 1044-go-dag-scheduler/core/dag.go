package core

import (
	"fmt"
	"sort"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusSkipped   TaskStatus = "skipped"
)

type Task struct {
	ID             string
	Dependencies   []string
	Duration       int
	Retries        int
	RetryInterval  int
	Status         TaskStatus
	StartTime      int64
	EndTime        int64
	DurationMs     int64
	Attempts       int
	LastError      string
	ShouldFail     bool
}

type DAG struct {
	tasks    map[string]*Task
	adjList  map[string][]string
	inDegree map[string]int
}

func NewDAG() *DAG {
	return &DAG{
		tasks:    make(map[string]*Task),
		adjList:  make(map[string][]string),
		inDegree: make(map[string]int),
	}
}

func (d *DAG) AddTask(task *Task) error {
	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if _, exists := d.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}
	d.tasks[task.ID] = task
	d.adjList[task.ID] = []string{}
	d.inDegree[task.ID] = 0
	return nil
}

func (d *DAG) AddDependency(taskID, depID string) error {
	if _, exists := d.tasks[taskID]; !exists {
		return fmt.Errorf("task %s does not exist", taskID)
	}
	if _, exists := d.tasks[depID]; !exists {
		return fmt.Errorf("dependency task %s does not exist", depID)
	}

	for _, dep := range d.adjList[depID] {
		if dep == taskID {
			return fmt.Errorf("dependency already exists from %s to %s", depID, taskID)
		}
	}

	d.adjList[depID] = append(d.adjList[depID], taskID)
	d.inDegree[taskID]++
	return nil
}

func (d *DAG) Build(tasks []*Task) error {
	for _, task := range tasks {
		if err := d.AddTask(task); err != nil {
			return err
		}
	}

	for _, task := range tasks {
		for _, depID := range task.Dependencies {
			if err := d.AddDependency(task.ID, depID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DAG) DetectCycle() ([]string, bool) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(string) ([]string, bool)
	dfs = func(node string) ([]string, bool) {
		if recStack[node] {
			return append(path, node), true
		}
		if visited[node] {
			return nil, false
		}

		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range d.adjList[node] {
			if cycle, found := dfs(neighbor); found {
				return cycle, true
			}
		}

		recStack[node] = false
		path = path[:len(path)-1]
		return nil, false
	}

	for id := range d.tasks {
		if !visited[id] {
			if cycle, found := dfs(id); found {
				startIdx := -1
				for i, node := range cycle {
					if node == cycle[len(cycle)-1] {
						startIdx = i
						break
					}
				}
				if startIdx != -1 {
					return cycle[startIdx:], true
				}
				return cycle, true
			}
		}
	}

	return nil, false
}

func (d *DAG) TopologicalSort() ([]string, error) {
	if cycle, hasCycle := d.DetectCycle(); hasCycle {
		return nil, fmt.Errorf("cycle detected: %v", cycle)
	}

	inDegree := make(map[string]int)
	for k, v := range d.inDegree {
		inDegree[k] = v
	}

	queue := []string{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	sort.Strings(queue)

	result := []string{}
	for len(queue) > 0 {
		sort.Strings(queue)
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, neighbor := range d.adjList[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(d.tasks) {
		return nil, fmt.Errorf("graph contains a cycle")
	}

	return result, nil
}

func (d *DAG) GetTasks() map[string]*Task {
	return d.tasks
}

func (d *DAG) GetTask(id string) (*Task, bool) {
	task, exists := d.tasks[id]
	return task, exists
}

func (d *DAG) GetAdjList() map[string][]string {
	return d.adjList
}

func (d *DAG) GetInDegree() map[string]int {
	return d.inDegree
}

func (d *DAG) GetDependencies(taskID string) []string {
	task, exists := d.tasks[taskID]
	if !exists {
		return nil
	}
	return task.Dependencies
}
