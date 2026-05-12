import Database from 'better-sqlite3';

const db = new Database('./data.db');

db.exec(`
  CREATE TABLE IF NOT EXISTS accounts_receivable (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    supplier_id TEXT NOT NULL,
    buyer_name TEXT NOT NULL,
    amount INTEGER NOT NULL,
    due_date TEXT NOT NULL,
    contract_no TEXT NOT NULL,
    created_at TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS factoring_finance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    receivable_id INTEGER NOT NULL,
    financing_rate REAL NOT NULL,
    loan_amount INTEGER NOT NULL,
    annual_interest_rate REAL NOT NULL DEFAULT 0.08,
    penalty_rate REAL NOT NULL DEFAULT 0.15,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TEXT NOT NULL,
    funded_at TEXT,
    settled_at TEXT,
    buyer_paid_at TEXT,
    interest_amount INTEGER,
    settlement_amount INTEGER,
    FOREIGN KEY (receivable_id) REFERENCES accounts_receivable(id)
  );

  CREATE TABLE IF NOT EXISTS warehouse_receipts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    supplier_id TEXT NOT NULL,
    product_name TEXT NOT NULL,
    quantity REAL NOT NULL,
    unit_price INTEGER NOT NULL,
    warehouse_address TEXT NOT NULL,
    market_value INTEGER NOT NULL,
    created_at TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS pledge_finance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    receipt_id INTEGER NOT NULL,
    pledge_rate REAL NOT NULL,
    loan_amount INTEGER NOT NULL,
    margin_threshold REAL NOT NULL DEFAULT 0.6,
    status TEXT NOT NULL DEFAULT 'pending',
    margin_call_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    FOREIGN KEY (receipt_id) REFERENCES warehouse_receipts(id)
  );
`);

export default db;
