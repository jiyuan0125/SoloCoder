package repository

import (
	"billing-service/model"
	"database/sql"
	"strconv"
	"time"
)

type ConfigRepository struct {
	db *Database
}

func NewConfigRepository(db *Database) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) Get(key string) (*model.Config, error) {
	query := `
		SELECT id, key, value, description, created_at, updated_at
		FROM configs
		WHERE key = ?
	`

	var config model.Config
	err := r.db.QueryRow(query, key).Scan(
		&config.ID,
		&config.Key,
		&config.Value,
		&config.Description,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *ConfigRepository) Set(key string, value string, description string) error {
	existing, err := r.Get(key)
	if err != nil {
		return err
	}

	now := time.Now()

	if existing == nil {
		query := `
			INSERT INTO configs (key, value, description, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`
		_, err := r.db.Exec(query, key, value, description, now, now)
		return err
	}

	query := `
		UPDATE configs
		SET value = ?, updated_at = ?
		WHERE key = ?
	`
	_, err = r.db.Exec(query, value, now, key)
	return err
}

func (r *ConfigRepository) GetSmsUnitPrice() (float64, error) {
	config, err := r.Get("sms.unit_price")
	if err != nil {
		return 0, err
	}

	if config == nil {
		return 0.1, nil
	}

	price, err := strconv.ParseFloat(config.Value, 64)
	if err != nil {
		return 0.1, nil
	}

	return price, nil
}

func (r *ConfigRepository) GetStorageUnitPrice() (float64, error) {
	config, err := r.Get("storage.unit_price")
	if err != nil {
		return 0, err
	}

	if config == nil {
		return 5.0, nil
	}

	price, err := strconv.ParseFloat(config.Value, 64)
	if err != nil {
		return 5.0, nil
	}

	return price, nil
}

func (r *ConfigRepository) GetPaymentDueDays() (int, error) {
	config, err := r.Get("payment.due_days")
	if err != nil {
		return 0, err
	}

	if config == nil {
		return 15, nil
	}

	days, err := strconv.Atoi(config.Value)
	if err != nil {
		return 15, nil
	}

	return days, nil
}

func (r *ConfigRepository) GetSevereOverdueDays() (int, error) {
	config, err := r.Get("overdue.severe_days")
	if err != nil {
		return 0, err
	}

	if config == nil {
		return 60, nil
	}

	days, err := strconv.Atoi(config.Value)
	if err != nil {
		return 60, nil
	}

	return days, nil
}
