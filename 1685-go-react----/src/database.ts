import Database from 'better-sqlite3';
import type { Database as DatabaseType } from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'elderly_care.db');
const db: DatabaseType = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS elders (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      id_card TEXT NOT NULL UNIQUE,
      phone TEXT NOT NULL,
      address TEXT NOT NULL,
      emergency_contact TEXT NOT NULL,
      emergency_contact_phone TEXT NOT NULL,
      service_level TEXT NOT NULL,
      transition_until TEXT,
      previous_level TEXT,
      visit_plan_needs_manual_adjustment INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS assessments (
      id TEXT PRIMARY KEY,
      elder_id TEXT NOT NULL,
      self_care_score INTEGER NOT NULL,
      cognitive_score INTEGER NOT NULL,
      total_score INTEGER NOT NULL,
      service_level TEXT NOT NULL,
      assessment_date TEXT NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (elder_id) REFERENCES elders(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS call_records (
      id TEXT PRIMARY KEY,
      elder_id TEXT NOT NULL,
      type TEXT NOT NULL,
      status TEXT NOT NULL,
      called_at TEXT NOT NULL,
      responded_at TEXT,
      response_time INTEGER,
      staff_id TEXT,
      created_at TEXT NOT NULL,
      FOREIGN KEY (elder_id) REFERENCES elders(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS visit_plans (
      id TEXT PRIMARY KEY,
      elder_id TEXT NOT NULL,
      scheduled_date TEXT NOT NULL,
      status TEXT NOT NULL,
      completed_at TEXT,
      notes TEXT,
      created_at TEXT NOT NULL,
      FOREIGN KEY (elder_id) REFERENCES elders(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS health_records (
      id TEXT PRIMARY KEY,
      elder_id TEXT NOT NULL,
      systolic INTEGER,
      diastolic INTEGER,
      blood_sugar REAL,
      record_date TEXT NOT NULL,
      is_abnormal INTEGER NOT NULL DEFAULT 0,
      medical_advice TEXT,
      created_at TEXT NOT NULL,
      FOREIGN KEY (elder_id) REFERENCES elders(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS notifications (
      id TEXT PRIMARY KEY,
      call_record_id TEXT NOT NULL,
      recipient_type TEXT NOT NULL,
      notified_at TEXT NOT NULL,
      message TEXT NOT NULL,
      FOREIGN KEY (call_record_id) REFERENCES call_records(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_assessments_elder_id ON assessments(elder_id);
    CREATE INDEX IF NOT EXISTS idx_call_records_elder_id ON call_records(elder_id);
    CREATE INDEX IF NOT EXISTS idx_call_records_status ON call_records(status);
    CREATE INDEX IF NOT EXISTS idx_visit_plans_elder_id ON visit_plans(elder_id);
    CREATE INDEX IF NOT EXISTS idx_visit_plans_scheduled_date ON visit_plans(scheduled_date);
    CREATE INDEX IF NOT EXISTS idx_health_records_elder_id ON health_records(elder_id);
    CREATE INDEX IF NOT EXISTS idx_health_records_record_date ON health_records(record_date);
  `);
}

export { db };
