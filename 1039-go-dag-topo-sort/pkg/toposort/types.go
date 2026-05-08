package toposort

type Task struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
}

type Graph struct {
	tasks map[string]*Task
}

func NewGraph() *Graph {
	return &Graph{
		tasks: make(map[string]*Task),
	}
}

func (g *Graph) AddTask(task Task) error {
	if task.ID == "" {
		return ErrTaskIDEmpty
	}
	for _, dep := range task.Dependencies {
		if dep == task.ID {
			return ErrSelfDependency{TaskID: task.ID}
		}
	}
	g.tasks[task.ID] = &Task{
		ID:           task.ID,
		Name:         task.Name,
		Dependencies: append([]string(nil), task.Dependencies...),
	}
	return nil
}

func (g *Graph) RemoveTask(taskID string) {
	delete(g.tasks, taskID)
	for _, task := range g.tasks {
		newDeps := make([]string, 0, len(task.Dependencies))
		for _, dep := range task.Dependencies {
			if dep != taskID {
				newDeps = append(newDeps, dep)
			}
		}
		task.Dependencies = newDeps
	}
}

func (g *Graph) UpdateDependencies(taskID string, dependencies []string) error {
	task, exists := g.tasks[taskID]
	if !exists {
		return ErrTaskNotFound{TaskID: taskID}
	}
	for _, dep := range dependencies {
		if dep == taskID {
			return ErrSelfDependency{TaskID: taskID}
		}
	}
	task.Dependencies = append([]string(nil), dependencies...)
	return nil
}

func (g *Graph) GetTask(taskID string) (Task, bool) {
	task, exists := g.tasks[taskID]
	if !exists {
		return Task{}, false
	}
	return *task, true
}

func (g *Graph) GetAllTasks() []Task {
	tasks := make([]Task, 0, len(g.tasks))
	for _, task := range g.tasks {
		tasks = append(tasks, *task)
	}
	return tasks
}

func (g *Graph) TaskCount() int {
	return len(g.tasks)
}
