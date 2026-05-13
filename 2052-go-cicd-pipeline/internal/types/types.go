package types

import (
	"time"
)

type TriggerType string

const (
	TriggerManual   TriggerType = "manual"
	TriggerGitPush  TriggerType = "git_push"
	TriggerSchedule TriggerType = "schedule"
)

type FailureStrategy string

const (
	FailureStop    FailureStrategy = "stop"
	FailureContinue FailureStrategy = "continue"
)

type TaskMode string

const (
	TaskModeParallel TaskMode = "parallel"
	TaskModeSerial   TaskMode = "serial"
)

type PhaseType string

const (
	PhaseBuild  PhaseType = "build"
	PhaseTest   PhaseType = "test"
	PhaseDeploy PhaseType = "deploy"
)

type ExecutionStatus string

const (
	StatusPending    ExecutionStatus = "pending"
	StatusRunning    ExecutionStatus = "running"
	StatusSuccess    ExecutionStatus = "success"
	StatusFailed     ExecutionStatus = "failed"
	StatusTimeout    ExecutionStatus = "timeout"
	StatusCancelled  ExecutionStatus = "cancelled"
	StatusSkipped    ExecutionStatus = "skipped"
)

type TargetEnv string

const (
	EnvDev        TargetEnv = "dev"
	EnvStaging    TargetEnv = "staging"
	EnvProduction TargetEnv = "production"
)

type Pipeline struct {
	ID            string
	Name          string
	Description   string
	Phases        []Phase
	TotalAmount   float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Triggers      []TriggerType
}

type Phase struct {
	ID           string
	PipelineID   string
	Type         PhaseType
	Name         string
	Tasks        []Task
	TaskMode     TaskMode
	PlannedAmount float64
	OrderIndex   int
}

type Task struct {
	ID               string
	PhaseID          string
	Name             string
	Script           string
	TimeoutSec       int
	FailureStrategy  FailureStrategy
	OrderIndex       int
}

type Execution struct {
	ID            string
	PipelineID    string
	TriggerType   TriggerType
	Status        ExecutionStatus
	StartedAt     *time.Time
	FinishedAt    *time.Time
	TargetEnv     TargetEnv
	TargetPhase   PhaseType
	CreatedAt     time.Time
	IsRollback    bool
	RollbackFrom  string
	PhaseResults  []PhaseResult
}

type PhaseResult struct {
	ID            string
	ExecutionID   string
	PhaseType     PhaseType
	PhaseName     string
	Status        ExecutionStatus
	StartedAt     *time.Time
	FinishedAt    *time.Time
	TaskResults   []TaskResult
}

type TaskResult struct {
	ID            string
	PhaseResultID string
	TaskID        string
	TaskName      string
	Status        ExecutionStatus
	StartedAt     *time.Time
	FinishedAt    *time.Time
	LogPath       string
}

type Artifact struct {
	ID          string
	ExecutionID string
	TaskResultID string
	Name        string
	FilePath    string
	Size        int64
	CreatedAt   time.Time
}

type CleanupRecord struct {
	ID          string
	ExecutionID string
	Action      string
	Status      ExecutionStatus
	Result      string
	CreatedAt   time.Time
	FinishedAt  *time.Time
}
