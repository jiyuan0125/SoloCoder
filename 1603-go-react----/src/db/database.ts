import Database, { Database as DatabaseType } from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'data', 'app.db');

let db: DatabaseType | null = null;

export const initDb = (): void => {
  if (!db) {
    db = new Database(dbPath);
    db.pragma('journal_mode = WAL');
  }

  db.exec(`
    CREATE TABLE IF NOT EXISTS metrics (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL UNIQUE,
      description TEXT,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS metric_data (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      metric_id INTEGER NOT NULL,
      value REAL NOT NULL,
      timestamp DATETIME NOT NULL,
      FOREIGN KEY (metric_id) REFERENCES metrics(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_metric_data_metric ON metric_data(metric_id);
    CREATE INDEX IF NOT EXISTS idx_metric_data_timestamp ON metric_data(timestamp);

    CREATE TABLE IF NOT EXISTS dashboards (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      description TEXT,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS chart_cards (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      dashboard_id INTEGER NOT NULL,
      title TEXT NOT NULL,
      chart_type TEXT NOT NULL CHECK (chart_type IN ('line', 'bar', 'pie', 'heatmap')),
      metric_name TEXT NOT NULL,
      aggregation TEXT NOT NULL CHECK (aggregation IN ('sum', 'avg', 'max', 'min', 'count', 'p50', 'p90', 'p99')),
      position TEXT,
      data_status TEXT NOT NULL DEFAULT 'available' CHECK (data_status IN ('available', 'missing')),
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (dashboard_id) REFERENCES dashboards(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_chart_cards_dashboard ON chart_cards(dashboard_id);
    CREATE INDEX IF NOT EXISTS idx_chart_cards_metric ON chart_cards(metric_name);
  `);
};

const getDb = (): DatabaseType => {
  if (!db) {
    throw new Error('数据库未初始化，请先调用 initDb()');
  }
  return db;
};

export default getDb;
