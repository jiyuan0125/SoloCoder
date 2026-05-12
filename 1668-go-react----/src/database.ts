import Database from 'better-sqlite3';
import path from 'path';

const DB_PATH = path.join(__dirname, '..', 'invoice.db');

export const db = new Database(DB_PATH);
db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS contracts (
      id TEXT PRIMARY KEY,
      code TEXT NOT NULL UNIQUE,
      name TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT 'pending',
      created_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS invoices (
      id TEXT PRIMARY KEY,
      type TEXT NOT NULL,
      code TEXT NOT NULL,
      number TEXT NOT NULL,
      issue_date TEXT NOT NULL,
      buyer_name TEXT NOT NULL,
      buyer_tax_id TEXT NOT NULL,
      seller_name TEXT NOT NULL,
      seller_tax_id TEXT NOT NULL,
      amount INTEGER NOT NULL,
      tax_rate REAL NOT NULL,
      tax_amount INTEGER NOT NULL,
      total_amount INTEGER NOT NULL,
      status TEXT NOT NULL DEFAULT 'draft',
      contract_id TEXT,
      original_invoice_id TEXT,
      is_red_invoice INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL,
      FOREIGN KEY (contract_id) REFERENCES contracts(id),
      FOREIGN KEY (original_invoice_id) REFERENCES invoices(id),
      UNIQUE(code, number)
    );

    CREATE INDEX IF NOT EXISTS idx_invoices_code_number ON invoices(code, number);
    CREATE INDEX IF NOT EXISTS idx_invoices_original_id ON invoices(original_invoice_id);
    CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
  `);
}
