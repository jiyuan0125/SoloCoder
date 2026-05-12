import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'maternity-center.db');
export const db = new Database(dbPath);

db.pragma('journal_mode = WAL');

db.exec(`
  CREATE TABLE IF NOT EXISTS mothers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    expected_due_date TEXT NOT NULL,
    room_type TEXT NOT NULL,
    emergency_contact_name TEXT NOT NULL,
    emergency_contact_phone TEXT NOT NULL,
    check_in_date TEXT,
    check_out_date TEXT,
    booking_date TEXT NOT NULL,
    status TEXT NOT NULL,
    total_deposit REAL DEFAULT 0
  );

  CREATE TABLE IF NOT EXISTS babies (
    id TEXT PRIMARY KEY,
    mother_id TEXT NOT NULL,
    name TEXT NOT NULL,
    birth_weight REAL NOT NULL,
    birth_length REAL NOT NULL,
    feeding_type TEXT NOT NULL,
    birth_date TEXT NOT NULL,
    FOREIGN KEY (mother_id) REFERENCES mothers(id)
  );

  CREATE TABLE IF NOT EXISTS baby_logs (
    id TEXT PRIMARY KEY,
    baby_id TEXT NOT NULL,
    log_date TEXT NOT NULL,
    feeding_time TEXT,
    feeding_amount REAL,
    diaper_change BOOLEAN DEFAULT 0,
    temperature REAL,
    sleep_duration REAL,
    jaundice_index REAL,
    current_weight REAL,
    anomaly_notes TEXT,
    FOREIGN KEY (baby_id) REFERENCES babies(id)
  );

  CREATE TABLE IF NOT EXISTS staff (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS shifts (
    id TEXT PRIMARY KEY,
    shift_type TEXT NOT NULL,
    shift_date TEXT NOT NULL,
    UNIQUE(shift_type, shift_date)
  );

  CREATE TABLE IF NOT EXISTS shift_assignments (
    id TEXT PRIMARY KEY,
    shift_id TEXT NOT NULL,
    staff_id TEXT NOT NULL,
    FOREIGN KEY (shift_id) REFERENCES shifts(id),
    FOREIGN KEY (staff_id) REFERENCES staff(id),
    UNIQUE(shift_id, staff_id)
  );
`);

export const ROOM_TYPES = {
  STANDARD: 'standard',
  DELUXE: 'deluxe',
  SUITE: 'suite'
} as const;

export const ROOM_PRICES: Record<string, number> = {
  [ROOM_TYPES.STANDARD]: 680,
  [ROOM_TYPES.DELUXE]: 980,
  [ROOM_TYPES.SUITE]: 1680
};

export const SHIFT_TYPES = {
  MORNING: 'morning',
  AFTERNOON: 'afternoon',
  NIGHT: 'night'
} as const;

export const STAFF_ROLES = {
  MATERNAL_CARE: 'maternal_care',
  PEDIATRIC_NURSE: 'pediatric_nurse'
} as const;
