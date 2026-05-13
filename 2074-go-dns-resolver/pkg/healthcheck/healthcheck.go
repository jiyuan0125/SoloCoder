package healthcheck

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type HealthCheckConfig struct {
	Domain          string
	IntervalMinutes int
	Enabled         bool
}

type HealthCheckStatus struct {
	Domain       string    `json:"domain"`
	LastIP       string    `json:"last_ip"`
	LastCheckAt  time.Time `json:"last_check_at"`
	NextCheckAt  time.Time `json:"next_check_at"`
	IntervalMin  int       `json:"interval_minutes"`
	Enabled      bool      `json:"enabled"`
	HasChanged   bool      `json:"has_changed"`
}

type Notification struct {
	Domain      string    `json:"domain"`
	OldIP       string    `json:"old_ip"`
	NewIP       string    `json:"new_ip"`
	ChangedAt   time.Time `json:"changed_at"`
	Message     string    `json:"message"`
}

type DNSQuerier interface {
	QueryARecords(ctx context.Context, domain string) ([]string, error)
}

type HealthChecker struct {
	db       *sql.DB
	querier  DNSQuerier
	notifyCh chan Notification
	mu       sync.RWMutex
	checks   map[string]*checkTask
	stopCh   chan struct{}
}

type checkTask struct {
	domain   string
	interval time.Duration
	ticker   *time.Ticker
	stop     chan struct{}
}

func NewHealthChecker(dbPath string, querier DNSQuerier) (*HealthChecker, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	hc := &HealthChecker{
		db:       db,
		querier:  querier,
		notifyCh: make(chan Notification, 100),
		checks:   make(map[string]*checkTask),
		stopCh:   make(chan struct{}),
	}

	if err := hc.loadExistingChecks(); err != nil {
		return nil, err
	}

	return hc, nil
}

func (hc *HealthChecker) loadExistingChecks() error {
	query := `SELECT domain, interval_seconds, enabled FROM health_checks WHERE enabled = 1`
	rows, err := hc.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var domain string
		var intervalSec int
		var enabled bool
		if err := rows.Scan(&domain, &intervalSec, &enabled); err != nil {
			return err
		}

		interval := time.Duration(intervalSec) * time.Second
		if interval < time.Minute {
			interval = time.Minute
		}

		hc.startCheck(domain, interval)
	}

	return nil
}

func (hc *HealthChecker) AddCheck(config HealthCheckConfig) error {
	if config.IntervalMinutes < 1 {
		config.IntervalMinutes = 1
	}

	interval := time.Duration(config.IntervalMinutes) * time.Minute

	hc.mu.Lock()
	defer hc.mu.Unlock()

	if _, exists := hc.checks[config.Domain]; exists {
		return fmt.Errorf("health check for domain %s already exists", config.Domain)
	}

	query := `
	INSERT INTO health_checks (domain, interval_seconds, enabled, next_check_at)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(domain) DO UPDATE SET
		interval_seconds = excluded.interval_seconds,
		enabled = excluded.enabled,
		next_check_at = excluded.next_check_at
	`

	_, err := hc.db.Exec(
		query,
		config.Domain,
		int64(interval.Seconds()),
		config.Enabled,
		time.Now().Add(interval),
	)
	if err != nil {
		return err
	}

	if config.Enabled {
		hc.startCheck(config.Domain, interval)
	}

	return nil
}

func (hc *HealthChecker) RemoveCheck(domain string) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if task, exists := hc.checks[domain]; exists {
		close(task.stop)
		task.ticker.Stop()
		delete(hc.checks, domain)
	}

	query := `DELETE FROM health_checks WHERE domain = ?`
	_, err := hc.db.Exec(query, domain)
	return err
}

func (hc *HealthChecker) startCheck(domain string, interval time.Duration) {
	if _, exists := hc.checks[domain]; exists {
		return
	}

	task := &checkTask{
		domain:   domain,
		interval: interval,
		ticker:   time.NewTicker(interval),
		stop:     make(chan struct{}),
	}

	hc.checks[domain] = task

	go hc.runCheck(task)
}

func (hc *HealthChecker) runCheck(task *checkTask) {
	log.Printf("Starting health check for %s (interval: %v)", task.domain, task.interval)

	if err := hc.performCheck(task.domain); err != nil {
		log.Printf("Initial check failed for %s: %v", task.domain, err)
	}

	for {
		select {
		case <-task.ticker.C:
			if err := hc.performCheck(task.domain); err != nil {
				log.Printf("Health check failed for %s: %v", task.domain, err)
			}
		case <-task.stop:
			log.Printf("Stopping health check for %s", task.domain)
			return
		}
	}
}

func (hc *HealthChecker) performCheck(domain string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ips, err := hc.querier.QueryARecords(ctx, domain)
	if err != nil {
		return err
	}

	currentIP := strings.Join(ips, ",")

	var lastIP string
	query := `SELECT last_ip FROM health_checks WHERE domain = ?`
	err = hc.db.QueryRow(query, domain).Scan(&lastIP)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	hasChanged := lastIP != "" && lastIP != currentIP

	if hasChanged {
		notification := Notification{
			Domain:    domain,
			OldIP:     lastIP,
			NewIP:     currentIP,
			ChangedAt: time.Now(),
			Message:   fmt.Sprintf("DNS A record for %s changed: %s -> %s", domain, lastIP, currentIP),
		}
		select {
		case hc.notifyCh <- notification:
		default:
			log.Printf("Notification channel full, dropping notification for %s", domain)
		}
		log.Printf("NOTIFICATION: %s", notification.Message)
	}

	updateQuery := `
	UPDATE health_checks SET
		last_ip = ?,
		last_check_at = ?,
		next_check_at = ?
	WHERE domain = ?
	`

	now := time.Now()
	_, err = hc.db.Exec(updateQuery, currentIP, now, now.Add(time.Minute), domain)
	return err
}

func (hc *HealthChecker) GetStatus(domain string) (*HealthCheckStatus, error) {
	query := `
	SELECT domain, last_ip, last_check_at, next_check_at, interval_seconds, enabled
	FROM health_checks WHERE domain = ?
	`

	var status HealthCheckStatus
	var intervalSec int
	var lastCheckAt, nextCheckAt sql.NullTime

	err := hc.db.QueryRow(query, domain).Scan(
		&status.Domain,
		&status.LastIP,
		&lastCheckAt,
		&nextCheckAt,
		&intervalSec,
		&status.Enabled,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	status.IntervalMin = intervalSec / 60
	if lastCheckAt.Valid {
		status.LastCheckAt = lastCheckAt.Time
	}
	if nextCheckAt.Valid {
		status.NextCheckAt = nextCheckAt.Time
	}

	return &status, nil
}

func (hc *HealthChecker) GetAllStatuses() ([]HealthCheckStatus, error) {
	query := `
	SELECT domain, last_ip, last_check_at, next_check_at, interval_seconds, enabled
	FROM health_checks
	ORDER BY domain
	`

	rows, err := hc.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []HealthCheckStatus
	for rows.Next() {
		var status HealthCheckStatus
		var intervalSec int
		var lastCheckAt, nextCheckAt sql.NullTime

		err := rows.Scan(
			&status.Domain,
			&status.LastIP,
			&lastCheckAt,
			&nextCheckAt,
			&intervalSec,
			&status.Enabled,
		)
		if err != nil {
			return nil, err
		}

		status.IntervalMin = intervalSec / 60
		if lastCheckAt.Valid {
			status.LastCheckAt = lastCheckAt.Time
		}
		if nextCheckAt.Valid {
			status.NextCheckAt = nextCheckAt.Time
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (hc *HealthChecker) Notifications() <-chan Notification {
	return hc.notifyCh
}

func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for _, task := range hc.checks {
		close(task.stop)
		task.ticker.Stop()
	}

	close(hc.notifyCh)
	hc.db.Close()
}
