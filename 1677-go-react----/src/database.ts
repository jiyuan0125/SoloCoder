import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'charity.db');
export const db = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

db.exec(`
CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  description TEXT NOT NULL,
  target_amount INTEGER NOT NULL,
  start_date TEXT NOT NULL,
  end_date TEXT NOT NULL,
  has_tax_deductible_qualification INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS donations (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  donor_name TEXT NOT NULL,
  id_card_last4 TEXT NOT NULL,
  amount INTEGER NOT NULL,
  certificate_number TEXT,
  donated_at TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS donation_records (
  id TEXT PRIMARY KEY,
  donation_id TEXT NOT NULL,
  donor_name TEXT NOT NULL,
  id_card_last4 TEXT NOT NULL,
  project_id TEXT NOT NULL,
  amount INTEGER NOT NULL,
  year INTEGER NOT NULL,
  is_tax_deductible INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  FOREIGN KEY (donation_id) REFERENCES donations(id),
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS certificates (
  id TEXT PRIMARY KEY,
  certificate_number TEXT UNIQUE NOT NULL,
  donation_id TEXT NOT NULL,
  donor_name TEXT NOT NULL,
  id_card_last4 TEXT NOT NULL,
  project_id TEXT NOT NULL,
  amount INTEGER NOT NULL,
  issued_at TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (donation_id) REFERENCES donations(id),
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS tax_deduction_vouchers (
  id TEXT PRIMARY KEY,
  voucher_number TEXT UNIQUE NOT NULL,
  donation_id TEXT NOT NULL,
  donor_name TEXT NOT NULL,
  id_card_last4 TEXT NOT NULL,
  project_id TEXT NOT NULL,
  deductible_amount INTEGER NOT NULL,
  tax_year INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  voided_at TEXT,
  issued_at TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (donation_id) REFERENCES donations(id),
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS donor_annual_summary (
  id TEXT PRIMARY KEY,
  donor_name TEXT NOT NULL,
  id_card_last4 TEXT NOT NULL,
  year INTEGER NOT NULL,
  total_donation_amount INTEGER NOT NULL DEFAULT 0,
  total_tax_deductible_amount INTEGER NOT NULL DEFAULT 0,
  tax_income_base INTEGER NOT NULL DEFAULT 0,
  tax_limit INTEGER NOT NULL DEFAULT 0,
  updated_at TEXT NOT NULL,
  UNIQUE(donor_name, id_card_last4, year)
);

CREATE TABLE IF NOT EXISTS certificate_sequences (
  date TEXT PRIMARY KEY,
  sequence INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS vouchers_sequences (
  date TEXT PRIMARY KEY,
  sequence INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS project_warnings (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  warning_type TEXT NOT NULL,
  message TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE INDEX IF NOT EXISTS idx_donations_project_id ON donations(project_id);
CREATE INDEX IF NOT EXISTS idx_donation_records_donor ON donation_records(donor_name, id_card_last4, year);
CREATE INDEX IF NOT EXISTS idx_donation_records_donation_id ON donation_records(donation_id);
CREATE INDEX IF NOT EXISTS idx_certificates_donation_id ON certificates(donation_id);
CREATE INDEX IF NOT EXISTS idx_vouchers_donation_id ON tax_deduction_vouchers(donation_id);
CREATE INDEX IF NOT EXISTS idx_warnings_project_id ON project_warnings(project_id);
`);
