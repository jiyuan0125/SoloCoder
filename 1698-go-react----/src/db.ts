import sqlite3 from 'sqlite3';
import { open } from 'sqlite';
import path from 'path';

export async function openDb() {
  const dbPath = path.join(__dirname, '..', 'epidemic.db');
  return open({
    filename: dbPath,
    driver: sqlite3.Database
  });
}

export async function initDb() {
  const db = await openDb();
  
  await db.exec(`
    CREATE TABLE IF NOT EXISTS persons (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      id_card TEXT UNIQUE NOT NULL,
      phone TEXT,
      address TEXT,
      is_key_person INTEGER DEFAULT 0,
      created_at TEXT DEFAULT (datetime('now'))
    );

    CREATE TABLE IF NOT EXISTS investigations (
      id TEXT PRIMARY KEY,
      person_id TEXT NOT NULL,
      investigation_date TEXT NOT NULL,
      investigation_method TEXT,
      travel_history TEXT,
      health_status TEXT,
      result TEXT CHECK(result IN ('normal', 'need_isolation', 'need_hospital')) NOT NULL,
      created_at TEXT DEFAULT (datetime('now')),
      FOREIGN KEY (person_id) REFERENCES persons(id)
    );

    CREATE TABLE IF NOT EXISTS isolation_records (
      id TEXT PRIMARY KEY,
      person_id TEXT NOT NULL,
      isolation_type TEXT CHECK(isolation_type IN ('home', 'centralized')) NOT NULL,
      start_date TEXT NOT NULL,
      end_date TEXT NOT NULL,
      status TEXT CHECK(status IN ('active', 'pending_discharge', 'discharged', 'transferred_hospital')) NOT NULL DEFAULT 'active',
      created_at TEXT DEFAULT (datetime('now')),
      FOREIGN KEY (person_id) REFERENCES persons(id)
    );

    CREATE TABLE IF NOT EXISTS health_records (
      id TEXT PRIMARY KEY,
      person_id TEXT NOT NULL,
      isolation_id TEXT,
      record_date TEXT NOT NULL,
      temperature REAL NOT NULL,
      symptoms TEXT,
      is_normal INTEGER DEFAULT 1,
      created_at TEXT DEFAULT (datetime('now')),
      FOREIGN KEY (person_id) REFERENCES persons(id),
      FOREIGN KEY (isolation_id) REFERENCES isolation_records(id)
    );

    CREATE TABLE IF NOT EXISTS test_records (
      id TEXT PRIMARY KEY,
      person_id TEXT NOT NULL,
      isolation_id TEXT,
      test_day INTEGER NOT NULL,
      test_date TEXT NOT NULL,
      result TEXT CHECK(result IN ('positive', 'negative', 'pending')) NOT NULL DEFAULT 'pending',
      created_at TEXT DEFAULT (datetime('now')),
      FOREIGN KEY (person_id) REFERENCES persons(id),
      FOREIGN KEY (isolation_id) REFERENCES isolation_records(id)
    );

    CREATE TABLE IF NOT EXISTS discharge_notifications (
      id TEXT PRIMARY KEY,
      isolation_id TEXT NOT NULL UNIQUE,
      person_id TEXT NOT NULL,
      notification_date TEXT NOT NULL,
      health_status TEXT NOT NULL,
      doctor_confirmed INTEGER DEFAULT 0,
      doctor_name TEXT,
      confirmed_at TEXT,
      created_at TEXT DEFAULT (datetime('now')),
      FOREIGN KEY (isolation_id) REFERENCES isolation_records(id),
      FOREIGN KEY (person_id) REFERENCES persons(id)
    );
  `);

  return db;
}
