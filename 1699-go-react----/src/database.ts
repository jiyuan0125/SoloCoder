import { open, Database } from 'sqlite';
import sqlite3 from 'sqlite3';

let db: Database | null = null;

export async function getDb(): Promise<Database> {
  if (!db) {
    db = await open({
      filename: './vaccine.db',
      driver: sqlite3.Database
    });
    await initDb(db);
  }
  return db;
}

async function initDb(db: Database): Promise<void> {
  await db.exec(`
    CREATE TABLE IF NOT EXISTS vaccine_batches (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      manufacturer TEXT NOT NULL,
      batch_number TEXT NOT NULL UNIQUE,
      specification TEXT NOT NULL,
      expiry_date TEXT NOT NULL,
      stock INTEGER NOT NULL DEFAULT 0,
      storage_temperature TEXT NOT NULL,
      safe_stock INTEGER NOT NULL DEFAULT 10,
      status TEXT NOT NULL DEFAULT 'normal'
    );

    CREATE TABLE IF NOT EXISTS reservations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      resident_id TEXT NOT NULL,
      resident_name TEXT NOT NULL,
      vaccine_name TEXT NOT NULL,
      date TEXT NOT NULL,
      time_slot TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT 'active',
      is_booster INTEGER NOT NULL DEFAULT 0,
      merged_from TEXT
    );

    CREATE TABLE IF NOT EXISTS vaccination_records (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      certificate_number TEXT NOT NULL UNIQUE,
      reservation_id INTEGER,
      resident_id TEXT NOT NULL,
      resident_name TEXT NOT NULL,
      vaccine_name TEXT NOT NULL,
      batch_number TEXT NOT NULL,
      injection_site TEXT NOT NULL,
      dose INTEGER NOT NULL,
      vaccination_date TEXT NOT NULL,
      next_vaccination_date TEXT,
      is_overdue INTEGER NOT NULL DEFAULT 0
    );

    CREATE TABLE IF NOT EXISTS adverse_reactions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      record_id INTEGER NOT NULL,
      reaction_id TEXT NOT NULL UNIQUE,
      symptoms TEXT NOT NULL,
      severity TEXT NOT NULL,
      report_date TEXT NOT NULL,
      is_handled INTEGER NOT NULL DEFAULT 0
    );

    CREATE TABLE IF NOT EXISTS notifications (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      type TEXT NOT NULL,
      message TEXT NOT NULL,
      target_id TEXT NOT NULL,
      created_at TEXT NOT NULL,
      is_read INTEGER NOT NULL DEFAULT 0
    );

    CREATE TABLE IF NOT EXISTS certificate_sequence (
      date_key TEXT PRIMARY KEY,
      current_sequence INTEGER NOT NULL DEFAULT 0
    );
  `);
}
