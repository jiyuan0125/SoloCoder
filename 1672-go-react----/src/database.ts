import Database, { Database as DatabaseType } from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '../', 'fund-pool.db');
const db: DatabaseType = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

db.exec(`
  CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    bank TEXT NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0,
    type TEXT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'CNY',
    allow_overdraft INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS transfers (
    id TEXT PRIMARY KEY,
    from_account_id TEXT NOT NULL,
    to_account_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    reason TEXT,
    status TEXT NOT NULL,
    operator TEXT NOT NULL,
    created_at TEXT NOT NULL,
    completed_at TEXT,
    failure_reason TEXT,
    FOREIGN KEY (from_account_id) REFERENCES accounts(id),
    FOREIGN KEY (to_account_id) REFERENCES accounts(id)
  );

  CREATE TABLE IF NOT EXISTS interest_records (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    balance_snapshot INTEGER NOT NULL,
    daily_interest INTEGER NOT NULL,
    date TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id),
    UNIQUE(account_id, date)
  );

  CREATE TABLE IF NOT EXISTS interest_settlements (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    total_interest INTEGER NOT NULL,
    settlement_date TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id),
    UNIQUE(account_id, year, month)
  );

  CREATE TABLE IF NOT EXISTS system_config (
    id TEXT PRIMARY KEY DEFAULT 'config',
    daily_interest_rate REAL NOT NULL DEFAULT 0.0001,
    updated_at TEXT NOT NULL,
    CHECK (id = 'config')
  );

  CREATE TABLE IF NOT EXISTS locks (
    id TEXT PRIMARY KEY,
    locked INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL
  );

  INSERT OR IGNORE INTO system_config (id, daily_interest_rate, updated_at)
  VALUES ('config', 0.0001, datetime('now'));

  INSERT OR IGNORE INTO locks (id, locked, updated_at)
  VALUES ('interest_calc', 0, datetime('now'));
`);

export default db;
