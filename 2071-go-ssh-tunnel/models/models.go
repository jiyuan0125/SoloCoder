package models

import "time"

type TunnelType string

const (
	TunnelTypeLocal  TunnelType = "local"
	TunnelTypeRemote TunnelType = "remote"
)

type TunnelStatus string

const (
	TunnelStatusPending     TunnelStatus = "pending"
	TunnelStatusRunning     TunnelStatus = "running"
	TunnelStatusStopped     TunnelStatus = "stopped"
	TunnelStatusDisconnected TunnelStatus = "disconnected"
	TunnelStatusReconnecting TunnelStatus = "reconnecting"
)

type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeKey      AuthType = "private_key"
)

type TunnelConfig struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Type         TunnelType `json:"type"`
	SSHServer    string     `json:"ssh_server"`
	SSHUser      string     `json:"ssh_user"`
	AuthType     AuthType   `json:"auth_type"`
	AuthData     string     `json:"-"`
	LocalPort    int        `json:"local_port"`
	RemoteHost   string     `json:"remote_host"`
	RemotePort   int        `json:"remote_port"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type TunnelState struct {
	ID              int64        `json:"id"`
	TunnelID        int64        `json:"tunnel_id"`
	Status          TunnelStatus `json:"status"`
	BytesUp         int64        `json:"bytes_up"`
	BytesDown       int64        `json:"bytes_down"`
	ReconnectCount  int          `json:"reconnect_count"`
	LastConnectedAt *time.Time   `json:"last_connected_at,omitempty"`
	LastDisconnectedAt *time.Time `json:"last_disconnected_at,omitempty"`
	LastDisconnectReason string   `json:"last_disconnect_reason,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type ConnectionLog struct {
	ID               int64     `json:"id"`
	TunnelID         int64     `json:"tunnel_id"`
	ConnectedAt      time.Time `json:"connected_at"`
	DisconnectedAt   *time.Time `json:"disconnected_at,omitempty"`
	DisconnectReason string    `json:"disconnect_reason,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type ApprovalStatus string

const (
	ApprovalPending    ApprovalStatus = "pending"
	ApprovalReviewing  ApprovalStatus = "reviewing"
	ApprovalApproved   ApprovalStatus = "approved"
	ApprovalRejected   ApprovalStatus = "rejected"
	ApprovalExecuting  ApprovalStatus = "executing"
	ApprovalCompleted  ApprovalStatus = "completed"
)

type TunnelRecord struct {
	ID             int64          `json:"id"`
	TunnelConfigID int64          `json:"tunnel_config_id"`
	Status         ApprovalStatus `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type OperationType string

const (
	OpTypeCreate   OperationType = "create"
	OpTypeUpdate   OperationType = "update"
	OpTypeSubmit   OperationType = "submit"
	OpTypeReview   OperationType = "review"
	OpTypeApprove  OperationType = "approve"
	OpTypeReject   OperationType = "reject"
	OpTypeExecute  OperationType = "execute"
	OpTypeComplete OperationType = "complete"
	OpTypeStop     OperationType = "stop"
	OpTypeRestart  OperationType = "restart"
)

type OperationHistory struct {
	ID          int64         `json:"id"`
	RecordID    int64         `json:"record_id"`
	OpType      OperationType `json:"op_type"`
	Operator    string        `json:"operator"`
	Description string        `json:"description"`
	Note        *string       `json:"note,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}
