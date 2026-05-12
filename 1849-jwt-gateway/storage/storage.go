package storage

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"strings"
	"sync"
	"time"
)

type KeyStatus string

const (
	StatusGenerated  KeyStatus = "generated"
	StatusActive     KeyStatus = "active"
	StatusRotating   KeyStatus = "rotating"
	StatusDeprecated KeyStatus = "deprecated"
)

type Key struct {
	ID        string
	Secret    []byte
	Status    KeyStatus
	CreatedAt time.Time
	RotatedAt time.Time
}

type KeyManager struct {
	keys           map[string]*Key
	activeKeyID    string
	rotationPeriod time.Duration
	mu             sync.RWMutex
	nextID         int
}

func NewKeyManager(rotationPeriod time.Duration) *KeyManager {
	km := &KeyManager{
		keys:           make(map[string]*Key),
		rotationPeriod: rotationPeriod,
	}
	km.GenerateKey()
	km.activateLatest()
	return km
}

func (km *KeyManager) GenerateKey() string {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.nextID++
	id := strconv.Itoa(km.nextID)
	secret := make([]byte, 32)
	rand.Read(secret)

	km.keys[id] = &Key{
		ID:        id,
		Secret:    secret,
		Status:    StatusGenerated,
		CreatedAt: time.Now(),
	}
	return id
}

func (km *KeyManager) activateLatest() {
	for id, key := range km.keys {
		if key.Status == StatusGenerated {
			key.Status = StatusActive
			km.activeKeyID = id
			break
		}
	}
}

func (km *KeyManager) RotateKey() (string, string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	oldID := km.activeKeyID
	if oldID == "" {
		return "", ""
	}

	oldKey := km.keys[oldID]
	if oldKey == nil {
		return "", ""
	}

	newID := km.GenerateKey()
	newKey := km.keys[newID]
	newKey.Status = StatusActive
	km.activeKeyID = newID

	oldKey.Status = StatusRotating
	oldKey.RotatedAt = time.Now()

	go km.deprecateAfter(oldID, km.rotationPeriod)

	return oldID, newID
}

func (km *KeyManager) deprecateAfter(id string, period time.Duration) {
	time.Sleep(period)
	km.mu.Lock()
	defer km.mu.Unlock()
	if key, ok := km.keys[id]; ok && key.Status == StatusRotating {
		key.Status = StatusDeprecated
	}
}

func (km *KeyManager) GetKey(id string) *Key {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.keys[id]
}

func (km *KeyManager) GetActiveKey() *Key {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.keys[km.activeKeyID]
}

func (km *KeyManager) ListKeys() []*Key {
	km.mu.RLock()
	defer km.mu.RUnlock()
	keys := make([]*Key, 0, len(km.keys))
	for _, key := range km.keys {
		keys = append(keys, key)
	}
	return keys
}

type AuditEntry struct {
	Timestamp time.Time
	ClientID  string
	Path      string
	Result    string
}

type AuditLog struct {
	entries []AuditEntry
	mu      sync.RWMutex
	max     int
}

func NewAuditLog() *AuditLog {
	return &AuditLog{
		entries: make([]AuditEntry, 0),
		max:     10000,
	}
}

func (al *AuditLog) Add(entry AuditEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.entries = append(al.entries, entry)
	if len(al.entries) > al.max {
		al.entries = al.entries[len(al.entries)-al.max:]
	}
}

func (al *AuditLog) List() []AuditEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()
	result := make([]AuditEntry, len(al.entries))
	copy(result, al.entries)
	return result
}

func ParseBackends(backends string) []string {
	if backends == "" {
		return nil
	}
	parts := strings.Split(backends, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func MustParseDuration(seconds string) time.Duration {
	s, err := strconv.Atoi(seconds)
	if err != nil {
		return 3600 * time.Second
	}
	return time.Duration(s) * time.Second
}

func EncodeKey(secret []byte) string {
	return base64.StdEncoding.EncodeToString(secret)
}
