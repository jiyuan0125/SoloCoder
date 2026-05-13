package resource

import (
	"time"

	"serviceregistry/internal/model"
	"serviceregistry/internal/storage"
)

type ResourceStorage interface {
	SaveResource(r *model.Resource) error
	GetResource(id string) (*model.Resource, error)
	AddAssociation(assoc *model.ResourceAssociation) error
	GetResourceSummary(resourceID string) ([]*model.ResourceAssociation, error)
}

type Manager struct {
	storage ResourceStorage
}

func NewManager(storage ResourceStorage) *Manager {
	return &Manager{storage: storage}
}

func (m *Manager) Create(id string, rType model.ResourceType, name string) error {
	r := &model.Resource{
		ID:   id,
		Type: rType,
		Name: name,
	}
	return m.storage.SaveResource(r)
}

func (m *Manager) Get(id string) (*model.Resource, error) {
	return m.storage.GetResource(id)
}

func (m *Manager) AddAssociation(resourceID string, targetID string, targetType model.ResourceType, operation string) error {
	assoc := &model.ResourceAssociation{
		ResourceID: resourceID,
		TargetID:   targetID,
		TargetType: targetType,
		Operation:  operation,
		Timestamp:  time.Now(),
	}
	return m.storage.AddAssociation(assoc)
}

func (m *Manager) GetSummary(resourceID string) ([]*model.ResourceAssociation, error) {
	return m.storage.GetResourceSummary(resourceID)
}

var _ ResourceStorage = (*storage.SQLiteStorage)(nil)
