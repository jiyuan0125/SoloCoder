import Database, { Database as DatabaseType } from 'better-sqlite3';
import path from 'path';

const db: DatabaseType = new Database(path.join(process.cwd(), 'abtesting.db'));

db.pragma('journal_mode = WAL');

const initDb = (): void => {
  db.exec(`
    CREATE TABLE IF NOT EXISTS experiments (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      status TEXT NOT NULL,
      traffic_type TEXT NOT NULL,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS variants (
      id TEXT PRIMARY KEY,
      experiment_id TEXT NOT NULL,
      name TEXT NOT NULL,
      traffic_percentage REAL NOT NULL,
      is_control INTEGER NOT NULL,
      FOREIGN KEY (experiment_id) REFERENCES experiments(id)
    );

    CREATE TABLE IF NOT EXISTS metrics (
      id TEXT PRIMARY KEY,
      experiment_id TEXT NOT NULL,
      name TEXT NOT NULL,
      is_core INTEGER NOT NULL,
      FOREIGN KEY (experiment_id) REFERENCES experiments(id)
    );

    CREATE TABLE IF NOT EXISTS metric_data (
      id TEXT PRIMARY KEY,
      experiment_id TEXT NOT NULL,
      variant_id TEXT NOT NULL,
      metric_id TEXT NOT NULL,
      user_key TEXT NOT NULL,
      value REAL NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (experiment_id) REFERENCES experiments(id),
      FOREIGN KEY (variant_id) REFERENCES variants(id),
      FOREIGN KEY (metric_id) REFERENCES metrics(id)
    );

    CREATE TABLE IF NOT EXISTS assignments (
      id TEXT PRIMARY KEY,
      experiment_id TEXT NOT NULL,
      user_key TEXT NOT NULL,
      variant_id TEXT NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (experiment_id) REFERENCES experiments(id),
      FOREIGN KEY (variant_id) REFERENCES variants(id),
      UNIQUE(experiment_id, user_key)
    );

    CREATE TABLE IF NOT EXISTS grayscale_configs (
      id TEXT PRIMARY KEY,
      experiment_id TEXT NOT NULL,
      steps TEXT NOT NULL,
      current_step_index INTEGER NOT NULL,
      is_active INTEGER NOT NULL,
      FOREIGN KEY (experiment_id) REFERENCES experiments(id),
      UNIQUE(experiment_id)
    );

    CREATE INDEX IF NOT EXISTS idx_assignments_experiment_user ON assignments(experiment_id, user_key);
    CREATE INDEX IF NOT EXISTS idx_metric_data_experiment ON metric_data(experiment_id);
  `);
};

initDb();

export default db;
