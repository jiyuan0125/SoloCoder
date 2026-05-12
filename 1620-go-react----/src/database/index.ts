import type { Database } from 'better-sqlite3';
import BetterSqlite3 from 'better-sqlite3';
import path from 'path';
import { TABLES } from '../constants';

const DB_PATH = path.join(process.cwd(), 'sync.db');

const db: Database = new BetterSqlite3(DB_PATH);

db.pragma('journal_mode = WAL');

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS ${TABLES.TASKS} (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      mode TEXT NOT NULL,
      source_system TEXT NOT NULL,
      target_system TEXT NOT NULL,
      watermark INTEGER NOT NULL DEFAULT 0,
      conflict_strategy TEXT NOT NULL,
      status TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS ${TABLES.REPORTS} (
      id TEXT PRIMARY KEY,
      task_id TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      insert_count INTEGER NOT NULL DEFAULT 0,
      update_count INTEGER NOT NULL DEFAULT 0,
      delete_count INTEGER NOT NULL DEFAULT 0,
      conflict_count INTEGER NOT NULL DEFAULT 0,
      manual_intervention_items TEXT,
      details TEXT,
      FOREIGN KEY (task_id) REFERENCES ${TABLES.TASKS}(id)
    );

    CREATE TABLE IF NOT EXISTS ${TABLES.RETRY_QUEUE} (
      id TEXT PRIMARY KEY,
      task_id TEXT NOT NULL,
      data_id TEXT NOT NULL,
      data TEXT NOT NULL,
      retry_count INTEGER NOT NULL DEFAULT 0,
      max_retries INTEGER NOT NULL DEFAULT 3,
      last_error TEXT,
      status TEXT NOT NULL DEFAULT 'pending',
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL,
      FOREIGN KEY (task_id) REFERENCES ${TABLES.TASKS}(id)
    );

    CREATE INDEX IF NOT EXISTS idx_reports_task_id ON ${TABLES.REPORTS}(task_id);
    CREATE INDEX IF NOT EXISTS idx_retry_task_id ON ${TABLES.RETRY_QUEUE}(task_id);
    CREATE INDEX IF NOT EXISTS idx_retry_status ON ${TABLES.RETRY_QUEUE}(status);
  `);
}

export default db;
