import sqlite3 from 'sqlite3';
import { open, Database } from 'sqlite';

let db: Database | null = null;

export async function initDb(): Promise<Database> {
  if (db) return db;

  db = await open({
    filename: ':memory:',
    driver: sqlite3.Database
  });

  await db.exec(`
    CREATE TABLE services (
      name TEXT PRIMARY KEY,
      created_at INTEGER NOT NULL
    );

    CREATE TABLE metrics (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      service_name TEXT NOT NULL,
      metric_name TEXT NOT NULL,
      value REAL NOT NULL,
      timestamp INTEGER NOT NULL,
      UNIQUE(service_name, metric_name, timestamp),
      FOREIGN KEY (service_name) REFERENCES services(name)
    );

    CREATE TABLE alert_rules (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      metric_name TEXT NOT NULL,
      operator TEXT NOT NULL CHECK (operator IN ('gt', 'lt', 'eq')),
      threshold REAL NOT NULL,
      duration INTEGER NOT NULL,
      level TEXT NOT NULL CHECK (level IN ('warning', 'critical', 'emergency'))
    );

    CREATE TABLE alerts (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      rule_id INTEGER NOT NULL,
      service_name TEXT NOT NULL,
      current_value REAL NOT NULL,
      level TEXT NOT NULL,
      status TEXT NOT NULL CHECK (status IN ('pending', 'processing', 'recovered')),
      timestamp INTEGER NOT NULL,
      last_value REAL NOT NULL,
      last_timestamp INTEGER NOT NULL,
      FOREIGN KEY (rule_id) REFERENCES alert_rules(id),
      UNIQUE(rule_id, service_name, status)
    );

    CREATE INDEX idx_metrics_service ON metrics(service_name);
    CREATE INDEX idx_metrics_timestamp ON metrics(timestamp);
    CREATE INDEX idx_metrics_service_metric ON metrics(service_name, metric_name);
  `);

  return db;
}

export function getDb(): Database {
  if (!db) {
    throw new Error('Database not initialized');
  }
  return db;
}
