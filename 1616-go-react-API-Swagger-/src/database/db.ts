import Database from 'better-sqlite3';
import path from 'path';

const dbPath = process.env.DB_PATH || path.join(__dirname, '../../data/api-platform.db');

const db = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

const initDb = (): void => {
  db.exec(`
    CREATE TABLE IF NOT EXISTS services (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      description TEXT,
      base_url TEXT,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL,
      has_swagger INTEGER NOT NULL DEFAULT 0
    );

    CREATE TABLE IF NOT EXISTS api_endpoints (
      id TEXT PRIMARY KEY,
      service_id TEXT NOT NULL,
      path TEXT NOT NULL,
      method TEXT NOT NULL,
      summary TEXT,
      description TEXT,
      tags TEXT,
      parameters TEXT,
      request_body TEXT,
      responses TEXT,
      created_at TEXT NOT NULL,
      FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
      UNIQUE(service_id, path, method)
    );

    CREATE TABLE IF NOT EXISTS debug_records (
      id TEXT PRIMARY KEY,
      service_id TEXT NOT NULL,
      api_id TEXT NOT NULL,
      request TEXT NOT NULL,
      response TEXT NOT NULL,
      status_code INTEGER NOT NULL,
      duration_ms INTEGER NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
      FOREIGN KEY (api_id) REFERENCES api_endpoints(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS config (
      key TEXT PRIMARY KEY,
      value TEXT NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_api_endpoints_service_id ON api_endpoints(service_id);
    CREATE INDEX IF NOT EXISTS idx_debug_records_service_id ON debug_records(service_id);
    CREATE INDEX IF NOT EXISTS idx_debug_records_api_id ON debug_records(api_id);
  `);
};

const getDb = (): Database.Database => db;

export { getDb, initDb };
