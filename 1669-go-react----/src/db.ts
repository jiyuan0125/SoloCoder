import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '..', 'reconciliation.db');
const db = new Database(dbPath);
export { db };

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

db.exec(`
  CREATE TABLE IF NOT EXISTS channel_transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    transaction_id TEXT UNIQUE NOT NULL,
    transaction_time DATETIME NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    counterparty TEXT NOT NULL,
    channel_name TEXT NOT NULL,
    status TEXT NOT NULL,
    matched_order_id TEXT,
    matched_report_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS internal_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id TEXT UNIQUE NOT NULL,
    order_time DATETIME NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    payment_method TEXT NOT NULL,
    status TEXT NOT NULL,
    matched_transaction_id TEXT,
    matched_report_id INTEGER,
    is_supplemental INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS reconciliation_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_date DATE UNIQUE NOT NULL,
    matched_count INTEGER NOT NULL DEFAULT 0,
    channel_extra_count INTEGER NOT NULL DEFAULT 0,
    order_extra_count INTEGER NOT NULL DEFAULT 0,
    reconciliation_time DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS adjustment_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    original_amount DECIMAL(12, 2),
    adjusted_amount DECIMAL(12, 2),
    difference DECIMAL(12, 2),
    transaction_id TEXT,
    order_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (report_id) REFERENCES reconciliation_reports(id)
  );

  CREATE TABLE IF NOT EXISTS processing_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
`);

console.log('Database initialized at:', dbPath);
