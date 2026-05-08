package api

type Task struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
}

type AddTaskRequest struct {
	Task Task `json:"task"`
}

type AddTaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type UpdateDependenciesRequest struct {
	TaskID       string   `json:"task_id"`
	Dependencies []string `json:"dependencies"`
}

type UpdateDependenciesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RemoveTaskRequest struct {
	TaskID string `json:"task_id"`
}

type RemoveTaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type SortResponse struct {
	Success           bool     `json:"success"`
	Order             []string `json:"order,omitempty"`
	MaxParallelism    int      `json:"max_parallelism,omitempty"`
	CriticalPath      []string `json:"critical_path,omitempty"`
	MinCompletionTime int      `json:"min_completion_time,omitempty"`
	Error             string   `json:"error,omitempty"`
	ErrorType         string   `json:"error_type,omitempty"`
	CyclePath         []string `json:"cycle_path,omitempty"`
}

type ListTasksResponse struct {
	Success bool   `json:"success"`
	Tasks   []Task `json:"tasks,omitempty"`
}

type ClearTasksResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type DotResponse struct {
	Success bool   `json:"success"`
	Dot     string `json:"dot,omitempty"`
	Error   string `json:"error,omitempty"`
}

type BatchAddTasksRequest struct {
	Tasks []Task `json:"tasks"`
}

type BatchAddTasksResponse struct {
	Success bool   `json:"success"`
	Added   int    `json:"added,omitempty"`
	Message string `json:"message,omitempty"`
}
