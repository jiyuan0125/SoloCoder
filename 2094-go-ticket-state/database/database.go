package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ticket-system/models"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(dsn string) (*DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DB{db}, nil
}

func (db *DB) Init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tickets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL,
		priority TEXT NOT NULL,
		creator_id TEXT NOT NULL,
		assignee_id TEXT,
		parent_ticket_id INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		pending_since DATETIME,
		resolved_since DATETIME,
		FOREIGN KEY (parent_ticket_id) REFERENCES tickets(id)
	);

	CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
	CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);
	CREATE INDEX IF NOT EXISTS idx_tickets_assignee ON tickets(assignee_id);
	CREATE INDEX IF NOT EXISTS idx_tickets_created ON tickets(created_at);
	CREATE INDEX IF NOT EXISTS idx_tickets_pending_since ON tickets(pending_since);
	CREATE INDEX IF NOT EXISTS idx_tickets_resolved_since ON tickets(resolved_since);

	CREATE TABLE IF NOT EXISTS ticket_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ticket_id INTEGER NOT NULL,
		from_status TEXT NOT NULL,
		to_status TEXT NOT NULL,
		user_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (ticket_id) REFERENCES tickets(id)
	);

	CREATE INDEX IF NOT EXISTS idx_history_ticket ON ticket_history(ticket_id);

	CREATE TABLE IF NOT EXISTS ticket_links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		parent_id INTEGER NOT NULL,
		child_id INTEGER NOT NULL,
		is_processed INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (parent_id) REFERENCES tickets(id),
		FOREIGN KEY (child_id) REFERENCES tickets(id),
		UNIQUE(parent_id, child_id)
	);

	CREATE INDEX IF NOT EXISTS idx_links_parent ON ticket_links(parent_id);
	CREATE INDEX IF NOT EXISTS idx_links_child ON ticket_links(child_id);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to init schema: %w", err)
	}
	return nil
}

func (db *DB) CreateTicket(ctx context.Context, req *models.CreateTicketRequest) (*models.Ticket, error) {
	now := time.Now()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO tickets (title, description, status, priority, creator_id, assignee_id, parent_ticket_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NULL, ?, ?, ?)
	`, req.Title, req.Description, models.TicketStatusNew, req.Priority, req.CreatorID, req.ParentTicketID, now, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if req.ParentTicketID != nil {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO ticket_links (parent_id, child_id, is_processed, created_at)
			VALUES (?, ?, 0, ?)
		`, *req.ParentTicketID, id, now)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return db.GetTicketByID(ctx, id)
}

func (db *DB) GetTicketByID(ctx context.Context, id int64) (*models.Ticket, error) {
	t := &models.Ticket{}
	var assigneeID sql.NullString
	var parentID sql.NullInt64
	var pendingSince, resolvedSince sql.NullTime

	err := db.QueryRowContext(ctx, `
		SELECT id, title, description, status, priority, creator_id, assignee_id, parent_ticket_id, created_at, updated_at, pending_since, resolved_since
		FROM tickets WHERE id = ?
	`, id).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.CreatorID, &assigneeID, &parentID, &t.CreatedAt, &t.UpdatedAt, &pendingSince, &resolvedSince)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.String
	}
	if parentID.Valid {
		t.ParentTicketID = &parentID.Int64
	}
	if pendingSince.Valid {
		t.PendingSince = &pendingSince.Time
	}
	if resolvedSince.Valid {
		t.ResolvedSince = &resolvedSince.Time
	}

	return t, nil
}

type QueryTicketsParams struct {
	Priority    *string
	Status      *string
	AssigneeID  *string
	CreatorID   *string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

func (db *DB) QueryTickets(ctx context.Context, params *QueryTicketsParams) ([]*models.Ticket, error) {
	query := `
		SELECT id, title, description, status, priority, creator_id, assignee_id, parent_ticket_id, created_at, updated_at, pending_since, resolved_since
		FROM tickets
		WHERE 1=1
	`
	var args []interface{}

	if params.Priority != nil {
		query += " AND priority = ?"
		args = append(args, *params.Priority)
	}
	if params.Status != nil {
		query += " AND status = ?"
		args = append(args, *params.Status)
	}
	if params.AssigneeID != nil {
		query += " AND assignee_id = ?"
		args = append(args, *params.AssigneeID)
	}
	if params.CreatorID != nil {
		query += " AND creator_id = ?"
		args = append(args, *params.CreatorID)
	}
	if params.CreatedFrom != nil {
		query += " AND created_at >= ?"
		args = append(args, *params.CreatedFrom)
	}
	if params.CreatedTo != nil {
		query += " AND created_at <= ?"
		args = append(args, *params.CreatedTo)
	}

	query += " ORDER BY (CASE priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END) ASC, created_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := make([]*models.Ticket, 0)
	for rows.Next() {
		t := &models.Ticket{}
		var assigneeID sql.NullString
		var parentID sql.NullInt64
		var pendingSince, resolvedSince sql.NullTime

		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.CreatorID, &assigneeID, &parentID, &t.CreatedAt, &t.UpdatedAt, &pendingSince, &resolvedSince)
		if err != nil {
			return nil, err
		}

		if assigneeID.Valid {
			t.AssigneeID = &assigneeID.String
		}
		if parentID.Valid {
			t.ParentTicketID = &parentID.Int64
		}
		if pendingSince.Valid {
			t.PendingSince = &pendingSince.Time
		}
		if resolvedSince.Valid {
			t.ResolvedSince = &resolvedSince.Time
		}

		tickets = append(tickets, t)
	}

	return tickets, rows.Err()
}

func (db *DB) UpdateTicketStatus(ctx context.Context, tx *sql.Tx, ticketID int64, fromStatus, toStatus, userID string) error {
	now := time.Now()

	var setClause string
	var args []interface{}

	switch toStatus {
	case models.TicketStatusPendingCust:
		setClause = "status = ?, updated_at = ?, pending_since = ?"
		args = append(args, toStatus, now, now)
	case models.TicketStatusResolved:
		setClause = "status = ?, updated_at = ?, resolved_since = ?, pending_since = NULL"
		args = append(args, toStatus, now, now)
	default:
		setClause = "status = ?, updated_at = ?, pending_since = NULL"
		args = append(args, toStatus, now)
	}

	query := "UPDATE tickets SET " + setClause + " WHERE id = ?"
	args = append(args, ticketID)

	var execDB interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	} = db.DB
	if tx != nil {
		execDB = tx
	}

	result, err := execDB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	_, err = execDB.ExecContext(ctx, `
		INSERT INTO ticket_history (ticket_id, from_status, to_status, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, ticketID, fromStatus, toStatus, userID, now)
	return err
}

func (db *DB) AssignTicket(ctx context.Context, ticketID int64, userID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var assigneeID sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT assignee_id FROM tickets WHERE id = ? FOR UPDATE", ticketID).Scan(&assigneeID)
	if err == sql.ErrNoRows {
		return err
	}
	if err != nil {
		return err
	}

	if assigneeID.Valid {
		return fmt.Errorf("already assigned to %s", assigneeID.String)
	}

	now := time.Now()
	result, err := tx.ExecContext(ctx, `
		UPDATE tickets SET assignee_id = ?, status = ?, updated_at = ? WHERE id = ?
	`, userID, models.TicketStatusAssigned, now, ticketID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_history (ticket_id, from_status, to_status, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, ticketID, models.TicketStatusNew, models.TicketStatusAssigned, userID, now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) GetPendingAutoResolve(ctx context.Context, cutoff time.Time) ([]int64, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id FROM tickets 
		WHERE status = ? AND pending_since IS NOT NULL AND pending_since <= ?
	`, models.TicketStatusPendingCust, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (db *DB) GetPendingAutoClose(ctx context.Context, cutoff time.Time) ([]int64, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id FROM tickets 
		WHERE status = ? AND resolved_since IS NOT NULL AND resolved_since <= ?
	`, models.TicketStatusResolved, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (db *DB) AutoResolveTicket(ctx context.Context, id int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	result, err := tx.ExecContext(ctx, `
		UPDATE tickets SET status = ?, updated_at = ?, resolved_since = ?, pending_since = NULL
		WHERE id = ? AND status = ?
	`, models.TicketStatusResolved, now, now, id, models.TicketStatusPendingCust)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return tx.Commit()
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_history (ticket_id, from_status, to_status, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, id, models.TicketStatusPendingCust, models.TicketStatusResolved, "system", now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) AutoCloseTicket(ctx context.Context, id int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	result, err := tx.ExecContext(ctx, `
		UPDATE tickets SET status = ?, updated_at = ?
		WHERE id = ? AND status = ?
	`, models.TicketStatusClosed, now, id, models.TicketStatusResolved)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return tx.Commit()
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_history (ticket_id, from_status, to_status, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, id, models.TicketStatusResolved, models.TicketStatusClosed, "system", now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) GetUnprocessedLinks(ctx context.Context, parentID int64) ([]*models.TicketLink, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, parent_id, child_id, is_processed, created_at
		FROM ticket_links WHERE parent_id = ? AND is_processed = 0
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]*models.TicketLink, 0)
	for rows.Next() {
		link := &models.TicketLink{}
		err := rows.Scan(&link.ID, &link.ParentID, &link.ChildID, &link.IsProcessed, &link.CreatedAt)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

func (db *DB) MarkLinksProcessed(ctx context.Context, linkIDs []int64) error {
	if len(linkIDs) == 0 {
		return nil
	}

	query := "UPDATE ticket_links SET is_processed = 1 WHERE id IN ("
	args := make([]interface{}, 0, len(linkIDs))
	for i, id := range linkIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"

	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func (db *DB) GetTicketHistory(ctx context.Context, ticketID int64) ([]*models.TicketHistory, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, ticket_id, from_status, to_status, user_id, created_at
		FROM ticket_history WHERE ticket_id = ? ORDER BY created_at ASC
	`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]*models.TicketHistory, 0)
	for rows.Next() {
		h := &models.TicketHistory{}
		err := rows.Scan(&h.ID, &h.TicketID, &h.FromStatus, &h.ToStatus, &h.UserID, &h.CreatedAt)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}
