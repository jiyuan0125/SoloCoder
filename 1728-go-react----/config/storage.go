package config

import (
	"fmt"
	"lab-safety/models"
	"sync"
	"time"
)

var (
	Storage = &StorageType{
		Labs:          make(map[string]*models.Lab),
		Chemicals:     make(map[string]*models.Chemical),
		Usages:        make(map[string]*models.UsageRecord),
		Trainings:     make(map[string]*models.Training),
		Checks:        make(map[string]*models.SafetyCheck),
		Todos:         make(map[string]*models.Todo),
		AuditLogs:     make(map[string]*models.AuditLog),
		Departments:   make(map[string]*models.Department),
		Requests:      make(map[string]*models.Request),
		NextID:        1,
	}

	validDangerLevels    = map[string]bool{"甲": true, "乙": true, "丙": true, "丁": true}
	validHazardCategories = map[string]bool{
		"易燃": true, "易爆": true, "有毒": true, "腐蚀": true, "氧化": true, "其他": true,
	}
)

type StorageType struct {
	Labs          map[string]*models.Lab
	Chemicals     map[string]*models.Chemical
	Usages        map[string]*models.UsageRecord
	Trainings     map[string]*models.Training
	Checks        map[string]*models.SafetyCheck
	Todos         map[string]*models.Todo
	AuditLogs     map[string]*models.AuditLog
	Departments   map[string]*models.Department
	Requests      map[string]*models.Request
	NextID        int
	mutex         sync.RWMutex
}

func (s *StorageType) Lock() {
	s.mutex.Lock()
}

func (s *StorageType) Unlock() {
	s.mutex.Unlock()
}

func (s *StorageType) RLock() {
	s.mutex.RLock()
}

func (s *StorageType) RUnlock() {
	s.mutex.RUnlock()
}

func ValidDangerLevel(level string) bool {
	return validDangerLevels[level]
}

func ValidHazardCategory(cat string) bool {
	return validHazardCategories[cat]
}

func InitStorage() {
}

func StartExpiryChecker() {
	go func() {
		for {
			now := time.Now().Truncate(24 * time.Hour)
			next := now.Add(24 * time.Hour)
			time.Sleep(time.Until(next))
			checkExpiry()
		}
	}()
}

func checkExpiry() {
	Storage.Lock()
	defer Storage.Unlock()

	today := time.Now()
	for _, c := range Storage.Chemicals {
		if c.ExpiryDate != nil && c.ExpiryDate.Before(today) && c.Status != "已过期" {
			c.Status = "已过期"
			AddAuditLog("危化品过期", "system", "危化品 ID: %s, 名称: %s", c.ID, c.Name)
		}
	}
}

func AddAuditLog(action, operator, format string, args ...interface{}) {
	log := &models.AuditLog{
		ID:        generateID(),
		Action:    action,
		Operator:  operator,
		Timestamp: time.Now(),
		Details:   fmt.Sprintf(format, args...),
	}
	Storage.AuditLogs[log.ID] = log
}

func generateID() string {
	id := Storage.NextID
	Storage.NextID++
	return fmt.Sprintf("%d", id)
}
