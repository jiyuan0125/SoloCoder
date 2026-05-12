import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'error-tracking.db');
const db = new Database(dbPath);

db.exec(`
  PRAGMA journal_mode = WAL;
  PRAGMA foreign_keys = ON;

  CREATE TABLE IF NOT EXISTS error_groups (
    id TEXT PRIMARY KEY,
    service_name TEXT NOT NULL,
    error_type TEXT NOT NULL,
    error_message TEXT NOT NULL,
    top_frame TEXT,
    status TEXT NOT NULL DEFAULT 'unhandled',
    occurrence_count INTEGER NOT NULL DEFAULT 1,
    last_occurred_at INTEGER NOT NULL,
    first_occurred_at INTEGER NOT NULL,
    is_high_frequency INTEGER NOT NULL DEFAULT 0,
    affected_users INTEGER NOT NULL DEFAULT 0,
    created_user_id_set TEXT NOT NULL DEFAULT '[]',
    hash TEXT NOT NULL UNIQUE
  );

  CREATE TABLE IF NOT EXISTS error_events (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    service_name TEXT NOT NULL,
    error_type TEXT NOT NULL,
    error_message TEXT NOT NULL,
    stack_trace TEXT,
    top_frame TEXT,
    occurred_at INTEGER NOT NULL,
    context TEXT,
    user_id TEXT,
    FOREIGN KEY (group_id) REFERENCES error_groups(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY,
    version TEXT NOT NULL,
    deployed_at INTEGER NOT NULL
  );

  CREATE INDEX IF NOT EXISTS idx_error_events_group_id ON error_events(group_id);
  CREATE INDEX IF NOT EXISTS idx_error_events_occurred_at ON error_events(occurred_at);
  CREATE INDEX IF NOT EXISTS idx_error_groups_status ON error_groups(status);
  CREATE INDEX IF NOT EXISTS idx_error_groups_last_occurred ON error_groups(last_occurred_at);
`);

export default db;
