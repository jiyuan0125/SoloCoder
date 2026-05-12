import Database from 'better-sqlite3';

const db = new Database('settlement.db');

export function initDB() {
  db.exec(`
    PRAGMA foreign_keys = ON;

    CREATE TABLE IF NOT EXISTS sellers (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      level TEXT NOT NULL,
      created_at INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS transactions (
      id TEXT PRIMARY KEY,
      seller_id TEXT NOT NULL,
      type TEXT NOT NULL,
      amount INTEGER NOT NULL,
      cycle TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      FOREIGN KEY (seller_id) REFERENCES sellers(id)
    );

    CREATE TABLE IF NOT EXISTS settlements (
      id TEXT PRIMARY KEY,
      seller_id TEXT NOT NULL,
      cycle TEXT NOT NULL,
      total_receivable INTEGER NOT NULL,
      total_refund INTEGER NOT NULL,
      net_amount INTEGER NOT NULL,
      fee INTEGER NOT NULL,
      payable_amount INTEGER NOT NULL,
      status TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      confirmed_at INTEGER,
      paid_at INTEGER,
      FOREIGN KEY (seller_id) REFERENCES sellers(id),
      UNIQUE(seller_id, cycle)
    );

    CREATE TABLE IF NOT EXISTS fee_records (
      id TEXT PRIMARY KEY,
      settlement_id TEXT NOT NULL,
      seller_id TEXT NOT NULL,
      cycle TEXT NOT NULL,
      amount INTEGER NOT NULL,
      created_at INTEGER NOT NULL,
      FOREIGN KEY (settlement_id) REFERENCES settlements(id)
    );
  `);
}

export default db;
