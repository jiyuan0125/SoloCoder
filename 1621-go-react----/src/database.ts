import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'backup-service.db');
const db = new Database(dbPath) as any;

db.exec(`
  CREATE TABLE IF NOT EXISTS backup_tasks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    cron_expression TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    data_range TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS snapshots (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL,
    data_range TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    status TEXT NOT NULL,
    sha256 TEXT,
    file_path TEXT NOT NULL,
    error_reason TEXT
  );

  CREATE TABLE IF NOT EXISTS restore_operations (
    id TEXT PRIMARY KEY,
    snapshot_id TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT,
    status TEXT NOT NULL,
    error_reason TEXT,
    FOREIGN KEY (snapshot_id) REFERENCES snapshots(id)
  );

  CREATE INDEX IF NOT EXISTS idx_snapshots_created_at ON snapshots(created_at);
  CREATE INDEX IF NOT EXISTS idx_snapshots_status ON snapshots(status);
  CREATE INDEX IF NOT EXISTS idx_restore_snapshot_id ON restore_operations(snapshot_id);
`);

export default db;
