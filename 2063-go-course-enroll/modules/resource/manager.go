package resource

import (
	"context"
	"fmt"
	"time"

	"course-enroll/storage"
)

type Manager struct {
	store storage.Store
}

func NewManager(store storage.Store) *Manager {
	return &Manager{store: store}
}

type ResourceInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type Association struct {
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Operation    string    `json:"operation"`
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type ResourceSummary struct {
	Resource      *ResourceInfo  `json:"resource"`
	Associations  []*Association `json:"associations"`
	OperationCount map[string]int `json:"operation_count"`
}

func (m *Manager) CreateResource(ctx context.Context, info *ResourceInfo) error {
	tx, err := m.store.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	err = m.store.CreateResource(ctx, tx, &storage.Resource{
		ID:   info.ID,
		Type: info.Type,
		Name: info.Name,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) GetResourceSummary(ctx context.Context, id, resourceType string) (*ResourceSummary, error) {
	resource, err := m.store.GetResource(ctx, id, resourceType)
	if err != nil {
		return nil, fmt.Errorf("get resource: %w", err)
	}

	assocs, err := m.store.ListResourceAssociations(ctx, id, resourceType)
	if err != nil {
		return nil, fmt.Errorf("list associations: %w", err)
	}

	operationCount := make(map[string]int)
	associationList := make([]*Association, len(assocs))

	for i, a := range assocs {
		associationList[i] = &Association{
			ResourceType: a.ResourceType,
			ResourceID:   a.ResourceID,
			Operation:    a.Operation,
			EntityType:   a.EntityType,
			EntityID:     a.EntityID,
			CreatedAt:    a.CreatedAt,
		}
		operationCount[a.Operation]++
	}

	summary := &ResourceSummary{
		Associations:   associationList,
		OperationCount: operationCount,
	}

	if resource != nil {
		summary.Resource = &ResourceInfo{
			ID:   resource.ID,
			Type: resource.Type,
			Name: resource.Name,
		}
	}

	return summary, nil
}
