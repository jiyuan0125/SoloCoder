import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '..', 'volunteers.db');
const db = new Database(dbPath) as any;

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

db.exec(`
  CREATE TABLE IF NOT EXISTS activities (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL,
    date_time TEXT NOT NULL,
    location TEXT NOT NULL,
    max_volunteers INTEGER NOT NULL,
    registration_deadline TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS volunteers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL UNIQUE,
    id_card TEXT NOT NULL,
    age INTEGER NOT NULL,
    guardian_name TEXT,
    guardian_phone TEXT,
    emergency_contact TEXT NOT NULL,
    entry_training_completed INTEGER DEFAULT 0,
    entry_training_date TEXT,
    last_annual_training_date TEXT,
    created_at TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS registrations (
    id TEXT PRIMARY KEY,
    activity_id TEXT NOT NULL,
    volunteer_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    is_special_position INTEGER DEFAULT 0,
    hours INTEGER DEFAULT 0,
    registered_at TEXT NOT NULL,
    reviewed_at TEXT,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (volunteer_id) REFERENCES volunteers(id) ON DELETE CASCADE,
    UNIQUE(activity_id, volunteer_id)
  );

  CREATE TABLE IF NOT EXISTS training_records (
    id TEXT PRIMARY KEY,
    volunteer_id TEXT NOT NULL,
    training_type TEXT NOT NULL,
    knowledge_score INTEGER NOT NULL,
    safety_score INTEGER NOT NULL,
    passed INTEGER NOT NULL,
    completed_at TEXT NOT NULL,
    FOREIGN KEY (volunteer_id) REFERENCES volunteers(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS points (
    id TEXT PRIMARY KEY,
    volunteer_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    description TEXT NOT NULL,
    source TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (volunteer_id) REFERENCES volunteers(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS reminders (
    id TEXT PRIMARY KEY,
    volunteer_id TEXT NOT NULL,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    due_date TEXT NOT NULL,
    is_read INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    FOREIGN KEY (volunteer_id) REFERENCES volunteers(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    volunteer_id TEXT NOT NULL,
    activity_id TEXT,
    content TEXT NOT NULL,
    sent_at TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    FOREIGN KEY (volunteer_id) REFERENCES volunteers(id) ON DELETE CASCADE,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE SET NULL
  );

  CREATE INDEX IF NOT EXISTS idx_registrations_activity ON registrations(activity_id);
  CREATE INDEX IF NOT EXISTS idx_registrations_volunteer ON registrations(volunteer_id);
  CREATE INDEX IF NOT EXISTS idx_points_volunteer ON points(volunteer_id);
  CREATE INDEX IF NOT EXISTS idx_reminders_volunteer ON reminders(volunteer_id);
`);

export default db;
