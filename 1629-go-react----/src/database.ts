import Database from 'better-sqlite3';
import { CircuitBreaker, CircuitBreakerConfig, RequestStats, StateTransitionLog } from './types';

const db = new Database('./circuit-breaker.db');

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS circuit_breaker_configs (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      endpoint TEXT NOT NULL,
      trigger_conditions TEXT NOT NULL,
      degradation_action TEXT NOT NULL,
      default_value TEXT,
      time_window_seconds INTEGER NOT NULL DEFAULT 60,
      open_duration_seconds INTEGER NOT NULL DEFAULT 30,
      half_open_request_limit INTEGER NOT NULL DEFAULT 5,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS circuit_breakers (
      id TEXT PRIMARY KEY,
      config_id TEXT NOT NULL,
      state TEXT NOT NULL DEFAULT 'CLOSED',
      state_changed_at TEXT NOT NULL,
      last_open_reason TEXT,
      FOREIGN KEY (config_id) REFERENCES circuit_breaker_configs(id)
    );

    CREATE TABLE IF NOT EXISTS request_stats (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      circuit_breaker_id TEXT NOT NULL,
      minute TEXT NOT NULL,
      success_count INTEGER NOT NULL DEFAULT 0,
      failure_count INTEGER NOT NULL DEFAULT 0,
      total_response_time_ms INTEGER NOT NULL DEFAULT 0,
      FOREIGN KEY (circuit_breaker_id) REFERENCES circuit_breakers(id),
      UNIQUE(circuit_breaker_id, minute)
    );

    CREATE TABLE IF NOT EXISTS state_transition_logs (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      circuit_breaker_id TEXT NOT NULL,
      from_state TEXT NOT NULL,
      to_state TEXT NOT NULL,
      reason TEXT NOT NULL,
      timestamp TEXT NOT NULL,
      FOREIGN KEY (circuit_breaker_id) REFERENCES circuit_breakers(id)
    );

    CREATE INDEX IF NOT EXISTS idx_circuit_breakers_config_id ON circuit_breakers(config_id);
    CREATE INDEX IF NOT EXISTS idx_request_stats_circuit_breaker ON request_stats(circuit_breaker_id, minute);
    CREATE INDEX IF NOT EXISTS idx_state_transition_logs_circuit_breaker ON state_transition_logs(circuit_breaker_id, timestamp);
  `);
}

export function getDatabase(): Database.Database {
  return db;
}
