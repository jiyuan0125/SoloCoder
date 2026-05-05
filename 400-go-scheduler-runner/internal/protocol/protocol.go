package protocol

const (
	OpListTasks       = "list_tasks"
	OpGetTask         = "get_task"
	OpAddTask         = "add_task"
	OpDeleteTask      = "delete_task"
	OpListExecutions  = "list_executions"
	OpGetExecution    = "get_execution"
	OpPing            = "ping"
)

type Request struct {
	Operation string                 `json:"operation"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type Response struct {
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type ServerConfig struct {
	TCPAddr       string `json:"tcp_addr"`
	HTTPAddr      string `json:"http_addr"`
	ConfigFile    string `json:"config_file"`
	MaxRecords    int    `json:"max_records"`
	ShutdownWait  int    `json:"shutdown_wait"`
}
