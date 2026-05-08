package lock

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type LockMode int

const (
	LockModeExclusive LockMode = iota
	LockModeShared
)

const (
	DefaultTimeout         = 30 * time.Second
	DefaultMaxWaitDuration = 5 * time.Minute
	LockFileSuffix         = ".lock"
)

type LockHolderInfo struct {
	PID       int       `json:"pid"`
	StartTime time.Time `json:"start_time"`
	Mode      string    `json:"mode"`
}

func (h *LockHolderInfo) ToBytes() ([]byte, error) {
	return json.Marshal(h)
}

func ParseLockHolderInfo(data []byte) (*LockHolderInfo, error) {
	var info LockHolderInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (h *LockHolderInfo) IsProcessAlive() bool {
	p, err := os.FindProcess(h.PID)
	if err != nil {
		return false
	}
	err = p.Signal(os.Signal(nil))
	return err == nil
}

type LockConfig struct {
	Timeout         time.Duration
	MaxWaitDuration time.Duration
}

func DefaultConfig() LockConfig {
	return LockConfig{
		Timeout:         DefaultTimeout,
		MaxWaitDuration: DefaultMaxWaitDuration,
	}
}

type Lock struct {
	filePath     string
	lockFilePath string
	file         *os.File
	mode         LockMode
	held         bool
	mu           sync.Mutex
	localMu      sync.RWMutex
	holdCount    int
	holdMu       sync.Mutex
}
