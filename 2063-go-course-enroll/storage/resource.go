package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *SQLiteStore) CreateResource(ctx context.Context, tx Tx, resource *Resource) error {
	stx := tx.(*sqliteTx).tx

	_, err := stx.ExecContext(ctx, `
		INSERT OR IGNORE INTO resources (id, type, name) VALUES (?, ?, ?)
	`, resource.ID, resource.Type, resource.Name)
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetResource(ctx context.Context, id, resourceType string) (*Resource, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, type, name FROM resources WHERE id = ? AND type = ?
	`, id, resourceType)

	resource := &Resource{}
	err := row.Scan(&resource.ID, &resource.Type, &resource.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get resource: %w", err)
	}
	return resource, nil
}

func (s *SQLiteStore) LogResourceAssociation(ctx context.Context, tx Tx, assoc *ResourceAssociation) error {
	stx := tx.(*sqliteTx).tx

	assoc.CreatedAt = time.Now()

	_, err := stx.ExecContext(ctx, `
		INSERT INTO resource_associations (resource_type, resource_id, operation, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		assoc.ResourceType,
		assoc.ResourceID,
		assoc.Operation,
		assoc.EntityType,
		assoc.EntityID,
		timeToStr(assoc.CreatedAt),
	)
	if err != nil {
		return fmt.Errorf("log resource association: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListResourceAssociations(ctx context.Context, id, resourceType string) ([]*ResourceAssociation, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT resource_type, resource_id, operation, entity_type, entity_id, created_at
		FROM resource_associations WHERE resource_id = ? AND resource_type = ?
		ORDER BY created_at DESC
	`, id, resourceType)
	if err != nil {
		return nil, fmt.Errorf("list resource associations: %w", err)
	}
	defer rows.Close()

	var assocs []*ResourceAssociation
	for rows.Next() {
		assoc := &ResourceAssociation{}
		var createdAt string

		err := rows.Scan(
			&assoc.ResourceType,
			&assoc.ResourceID,
			&assoc.Operation,
			&assoc.EntityType,
			&assoc.EntityID,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan association: %w", err)
		}
		assoc.CreatedAt = strToTime(createdAt)
		assocs = append(assocs, assoc)
	}

	return assocs, nil
}
