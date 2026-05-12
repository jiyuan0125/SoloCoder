import Database from 'better-sqlite3';
import fs from 'fs';
import path from 'path';

const dbDir = path.join(__dirname, '..', 'data');
if (!fs.existsSync(dbDir)) {
  fs.mkdirSync(dbDir, { recursive: true });
}

const dbPath = path.join(dbDir, 'organizations.db');
const db = new Database(dbPath);

db.pragma('journal_mode = WAL');

function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS organizations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      normalized_name TEXT NOT NULL,
      credit_code TEXT NOT NULL UNIQUE,
      type TEXT NOT NULL CHECK(type IN ('社会团体', '民办非企业', '基金会')),
      business_supervisor TEXT NOT NULL,
      legal_representative TEXT NOT NULL,
      address TEXT NOT NULL,
      business_scope TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT '注册中' CHECK(status IN ('注册中', '已注册', '已注销')),
      re_submit_count INTEGER DEFAULT 0,
      last_reject_time TEXT,
      certificate_no TEXT,
      certificate_issue_date TEXT,
      certificate_expiry_date TEXT,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS registration_applications (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      organization_id INTEGER NOT NULL,
      status TEXT NOT NULL DEFAULT '待审核' CHECK(status IN ('待审核', '已通过', '已驳回')),
      reject_reason TEXT,
      submit_time TEXT NOT NULL,
      FOREIGN KEY (organization_id) REFERENCES organizations(id)
    );

    CREATE TABLE IF NOT EXISTS annual_inspections (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      organization_id INTEGER NOT NULL,
      year INTEGER NOT NULL,
      annual_report TEXT NOT NULL,
      financial_audit TEXT NOT NULL,
      activity_list TEXT NOT NULL,
      result TEXT CHECK(result IN ('合格', '基本合格', '不合格')),
      submit_time TEXT NOT NULL,
      review_time TEXT,
      FOREIGN KEY (organization_id) REFERENCES organizations(id)
    );

    CREATE TABLE IF NOT EXISTS evaluations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      organization_id INTEGER NOT NULL,
      evaluation_year INTEGER NOT NULL,
      score1 INTEGER NOT NULL CHECK(score1 BETWEEN 0 AND 25),
      score2 INTEGER NOT NULL CHECK(score2 BETWEEN 0 AND 25),
      score3 INTEGER NOT NULL CHECK(score3 BETWEEN 0 AND 25),
      score4 INTEGER NOT NULL CHECK(score4 BETWEEN 0 AND 25),
      total_score INTEGER NOT NULL,
      grade TEXT NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (organization_id) REFERENCES organizations(id)
    );

    CREATE TABLE IF NOT EXISTS reminders (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      organization_id INTEGER NOT NULL,
      type TEXT NOT NULL,
      content TEXT NOT NULL,
      created_at TEXT NOT NULL,
      is_read INTEGER DEFAULT 0,
      FOREIGN KEY (organization_id) REFERENCES organizations(id)
    );
  `);
}

initDatabase();

export default db;
