import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'metrics.db');
const db = new Database(dbPath);

db.pragma('journal_mode = WAL');

db.exec(`
  CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    tags TEXT NOT NULL,
    UNIQUE(name, tags)
  );

  CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);

  CREATE TABLE IF NOT EXISTS metric_points (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_id INTEGER NOT NULL,
    timestamp INTEGER NOT NULL,
    value REAL NOT NULL,
    FOREIGN KEY (metric_id) REFERENCES metrics(id) ON DELETE CASCADE,
    UNIQUE(metric_id, timestamp)
  );

  CREATE INDEX IF NOT EXISTS idx_metric_points_time ON metric_points(metric_id, timestamp);

  CREATE TABLE IF NOT EXISTS templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    queries TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS template_shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE,
    UNIQUE(template_id, user_id)
  );
`);

export default db;
