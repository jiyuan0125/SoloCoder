import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'data', 'artifacts.db');

export const db = new Database(dbPath);

db.pragma('journal_mode = WAL');

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS projects (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL UNIQUE,
      description TEXT,
      storage_quota_bytes INTEGER NOT NULL DEFAULT 10737418240,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS artifacts (
      id TEXT PRIMARY KEY,
      project_id TEXT NOT NULL,
      name TEXT NOT NULL,
      type TEXT NOT NULL CHECK(type IN ('docker', 'jar', 'npm', 'static')),
      version TEXT NOT NULL,
      storage_path TEXT NOT NULL,
      sha256 TEXT NOT NULL,
      size_bytes INTEGER NOT NULL,
      download_count INTEGER NOT NULL DEFAULT 0,
      is_latest INTEGER NOT NULL DEFAULT 0,
      is_snapshot INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
      UNIQUE(project_id, name, type, version)
    );

    CREATE INDEX IF NOT EXISTS idx_artifacts_project ON artifacts(project_id);
    CREATE INDEX IF NOT EXISTS idx_artifacts_latest ON artifacts(is_latest);
    CREATE INDEX IF NOT EXISTS idx_artifacts_project_name_type ON artifacts(project_id, name, type);
  `);
}

export const DEFAULT_QUOTA_BYTES = 10 * 1024 * 1024 * 1024;
export const MAX_SNAPSHOT_VERSIONS = 10;
