package common

import "time"

type OperationType int

const (
	Create OperationType = iota
	Update
	Delete
)

type Operation struct {
	Type OperationType `json:"type"`
	Key  string        `json:"key"`
	Data interface{}   `json:"data,omitempty"`
}

type TransactionStatus int

const (
	Pending TransactionStatus = iota
	Prepared
	Committed
	Aborted
	RolledBack
)

func (s TransactionStatus) String() string {
	switch s {
	case Pending:
		return "pending"
	case Prepared:
		return "prepared"
	case Committed:
		return "committed"
	case Aborted:
		return "aborted"
	case RolledBack:
		return "rolledback"
	default:
		return "unknown"
	}
}

type BeginTxRequest struct {
}

type BeginTxResponse struct {
	Success bool   `json:"success"`
	TxID    string `json:"tx_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

type PrepareTxRequest struct {
	TxID       string                     `json:"tx_id"`
	Operations map[string][]Operation     `json:"operations"`
}

type PrepareTxResponse struct {
	Success bool              `json:"success"`
	Status  string            `json:"status,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type CommitTxRequest struct {
	TxID string `json:"tx_id"`
}

type CommitTxResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status,omitempty"`
	Error   string `json:"error,omitempty"`
}

type RollbackTxRequest struct {
	TxID string `json:"tx_id"`
}

type RollbackTxResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status,omitempty"`
	Error   string `json:"error,omitempty"`
}

type StatusTxRequest struct {
	TxID string `json:"tx_id"`
}

type StatusTxResponse struct {
	Success   bool                   `json:"success"`
	TxID      string                 `json:"tx_id,omitempty"`
	Status    string                 `json:"status,omitempty"`
	CreatedAt time.Time              `json:"created_at,omitempty"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

type ListParticipantsResponse struct {
	Success bool                      `json:"success"`
	Participants []ParticipantStatus  `json:"participants,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

type ParticipantStatus struct {
	ID           string   `json:"id"`
	PreparedTx   []string `json:"prepared_tx"`
	Unavailable  bool     `json:"unavailable"`
	StorageSize  int      `json:"storage_size"`
}

type GetParticipantDataRequest struct {
	ParticipantID string `json:"participant_id"`
}

type GetParticipantDataResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type AddParticipantRequest struct {
	ParticipantID string `json:"participant_id"`
}

type AddParticipantResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type RemoveParticipantRequest struct {
	ParticipantID string `json:"participant_id"`
}

type RemoveParticipantResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type SetParticipantUnavailableRequest struct {
	ParticipantID string `json:"participant_id"`
	Unavailable   bool   `json:"unavailable"`
}

type SetParticipantUnavailableResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
