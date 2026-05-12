import Database from 'better-sqlite3';
import path from 'path';

const DB_PATH = path.join(process.cwd(), 'feature-switches.db');

let dbInstance: Database.Database | null = null;

export function getDatabase(): Database.Database {
  if (!dbInstance) {
    dbInstance = new Database(DB_PATH);
    dbInstance.pragma('journal_mode = WAL');
    dbInstance.pragma('foreign_keys = ON');
    initializeDatabase(dbInstance);
  }
  return dbInstance;
}

function initializeDatabase(db: Database.Database): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS switches (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      key TEXT UNIQUE NOT NULL,
      name TEXT NOT NULL,
      description TEXT DEFAULT '',
      value_type TEXT NOT NULL CHECK(value_type IN ('boolean', 'json')),
      boolean_value INTEGER NOT NULL DEFAULT 0,
      json_value TEXT,
      conditions TEXT NOT NULL DEFAULT '{}',
      is_deleted INTEGER NOT NULL DEFAULT 0,
      created_by TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      updated_by TEXT NOT NULL,
      updated_at INTEGER NOT NULL,
      version INTEGER NOT NULL DEFAULT 1
    );

    CREATE INDEX IF NOT EXISTS idx_switches_key ON switches(key);
    CREATE INDEX IF NOT EXISTS idx_switches_updated ON switches(updated_at);
    CREATE INDEX IF NOT EXISTS idx_switches_version ON switches(version);
    CREATE INDEX IF NOT EXISTS idx_switches_deleted ON switches(is_deleted);

    CREATE TABLE IF NOT EXISTS switch_history (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      switch_key TEXT NOT NULL,
      operation TEXT NOT NULL CHECK(operation IN ('create', 'update', 'delete')),
      old_value TEXT,
      new_value TEXT,
      operator TEXT NOT NULL,
      timestamp INTEGER NOT NULL,
      version INTEGER NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_history_key ON switch_history(switch_key);
    CREATE INDEX IF NOT EXISTS idx_history_timestamp ON switch_history(timestamp);

    CREATE TABLE IF NOT EXISTS service_references (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      service_name TEXT NOT NULL,
      switch_key TEXT NOT NULL,
      is_active INTEGER NOT NULL DEFAULT 1,
      created_at INTEGER NOT NULL,
      last_poll_at INTEGER NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_refs_service ON service_references(service_name);
    CREATE INDEX IF NOT EXISTS idx_refs_switch ON service_references(switch_key);
    CREATE INDEX IF NOT EXISTS idx_refs_active ON service_references(is_active);
    CREATE UNIQUE INDEX IF NOT EXISTS idx_refs_unique ON service_references(service_name, switch_key);

    CREATE TABLE IF NOT EXISTS sync_state (
      id INTEGER PRIMARY KEY CHECK(id = 1),
      last_version INTEGER NOT NULL DEFAULT 0,
      updated_at INTEGER NOT NULL
    );
  `);

  db.prepare(`INSERT OR IGNORE INTO sync_state (id, last_version, updated_at) VALUES (1, 0, ?)`).run(Date.now());
}

export function closeDatabase(): void {
  if (dbInstance) {
    dbInstance.close();
    dbInstance = null;
  }
}
