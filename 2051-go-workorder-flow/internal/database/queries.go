package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"workorder-flow/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
)

func nullableTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func nullableInt64(i sql.NullInt64) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

func (db *DB) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	var teamID sql.NullInt64
	query := `SELECT id, username, email, phone, team_id, created_at, updated_at FROM users WHERE id = ?`
	err := db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Phone, &teamID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if teamID.Valid {
		user.TeamID = &teamID.Int64
	}
	return &user, nil
}

func (db *DB) GetTeamByID(ctx context.Context, id int64) (*models.Team, error) {
	var team models.Team
	query := `SELECT id, name, type, leader_id, created_at, updated_at FROM teams WHERE id = ?`
	err := db.QueryRowContext(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Type, &team.LeaderID, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &team, nil
}

func (db *DB) GetTeamByType(ctx context.Context, typ models.WorkOrderType) (*models.Team, error) {
	var team models.Team
	query := `SELECT id, name, type, leader_id, created_at, updated_at FROM teams WHERE type = ?`
	err := db.QueryRowContext(ctx, query, string(typ)).Scan(
		&team.ID, &team.Name, &team.Type, &team.LeaderID, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &team, nil
}

func (db *DB) GetDutyUserByTeamAndDate(ctx context.Context, teamID int64, date string) (*models.User, error) {
	var userID int64
	query := `SELECT ds.user_id FROM duty_schedules ds WHERE ds.team_id = ? AND ds.duty_date = ?`
	err := db.QueryRowContext(ctx, query, teamID, date).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return db.GetUserByID(ctx, userID)
}

func (db *DB) GetResourceTypeByID(ctx context.Context, id int64) (*models.ResourceType, error) {
	var rt models.ResourceType
	query := `SELECT id, name, description, created_at, updated_at FROM resource_types WHERE id = ?`
	err := db.QueryRowContext(ctx, query, id).Scan(
		&rt.ID, &rt.Name, &rt.Description, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rt, nil
}

func (db *DB) GetResourceTypeByName(ctx context.Context, name string) (*models.ResourceType, error) {
	var rt models.ResourceType
	query := `SELECT id, name, description, created_at, updated_at FROM resource_types WHERE name = ?`
	err := db.QueryRowContext(ctx, query, name).Scan(
		&rt.ID, &rt.Name, &rt.Description, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &rt, nil
}

func (db *DB) ListResourceTypes(ctx context.Context) ([]models.ResourceType, error) {
	var rts []models.ResourceType
	query := `SELECT id, name, description, created_at, updated_at FROM resource_types ORDER BY id`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rt models.ResourceType
		if err := rows.Scan(&rt.ID, &rt.Name, &rt.Description, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		rts = append(rts, rt)
	}
	return rts, nil
}

func (db *DB) GetResourceByID(ctx context.Context, id int64) (*models.Resource, error) {
	var r models.Resource
	query := `SELECT id, resource_type_id, name, description, created_at, updated_at FROM resources WHERE id = ?`
	err := db.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.ResourceTypeID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (db *DB) ListResourcesByType(ctx context.Context, resourceTypeID int64) ([]models.Resource, error) {
	var resources []models.Resource
	query := `SELECT id, resource_type_id, name, description, created_at, updated_at FROM resources WHERE resource_type_id = ? ORDER BY id`
	rows, err := db.QueryContext(ctx, query, resourceTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Resource
		if err := rows.Scan(&r.ID, &r.ResourceTypeID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (db *DB) CreateWorkOrder(ctx context.Context, wo *models.WorkOrder, resourceIDs []int64) (*models.WorkOrder, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `INSERT INTO work_orders (type, title, description, content, current_handler_id, current_team_id, process_status, work_order_status, submitter_id, escalated, deadline_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, query,
		wo.Type, wo.Title, wo.Description, wo.Content,
		wo.CurrentHandlerID, wo.CurrentTeamID, wo.ProcessStatus, wo.WorkOrderStatus,
		wo.SubmitterID, wo.Escalated, wo.DeadlineAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	for _, rid := range resourceIDs {
		_, err := tx.ExecContext(ctx, `INSERT INTO work_order_resources (work_order_id, resource_id) VALUES (?, ?)`, id, rid)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return db.GetWorkOrderByID(ctx, id)
}

func (db *DB) GetWorkOrderByID(ctx context.Context, id int64) (*models.WorkOrder, error) {
	var wo models.WorkOrder
	var currentHandlerID sql.NullInt64
	var escalated sql.NullInt64
	var rating sql.NullInt64
	var escalatedAt sql.NullTime
	var deadlineAt sql.NullTime
	var processStartAt sql.NullTime

	query := `SELECT id, type, title, description, content, current_handler_id, current_team_id, process_status, work_order_status, submitter_id, escalated, rating, escalated_at, deadline_at, process_start_at, created_at, updated_at FROM work_orders WHERE id = ?`
	err := db.QueryRowContext(ctx, query, id).Scan(
		&wo.ID, &wo.Type, &wo.Title, &wo.Description, &wo.Content,
		&currentHandlerID, &wo.CurrentTeamID, &wo.ProcessStatus, &wo.WorkOrderStatus,
		&wo.SubmitterID, &escalated, &rating, &escalatedAt, &deadlineAt, &processStartAt,
		&wo.CreatedAt, &wo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if currentHandlerID.Valid {
		wo.CurrentHandlerID = &currentHandlerID.Int64
	}
	wo.Escalated = escalated.Int64 > 0
	if rating.Valid {
		r := int(rating.Int64)
		wo.Rating = &r
	}
	wo.EscalatedAt = nullableTime(escalatedAt)
	wo.DeadlineAt = nullableTime(deadlineAt)
	wo.ProcessStartAt = nullableTime(processStartAt)

	return &wo, nil
}

func (db *DB) GetWorkOrderResources(ctx context.Context, workOrderID int64) ([]int64, error) {
	var resourceIDs []int64
	query := `SELECT resource_id FROM work_order_resources WHERE work_order_id = ?`
	rows, err := db.QueryContext(ctx, query, workOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rid int64
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		resourceIDs = append(resourceIDs, rid)
	}
	return resourceIDs, nil
}

func (db *DB) UpdateWorkOrder(ctx context.Context, wo *models.WorkOrder) error {
	query := `UPDATE work_orders SET current_handler_id = ?, current_team_id = ?, process_status = ?, work_order_status = ?, escalated = ?, rating = ?, escalated_at = ?, deadline_at = ?, process_start_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.ExecContext(ctx, query,
		wo.CurrentHandlerID, wo.CurrentTeamID, wo.ProcessStatus, wo.WorkOrderStatus,
		wo.Escalated, wo.Rating, wo.EscalatedAt, wo.DeadlineAt, wo.ProcessStartAt, wo.ID,
	)
	return err
}

func (db *DB) AddHistoryRecord(ctx context.Context, record *models.HistoryRecord) error {
	query := `INSERT INTO history_records (work_order_id, operation_type, operator_id, old_process_status, new_process_status, old_work_order_status, new_work_order_status, assigned_from, assigned_to, reason, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, query,
		record.WorkOrderID, record.OperationType, record.OperatorID,
		record.OldProcessStatus, record.NewProcessStatus,
		record.OldWorkOrderStatus, record.NewWorkOrderStatus,
		record.AssignedFrom, record.AssignedTo, record.Reason, record.Comment,
	)
	return err
}

func (db *DB) GetWorkOrderHistory(ctx context.Context, workOrderID int64) ([]models.HistoryRecord, error) {
	var records []models.HistoryRecord
	query := `SELECT id, work_order_id, operation_type, operator_id, old_process_status, new_process_status, old_work_order_status, new_work_order_status, assigned_from, assigned_to, reason, comment, created_at FROM history_records WHERE work_order_id = ? ORDER BY created_at DESC`
	rows, err := db.QueryContext(ctx, query, workOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.HistoryRecord
		var assignedFrom, assignedTo sql.NullInt64
		if err := rows.Scan(&r.ID, &r.WorkOrderID, &r.OperationType, &r.OperatorID,
			&r.OldProcessStatus, &r.NewProcessStatus,
			&r.OldWorkOrderStatus, &r.NewWorkOrderStatus,
			&assignedFrom, &assignedTo, &r.Reason, &r.Comment, &r.CreatedAt); err != nil {
			return nil, err
		}
		if assignedFrom.Valid {
			r.AssignedFrom = &assignedFrom.Int64
		}
		if assignedTo.Valid {
			r.AssignedTo = &assignedTo.Int64
		}
		records = append(records, r)
	}
	return records, nil
}

func (db *DB) AddCommunicationRecord(ctx context.Context, record *models.CommunicationRecord) error {
	query := `INSERT INTO communication_records (work_order_id, sender_id, content) VALUES (?, ?, ?)`
	_, err := db.ExecContext(ctx, query, record.WorkOrderID, record.SenderID, record.Content)
	return err
}

func (db *DB) GetWorkOrderCommunications(ctx context.Context, workOrderID int64) ([]models.CommunicationRecord, error) {
	var records []models.CommunicationRecord
	query := `SELECT id, work_order_id, sender_id, content, created_at FROM communication_records WHERE work_order_id = ? ORDER BY created_at DESC`
	rows, err := db.QueryContext(ctx, query, workOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.CommunicationRecord
		if err := rows.Scan(&r.ID, &r.WorkOrderID, &r.SenderID, &r.Content, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func (db *DB) ListWorkOrdersByResource(ctx context.Context, resourceID int64) ([]models.WorkOrder, error) {
	var orders []models.WorkOrder
	query := `SELECT wo.id, wo.type, wo.title, wo.description, wo.content, wo.current_handler_id, wo.current_team_id, wo.process_status, wo.work_order_status, wo.submitter_id, wo.escalated, wo.rating, wo.escalated_at, wo.deadline_at, wo.process_start_at, wo.created_at, wo.updated_at FROM work_orders wo JOIN work_order_resources wor ON wo.id = wor.work_order_id WHERE wor.resource_id = ? ORDER BY wo.created_at DESC`
	rows, err := db.QueryContext(ctx, query, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var wo models.WorkOrder
		var currentHandlerID sql.NullInt64
		var escalated sql.NullInt64
		var rating sql.NullInt64
		var escalatedAt sql.NullTime
		var deadlineAt sql.NullTime
		var processStartAt sql.NullTime

		if err := rows.Scan(&wo.ID, &wo.Type, &wo.Title, &wo.Description, &wo.Content,
			&currentHandlerID, &wo.CurrentTeamID, &wo.ProcessStatus, &wo.WorkOrderStatus,
			&wo.SubmitterID, &escalated, &rating, &escalatedAt, &deadlineAt, &processStartAt,
			&wo.CreatedAt, &wo.UpdatedAt); err != nil {
			return nil, err
		}

		if currentHandlerID.Valid {
			wo.CurrentHandlerID = &currentHandlerID.Int64
		}
		wo.Escalated = escalated.Int64 > 0
		if rating.Valid {
			r := int(rating.Int64)
			wo.Rating = &r
		}
		wo.EscalatedAt = nullableTime(escalatedAt)
		wo.DeadlineAt = nullableTime(deadlineAt)
		wo.ProcessStartAt = nullableTime(processStartAt)
		orders = append(orders, wo)
	}
	return orders, nil
}

func (db *DB) GetOverdueWorkOrders(ctx context.Context, now time.Time) ([]models.WorkOrder, error) {
	var orders []models.WorkOrder
	query := `SELECT id, type, title, description, content, current_handler_id, current_team_id, process_status, work_order_status, submitter_id, escalated, rating, escalated_at, deadline_at, process_start_at, created_at, updated_at FROM work_orders WHERE escalated = 0 AND deadline_at IS NOT NULL AND deadline_at < ? AND work_order_status NOT IN ('待确认', '已关闭') ORDER BY deadline_at ASC`
	rows, err := db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var wo models.WorkOrder
		var currentHandlerID sql.NullInt64
		var escalated sql.NullInt64
		var rating sql.NullInt64
		var escalatedAt sql.NullTime
		var deadlineAt sql.NullTime
		var processStartAt sql.NullTime

		if err := rows.Scan(&wo.ID, &wo.Type, &wo.Title, &wo.Description, &wo.Content,
			&currentHandlerID, &wo.CurrentTeamID, &wo.ProcessStatus, &wo.WorkOrderStatus,
			&wo.SubmitterID, &escalated, &rating, &escalatedAt, &deadlineAt, &processStartAt,
			&wo.CreatedAt, &wo.UpdatedAt); err != nil {
			return nil, err
		}

		if currentHandlerID.Valid {
			wo.CurrentHandlerID = &currentHandlerID.Int64
		}
		wo.Escalated = escalated.Int64 > 0
		if rating.Valid {
			r := int(rating.Int64)
			wo.Rating = &r
		}
		wo.EscalatedAt = nullableTime(escalatedAt)
		wo.DeadlineAt = nullableTime(deadlineAt)
		wo.ProcessStartAt = nullableTime(processStartAt)
		orders = append(orders, wo)
	}
	return orders, nil
}

func (db *DB) ListWorkOrders(ctx context.Context) ([]models.WorkOrder, error) {
	var orders []models.WorkOrder
	query := `SELECT id, type, title, description, content, current_handler_id, current_team_id, process_status, work_order_status, submitter_id, escalated, rating, escalated_at, deadline_at, process_start_at, created_at, updated_at FROM work_orders ORDER BY created_at DESC`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var wo models.WorkOrder
		var currentHandlerID sql.NullInt64
		var escalated sql.NullInt64
		var rating sql.NullInt64
		var escalatedAt sql.NullTime
		var deadlineAt sql.NullTime
		var processStartAt sql.NullTime

		if err := rows.Scan(&wo.ID, &wo.Type, &wo.Title, &wo.Description, &wo.Content,
			&currentHandlerID, &wo.CurrentTeamID, &wo.ProcessStatus, &wo.WorkOrderStatus,
			&wo.SubmitterID, &escalated, &rating, &escalatedAt, &deadlineAt, &processStartAt,
			&wo.CreatedAt, &wo.UpdatedAt); err != nil {
			return nil, err
		}

		if currentHandlerID.Valid {
			wo.CurrentHandlerID = &currentHandlerID.Int64
		}
		wo.Escalated = escalated.Int64 > 0
		if rating.Valid {
			r := int(rating.Int64)
			wo.Rating = &r
		}
		wo.EscalatedAt = nullableTime(escalatedAt)
		wo.DeadlineAt = nullableTime(deadlineAt)
		wo.ProcessStartAt = nullableTime(processStartAt)
		orders = append(orders, wo)
	}
	return orders, nil
}

func (db *DB) CreateDutySchedule(ctx context.Context, schedule *models.DutySchedule) error {
	query := `INSERT INTO duty_schedules (team_id, user_id, duty_date) VALUES (?, ?, ?)`
	_, err := db.ExecContext(ctx, query, schedule.TeamID, schedule.UserID, schedule.DutyDate)
	return err
}

func (db *DB) CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	query := `INSERT INTO teams (name, type, leader_id) VALUES (?, ?, ?)`
	res, err := db.ExecContext(ctx, query, team.Name, string(team.Type), team.LeaderID)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return db.GetTeamByID(ctx, id)
}

func (db *DB) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := `INSERT INTO users (username, email, phone, team_id) VALUES (?, ?, ?, ?)`
	res, err := db.ExecContext(ctx, query, user.Username, user.Email, user.Phone, user.TeamID)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return db.GetUserByID(ctx, id)
}

func (db *DB) CreateResourceType(ctx context.Context, rt *models.ResourceType) (*models.ResourceType, error) {
	query := `INSERT INTO resource_types (name, description) VALUES (?, ?)`
	res, err := db.ExecContext(ctx, query, rt.Name, rt.Description)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return db.GetResourceTypeByID(ctx, id)
}

func (db *DB) CreateResource(ctx context.Context, r *models.Resource) (*models.Resource, error) {
	query := `INSERT INTO resources (resource_type_id, name, description) VALUES (?, ?, ?)`
	res, err := db.ExecContext(ctx, query, r.ResourceTypeID, r.Name, r.Description)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return db.GetResourceByID(ctx, id)
}
