import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'gateway.db');
const db = new Database(dbPath);

db.pragma('journal_mode = WAL');

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS services (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL UNIQUE,
      upstream_url TEXT NOT NULL,
      route_prefix TEXT NOT NULL UNIQUE,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS clients (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      description TEXT,
      created_at INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS api_keys (
      id TEXT PRIMARY KEY,
      client_id TEXT NOT NULL,
      key_hash TEXT NOT NULL UNIQUE,
      key_prefix TEXT NOT NULL,
      expires_at INTEGER,
      permissions TEXT,
      is_active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL,
      FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS key_stats (
      id TEXT PRIMARY KEY,
      key_id TEXT NOT NULL,
      total_requests INTEGER NOT NULL DEFAULT 0,
      successful_requests INTEGER NOT NULL DEFAULT 0,
      failed_requests INTEGER NOT NULL DEFAULT 0,
      last_request_at INTEGER,
      FOREIGN KEY (key_id) REFERENCES api_keys(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS rate_limit_configs (
      id TEXT PRIMARY KEY,
      service_id TEXT,
      type TEXT NOT NULL,
      identifier TEXT NOT NULL,
      per_second INTEGER,
      per_minute INTEGER,
      per_hour INTEGER,
      created_at INTEGER NOT NULL,
      UNIQUE(type, identifier)
    );

    CREATE TABLE IF NOT EXISTS request_logs (
      id TEXT PRIMARY KEY,
      method TEXT NOT NULL,
      path TEXT NOT NULL,
      status_code INTEGER NOT NULL,
      response_time_ms INTEGER NOT NULL,
      upstream_name TEXT,
      client_ip TEXT,
      key_id TEXT,
      timestamp INTEGER NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_services_prefix ON services(route_prefix);
    CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
    CREATE INDEX IF NOT EXISTS idx_api_keys_client ON api_keys(client_id);
    CREATE INDEX IF NOT EXISTS idx_request_logs_time ON request_logs(timestamp);
    CREATE INDEX IF NOT EXISTS idx_key_stats_key ON key_stats(key_id);
  `);
}

export { db };
