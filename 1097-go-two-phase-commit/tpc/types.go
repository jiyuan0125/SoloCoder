package tpc

import (
	"sync"
	"time"
)

type OperationType int

const (
	Create OperationType = iota
	Update
	Delete
)

type Operation struct {
	Type OperationType
	Key  string
	Data interface{}
}

type TransactionStatus int

const (
	Pending TransactionStatus = iota
	Prepared
	Committed
	Aborted
	RolledBack
)

type TransactionState struct {
	ID        string
	Status    TransactionStatus
	Operations map[string][]Operation
	CreatedAt time.Time
	UpdatedAt time.Time
	PreparedAt time.Time
	CommitResult []ParticipantResult
}

type ParticipantResult struct {
	ParticipantID string
	Success       bool
	Error         string
}

type TransactionLogEntry struct {
	TxID     string
	Status   TransactionStatus
	Timestamp time.Time
}

type ITransactionCoordinator interface {
	Begin() (*TransactionState, error)
	Prepare(txID string, operations map[string][]Operation) (*TransactionState, error)
	Commit(txID string) (*TransactionState, error)
	Rollback(txID string) (*TransactionState, error)
	GetStatus(txID string) (*TransactionState, error)
	GetUnfinishedTransactions() []*TransactionState
	AddParticipant(participant ITransactionParticipant) error
	RemoveParticipant(participantID string) error
	GetParticipants() []ITransactionParticipant
	Recover() error
}

type ITransactionParticipant interface {
	ID() string
	Prepare(txID string, changes []Operation) error
	Commit(txID string) error
	Rollback(txID string) error
	GetData() map[string]interface{}
	GetStatus() ParticipantStatus
	SetUnavailable(unavailable bool)
}

type ParticipantStatus struct {
	ID           string
	PreparedTx   []string
	Unavailable  bool
	StorageSize  int
}
