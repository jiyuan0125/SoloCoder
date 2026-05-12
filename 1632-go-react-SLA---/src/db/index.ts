import Database from 'better-sqlite3';
import type { Database as DatabaseType } from 'better-sqlite3';
import path from 'path';

const DB_PATH = path.resolve(process.env.DB_PATH || './sla.db');

const db: DatabaseType = new Database(DB_PATH);

db.pragma('journal_mode = WAL');

const initDb = () => {
  db.exec(`
    CREATE TABLE IF NOT EXISTS services (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      description TEXT,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS sla_definitions (
      id TEXT PRIMARY KEY,
      service_id TEXT NOT NULL,
      availability_target REAL NOT NULL,
      response_time_target INTEGER NOT NULL,
      error_rate_target REAL NOT NULL,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL,
      FOREIGN KEY (service_id) REFERENCES services(id)
    );

    CREATE TABLE IF NOT EXISTS incidents (
      id TEXT PRIMARY KEY,
      service_id TEXT NOT NULL,
      start_time TEXT NOT NULL,
      end_time TEXT,
      duration_minutes INTEGER NOT NULL,
      description TEXT,
      is_automatic INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL,
      FOREIGN KEY (service_id) REFERENCES services(id)
    );

    CREATE TABLE IF NOT EXISTS metrics (
      id TEXT PRIMARY KEY,
      service_id TEXT NOT NULL,
      timestamp TEXT NOT NULL,
      response_time INTEGER NOT NULL,
      is_error INTEGER NOT NULL DEFAULT 0,
      FOREIGN KEY (service_id) REFERENCES services(id)
    );

    CREATE TABLE IF NOT EXISTS reports (
      id TEXT PRIMARY KEY,
      year INTEGER NOT NULL,
      month INTEGER NOT NULL,
      service_id TEXT NOT NULL,
      sla_definition_id TEXT NOT NULL,
      availability_achieved REAL NOT NULL,
      response_time_p99 INTEGER NOT NULL,
      error_rate REAL NOT NULL,
      is_confirmed INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL,
      FOREIGN KEY (service_id) REFERENCES services(id),
      FOREIGN KEY (sla_definition_id) REFERENCES sla_definitions(id)
    );

    CREATE TABLE IF NOT EXISTS report_details (
      id TEXT PRIMARY KEY,
      report_id TEXT NOT NULL,
      metric_type TEXT NOT NULL,
      target REAL NOT NULL,
      actual REAL NOT NULL,
      is_achieved INTEGER NOT NULL,
      is_confirmed INTEGER NOT NULL DEFAULT 0,
      FOREIGN KEY (report_id) REFERENCES reports(id)
    );

    CREATE TABLE IF NOT EXISTS breaches (
      id TEXT PRIMARY KEY,
      report_id TEXT NOT NULL,
      incident_id TEXT NOT NULL,
      metric_type TEXT NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (report_id) REFERENCES reports(id),
      FOREIGN KEY (incident_id) REFERENCES incidents(id),
      UNIQUE(report_id, incident_id, metric_type)
    );
  `);
};

initDb();

export default db;
