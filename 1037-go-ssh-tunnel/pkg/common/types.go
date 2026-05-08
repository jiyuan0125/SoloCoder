package common

import "time"

type TunnelMode string

const (
	ModeLocal  TunnelMode = "local"
	ModeRemote TunnelMode = "remote"
)

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password"
	AuthMethodKey      AuthMethod = "key"
)

type TunnelConfig struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Mode             TunnelMode `json:"mode"`
	SSHServer        string     `json:"ssh_server"`
	SSHPort          int        `json:"ssh_port"`
	SSHUser          string     `json:"ssh_user"`
	AuthMethod       AuthMethod `json:"auth_method"`
	Password         string     `json:"password,omitempty"`
	KeyFilePath      string     `json:"key_file_path,omitempty"`
	LocalAddress     string     `json:"local_address"`
	LocalPort        int        `json:"local_port"`
	RemoteAddress    string     `json:"remote_address"`
	RemotePort       int        `json:"remote_port"`
	KeepaliveInterval int       `json:"keepalive_interval"`
	MaxRetryInterval int        `json:"max_retry_interval"`
}

type TunnelStatus string

const (
	StatusConnecting    TunnelStatus = "connecting"
	StatusRunning       TunnelStatus = "running"
	StatusReconnecting  TunnelStatus = "reconnecting"
	StatusStopped       TunnelStatus = "stopped"
	StatusError         TunnelStatus = "error"
)

type TunnelState struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Status     TunnelStatus `json:"status"`
	Error      string       `json:"error,omitempty"`
	StartedAt  time.Time    `json:"started_at"`
	LastActive time.Time    `json:"last_active"`
}

type CreateTunnelRequest struct {
	TunnelConfig
}

type CreateTunnelResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Tunnel  *TunnelState `json:"tunnel,omitempty"`
}

type StopTunnelRequest struct {
	ID string `json:"id"`
}

type StopTunnelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ListTunnelsResponse struct {
	Success bool          `json:"success"`
	Tunnels []*TunnelState `json:"tunnels"`
}

type GetTunnelRequest struct {
	ID string `json:"id"`
}

type GetTunnelResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Tunnel  *TunnelState `json:"tunnel,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
