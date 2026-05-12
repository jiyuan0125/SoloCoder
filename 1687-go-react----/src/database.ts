import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '..', 'counseling.db');
const db = new Database(dbPath);

db.exec(`
  CREATE TABLE IF NOT EXISTS counselors (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    certificate TEXT NOT NULL,
    specialty TEXT NOT NULL,
    experience INTEGER NOT NULL,
    fee REAL NOT NULL,
    status TEXT DEFAULT 'pending_review',
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS time_slots (
    id TEXT PRIMARY KEY,
    counselor_id TEXT NOT NULL,
    is_recurring INTEGER NOT NULL DEFAULT 0,
    day_of_week INTEGER,
    date TEXT,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    is_booked INTEGER NOT NULL DEFAULT 0,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (counselor_id) REFERENCES counselors(id)
  );

  CREATE TABLE IF NOT EXISTS clients (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL UNIQUE,
    no_show_count INTEGER DEFAULT 0,
    banned_until TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS appointments (
    id TEXT PRIMARY KEY,
    counselor_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_name TEXT NOT NULL,
    client_phone TEXT NOT NULL,
    problem_description TEXT NOT NULL,
    time_slot_id TEXT NOT NULL,
    appointment_date TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    status TEXT DEFAULT 'pending_confirmation',
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (counselor_id) REFERENCES counselors(id),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (time_slot_id) REFERENCES time_slots(id)
  );

  CREATE TABLE IF NOT EXISTS consultation_records (
    id TEXT PRIMARY KEY,
    appointment_id TEXT NOT NULL UNIQUE,
    counselor_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    consultation_date TEXT NOT NULL,
    duration INTEGER NOT NULL,
    summary TEXT NOT NULL,
    follow_up_suggestions TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (appointment_id) REFERENCES appointments(id),
    FOREIGN KEY (counselor_id) REFERENCES counselors(id),
    FOREIGN KEY (client_id) REFERENCES clients(id)
  );

  CREATE INDEX IF NOT EXISTS idx_time_slots_counselor ON time_slots(counselor_id);
  CREATE INDEX IF NOT EXISTS idx_appointments_counselor ON appointments(counselor_id);
  CREATE INDEX IF NOT EXISTS idx_appointments_client ON appointments(client_id);
  CREATE INDEX IF NOT EXISTS idx_appointments_date ON appointments(appointment_date);
`);

export default db;
