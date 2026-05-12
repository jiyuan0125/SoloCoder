import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '..', 'disaster.db');
const db = new Database(dbPath);

db.pragma('journal_mode = WAL');

db.exec(`
  CREATE TABLE IF NOT EXISTS disaster_reports (
    id TEXT PRIMARY KEY,
    disaster_type TEXT NOT NULL,
    occurrence_time TEXT NOT NULL,
    location TEXT NOT NULL,
    affected_population TEXT DEFAULT '待核实',
    evacuated_population TEXT DEFAULT '待核实',
    death_missing_count TEXT DEFAULT '待核实',
    crop_area_affected TEXT DEFAULT '待核实',
    houses_damaged TEXT DEFAULT '待核实',
    direct_economic_loss TEXT DEFAULT '待核实',
    status TEXT DEFAULT '待核查',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    merged_from TEXT,
    report_count INTEGER DEFAULT 1,
    is_locked INTEGER DEFAULT 0
  );

  CREATE TABLE IF NOT EXISTS verification_records (
    id TEXT PRIMARY KEY,
    report_id TEXT NOT NULL,
    verifier_id TEXT NOT NULL,
    verifier_name TEXT NOT NULL,
    result TEXT NOT NULL,
    comments TEXT,
    verified_at TEXT NOT NULL,
    deadline TEXT NOT NULL,
    FOREIGN KEY (report_id) REFERENCES disaster_reports(id)
  );

  CREATE TABLE IF NOT EXISTS published_records (
    id TEXT PRIMARY KEY,
    report_id TEXT NOT NULL,
    publisher_id TEXT NOT NULL,
    publisher_name TEXT NOT NULL,
    content TEXT NOT NULL,
    version INTEGER DEFAULT 1,
    previous_content TEXT,
    published_at TEXT NOT NULL,
    FOREIGN KEY (report_id) REFERENCES disaster_reports(id)
  );

  CREATE TABLE IF NOT EXISTS modification_requests (
    id TEXT PRIMARY KEY,
    report_id TEXT NOT NULL,
    requester_id TEXT NOT NULL,
    requester_name TEXT NOT NULL,
    changes TEXT NOT NULL,
    reason TEXT,
    status TEXT DEFAULT '待审批',
    approved_by TEXT,
    approved_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (report_id) REFERENCES disaster_reports(id)
  );

  CREATE INDEX IF NOT EXISTS idx_reports_location_time ON disaster_reports(location, occurrence_time);
  CREATE INDEX IF NOT EXISTS idx_reports_status ON disaster_reports(status);
  CREATE INDEX IF NOT EXISTS idx_verification_report ON verification_records(report_id);
  CREATE INDEX IF NOT EXISTS idx_published_report ON published_records(report_id);
`);

export default db;
