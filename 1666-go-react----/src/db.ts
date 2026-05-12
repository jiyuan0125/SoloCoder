import Database, { Database as DatabaseType } from 'better-sqlite3';
import path from 'path';

const DB_PATH = process.env.DB_PATH || path.join(process.cwd(), 'data', 'sales_funnel.db');

const db: DatabaseType = new Database(DB_PATH);
db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

export { db };

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS salespeople (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      team_id TEXT NOT NULL,
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    );

    CREATE TABLE IF NOT EXISTS opportunities (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      customer_id TEXT NOT NULL,
      customer_name TEXT NOT NULL,
      amount REAL NOT NULL,
      salesperson_id TEXT NOT NULL,
      current_stage TEXT NOT NULL,
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now')),
      closed_at TEXT,
      loss_reason TEXT,
      FOREIGN KEY (salesperson_id) REFERENCES salespeople(id)
    );

    CREATE TABLE IF NOT EXISTS stage_history (
      id TEXT PRIMARY KEY,
      opportunity_id TEXT NOT NULL,
      from_stage TEXT,
      to_stage TEXT NOT NULL,
      reason TEXT,
      timestamp TEXT NOT NULL DEFAULT (datetime('now')),
      FOREIGN KEY (opportunity_id) REFERENCES opportunities(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_opportunities_salesperson ON opportunities(salesperson_id);
    CREATE INDEX IF NOT EXISTS idx_opportunities_stage ON opportunities(current_stage);
    CREATE INDEX IF NOT EXISTS idx_history_opportunity ON stage_history(opportunity_id);
    CREATE INDEX IF NOT EXISTS idx_history_timestamp ON stage_history(timestamp);
    CREATE INDEX IF NOT EXISTS idx_salespeople_team ON salespeople(team_id);
  `);
}

export type DbTransaction = ReturnType<DatabaseType['transaction']>;
