package common

import (
	"encoding/json"
	"time"
)

type CommandType string

const (
	CmdDeploy   CommandType = "deploy"
	CmdRollback CommandType = "rollback"
	CmdStatus   CommandType = "status"
	CmdHistory  CommandType = "history"
)

type DeployConfig struct {
	BinaryPath       string `json:"binary_path"`
	BackupDir        string `json:"backup_dir"`
	HealthCheckURL   string `json:"health_check_url"`
	HealthCheckTimeout int   `json:"health_check_timeout"`
	BuildCommand     string `json:"build_command"`
	PullCommand      string `json:"pull_command"`
}

type DeployRequest struct {
	Command    CommandType   `json:"command"`
	Config     *DeployConfig `json:"config,omitempty"`
	BackupPath string        `json:"backup_path,omitempty"`
}

type StepStatus string

const (
	StepPending    StepStatus = "pending"
	StepRunning    StepStatus = "running"
	StepCompleted  StepStatus = "completed"
	StepFailed     StepStatus = "failed"
)

type StepRecord struct {
	Name     string        `json:"name"`
	Status   StepStatus    `json:"status"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Output    string       `json:"output"`
	Error     string       `json:"error"`
}

type DeployStatus string

const (
	StatusQueued    DeployStatus = "queued"
	StatusRunning   DeployStatus = "running"
	StatusSuccess   DeployStatus = "success"
	StatusFailed    DeployStatus = "failed"
	StatusRolledBack DeployStatus = "rolled_back"
)

type Deployment struct {
	ID            string       `json:"id"`
	Status        DeployStatus `json:"status"`
	StartTime     time.Time    `json:"start_time"`
	EndTime       time.Time    `json:"end_time"`
	TotalDuration time.Duration `json:"total_duration"`
	Steps         []*StepRecord `json:"steps"`
	Config        *DeployConfig `json:"config"`
	BackupPath    string       `json:"backup_path"`
	IsRollback    bool         `json:"is_rollback"`
}

type DeployResponse struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	DeployID  string        `json:"deploy_id,omitempty"`
	Status    DeployStatus  `json:"status,omitempty"`
	Deployments []*Deployment `json:"deployments,omitempty"`
}

func MarshalRequest(req *DeployRequest) ([]byte, error) {
	return json.Marshal(req)
}

func UnmarshalRequest(data []byte) (*DeployRequest, error) {
	var req DeployRequest
	err := json.Unmarshal(data, &req)
	return &req, err
}

func MarshalResponse(resp *DeployResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func UnmarshalResponse(data []byte) (*DeployResponse, error) {
	var resp DeployResponse
	err := json.Unmarshal(data, &resp)
	return &resp, err
}

func NewStepRecord(name string) *StepRecord {
	return &StepRecord{
		Name:   name,
		Status: StepPending,
	}
}

func (s *StepRecord) Start() {
	s.Status = StepRunning
	s.StartTime = time.Now()
}

func (s *StepRecord) Complete(output string) {
	s.Status = StepCompleted
	s.EndTime = time.Now()
	s.Duration = s.EndTime.Sub(s.StartTime)
	s.Output = output
}

func (s *StepRecord) Fail(output string, err string) {
	s.Status = StepFailed
	s.EndTime = time.Now()
	s.Duration = s.EndTime.Sub(s.StartTime)
	s.Output = output
	s.Error = err
}
