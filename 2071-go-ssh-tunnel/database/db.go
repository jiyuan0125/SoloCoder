package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"ssh-tunnel-manager/models"
)

type DB struct {
	*sql.DB
}

func Init(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "tunnels.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &DB{db}, nil
}

func initSchema(db *sql.DB) error {
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

func (db *DB) CreateTunnelConfig(config *models.TunnelConfig) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tunnel_configs 
		(name, type, ssh_server, ssh_user, auth_type, auth_data, local_port, remote_host, remote_port, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, config.Name, config.Type, config.SSHServer, config.SSHUser, config.AuthType, config.AuthData,
		config.LocalPort, config.RemoteHost, config.RemotePort, time.Now(), time.Now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) GetTunnelConfig(id int64) (*models.TunnelConfig, error) {
	config := &models.TunnelConfig{}
	err := db.QueryRow(`
		SELECT id, name, type, ssh_server, ssh_user, auth_type, auth_data, local_port, remote_host, remote_port, created_at, updated_at
		FROM tunnel_configs WHERE id = ?
	`, id).Scan(&config.ID, &config.Name, &config.Type, &config.SSHServer, &config.SSHUser,
		&config.AuthType, &config.AuthData, &config.LocalPort, &config.RemoteHost, &config.RemotePort,
		&config.CreatedAt, &config.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return config, err
}

func (db *DB) ListTunnelConfigs() ([]*models.TunnelConfig, error) {
	rows, err := db.Query(`
		SELECT id, name, type, ssh_server, ssh_user, auth_type, auth_data, local_port, remote_host, remote_port, created_at, updated_at
		FROM tunnel_configs ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.TunnelConfig
	for rows.Next() {
		config := &models.TunnelConfig{}
		err := rows.Scan(&config.ID, &config.Name, &config.Type, &config.SSHServer, &config.SSHUser,
			&config.AuthType, &config.AuthData, &config.LocalPort, &config.RemoteHost, &config.RemotePort,
			&config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, rows.Err()
}

func (db *DB) UpdateTunnelConfig(config *models.TunnelConfig) error {
	_, err := db.Exec(`
		UPDATE tunnel_configs SET
		name = ?, type = ?, ssh_server = ?, ssh_user = ?, auth_type = ?, auth_data = ?, 
		local_port = ?, remote_host = ?, remote_port = ?, updated_at = ?
		WHERE id = ?
	`, config.Name, config.Type, config.SSHServer, config.SSHUser, config.AuthType, config.AuthData,
		config.LocalPort, config.RemoteHost, config.RemotePort, time.Now(), config.ID)
	return err
}

func (db *DB) DeleteTunnelConfig(id int64) error {
	_, err := db.Exec("DELETE FROM tunnel_configs WHERE id = ?", id)
	return err
}

func (db *DB) CreateTunnelState(state *models.TunnelState) error {
	_, err := db.Exec(`
		INSERT INTO tunnel_states 
		(tunnel_id, status, bytes_up, bytes_down, reconnect_count, last_connected_at, last_disconnected_at, last_disconnect_reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, state.TunnelID, state.Status, state.BytesUp, state.BytesDown, state.ReconnectCount,
		state.LastConnectedAt, state.LastDisconnectedAt, state.LastDisconnectReason,
		time.Now(), time.Now())
	return err
}

func (db *DB) GetTunnelState(tunnelID int64) (*models.TunnelState, error) {
	state := &models.TunnelState{}
	err := db.QueryRow(`
		SELECT id, tunnel_id, status, bytes_up, bytes_down, reconnect_count, last_connected_at, last_disconnected_at, last_disconnect_reason, created_at, updated_at
		FROM tunnel_states WHERE tunnel_id = ?
	`, tunnelID).Scan(&state.ID, &state.TunnelID, &state.Status, &state.BytesUp, &state.BytesDown,
		&state.ReconnectCount, &state.LastConnectedAt, &state.LastDisconnectedAt,
		&state.LastDisconnectReason, &state.CreatedAt, &state.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return state, err
}

func (db *DB) UpdateTunnelState(state *models.TunnelState) error {
	_, err := db.Exec(`
		UPDATE tunnel_states SET
		status = ?, bytes_up = ?, bytes_down = ?, reconnect_count = ?, 
		last_connected_at = ?, last_disconnected_at = ?, last_disconnect_reason = ?, updated_at = ?
		WHERE tunnel_id = ?
	`, state.Status, state.BytesUp, state.BytesDown, state.ReconnectCount,
		state.LastConnectedAt, state.LastDisconnectedAt, state.LastDisconnectReason,
		time.Now(), state.TunnelID)
	return err
}

func (db *DB) UpsertTunnelState(state *models.TunnelState) error {
	existing, err := db.GetTunnelState(state.TunnelID)
	if err != nil {
		return err
	}
	if existing == nil {
		return db.CreateTunnelState(state)
	}
	return db.UpdateTunnelState(state)
}

func (db *DB) CreateConnectionLog(log *models.ConnectionLog) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO connection_logs 
		(tunnel_id, connected_at, disconnected_at, disconnect_reason, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, log.TunnelID, log.ConnectedAt, log.DisconnectedAt, log.DisconnectReason, time.Now())
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func (db *DB) UpdateConnectionLog(id int64, disconnectedAt time.Time, reason string) error {
	_, err := db.Exec(`
		UPDATE connection_logs SET disconnected_at = ?, disconnect_reason = ? WHERE id = ?
	`, disconnectedAt, reason, id)
	return err
}

func (db *DB) ListConnectionLogs(tunnelID int64) ([]*models.ConnectionLog, error) {
	rows, err := db.Query(`
		SELECT id, tunnel_id, connected_at, disconnected_at, disconnect_reason, created_at
		FROM connection_logs WHERE tunnel_id = ? ORDER BY created_at DESC
	`, tunnelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.ConnectionLog
	for rows.Next() {
		log := &models.ConnectionLog{}
		err := rows.Scan(&log.ID, &log.TunnelID, &log.ConnectedAt, &log.DisconnectedAt,
			&log.DisconnectReason, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (db *DB) CreateTunnelRecord(record *models.TunnelRecord) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO tunnel_records 
		(tunnel_config_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, record.TunnelConfigID, record.Status, time.Now(), time.Now())
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func (db *DB) GetTunnelRecord(id int64) (*models.TunnelRecord, error) {
	record := &models.TunnelRecord{}
	err := db.QueryRow(`
		SELECT id, tunnel_config_id, status, created_at, updated_at
		FROM tunnel_records WHERE id = ?
	`, id).Scan(&record.ID, &record.TunnelConfigID, &record.Status, &record.CreatedAt, &record.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return record, err
}

func (db *DB) ListTunnelRecords() ([]*models.TunnelRecord, error) {
	rows, err := db.Query(`
		SELECT id, tunnel_config_id, status, created_at, updated_at
		FROM tunnel_records ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.TunnelRecord
	for rows.Next() {
		record := &models.TunnelRecord{}
		err := rows.Scan(&record.ID, &record.TunnelConfigID, &record.Status,
			&record.CreatedAt, &record.UpdatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (db *DB) UpdateTunnelRecord(record *models.TunnelRecord) error {
	_, err := db.Exec(`
		UPDATE tunnel_records SET status = ?, updated_at = ? WHERE id = ?
	`, record.Status, time.Now(), record.ID)
	return err
}

func (db *DB) CreateOperationHistory(history *models.OperationHistory) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO operation_histories 
		(record_id, op_type, operator, description, note, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, history.RecordID, history.OpType, history.Operator, history.Description, history.Note, time.Now())
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func (db *DB) ListOperationHistories(recordID int64) ([]*models.OperationHistory, error) {
	rows, err := db.Query(`
		SELECT id, record_id, op_type, operator, description, note, created_at
		FROM operation_histories WHERE record_id = ? ORDER BY created_at DESC
	`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*models.OperationHistory
	for rows.Next() {
		history := &models.OperationHistory{}
		err := rows.Scan(&history.ID, &history.RecordID, &history.OpType, &history.Operator,
			&history.Description, &history.Note, &history.CreatedAt)
		if err != nil {
			return nil, err
		}
		histories = append(histories, history)
	}
	return histories, rows.Err()
}

func (db *DB) UpdateOperationHistoryNote(id int64, note string) error {
	_, err := db.Exec(`
		UPDATE operation_histories SET note = ? WHERE id = ?
	`, note, id)
	return err
}

func (db *DB) GetOperationHistory(id int64) (*models.OperationHistory, error) {
	history := &models.OperationHistory{}
	err := db.QueryRow(`
		SELECT id, record_id, op_type, operator, description, note, created_at
		FROM operation_histories WHERE id = ?
	`, id).Scan(&history.ID, &history.RecordID, &history.OpType, &history.Operator,
		&history.Description, &history.Note, &history.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return history, err
}
