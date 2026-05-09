package main

import (
	"crypto/rsa"
	"fmt"
	"sync"

	"rsa-crypto/common"
)

type KeyManager struct {
	mu             sync.RWMutex
	publicKey      *rsa.PublicKey
	privateKey     *rsa.PrivateKey
	keySize        int
	currentPadding common.PaddingMode

	tasks      map[string]*common.KeyGenTask
	tasksMu    sync.RWMutex
	nextTaskID int
}

func NewKeyManager() *KeyManager {
	return &KeyManager{
		currentPadding: common.PaddingOAEP,
		tasks:          make(map[string]*common.KeyGenTask),
	}
}

func (km *KeyManager) SetKeys(pub *rsa.PublicKey, priv *rsa.PrivateKey, keySize int) {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.publicKey = pub
	km.privateKey = priv
	km.keySize = keySize
}

func (km *KeyManager) GetKeys() (*rsa.PublicKey, *rsa.PrivateKey) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.publicKey, km.privateKey
}

func (km *KeyManager) GetKeyInfo() common.KeyInfoResponse {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return common.KeyInfoResponse{
		KeySize:       km.keySize,
		Padding:       km.currentPadding,
		HasPublicKey:  km.publicKey != nil,
		HasPrivateKey: km.privateKey != nil,
	}
}

func (km *KeyManager) SetPadding(padding common.PaddingMode) {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.currentPadding = padding
}

func (km *KeyManager) CreateTask(keySize int) string {
	km.tasksMu.Lock()
	defer km.tasksMu.Unlock()
	km.nextTaskID++
	taskID := fmt.Sprintf("task-%d", km.nextTaskID)
	km.tasks[taskID] = &common.KeyGenTask{
		TaskID:  taskID,
		Status:  common.TaskStatusPending,
		KeySize: keySize,
	}
	return taskID
}

func (km *KeyManager) GetTask(taskID string) (*common.KeyGenTask, bool) {
	km.tasksMu.RLock()
	defer km.tasksMu.RUnlock()
	task, exists := km.tasks[taskID]
	return task, exists
}

func (km *KeyManager) UpdateTask(taskID string, status common.TaskStatus, pubKey, privKey string, errMsg string) {
	km.tasksMu.Lock()
	defer km.tasksMu.Unlock()
	if task, exists := km.tasks[taskID]; exists {
		task.Status = status
		task.PublicKey = pubKey
		task.PrivateKey = privKey
		task.Error = errMsg
	}
}
