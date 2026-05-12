import sqlite3 from 'sqlite3';
import { promisify } from 'util';

let db: sqlite3.Database;

export interface DatabaseWrapper {
  run: (sql: string, params?: any[]) => Promise<{ lastID: number; changes: number }>;
  get: <T = any>(sql: string, params?: any[]) => Promise<T | undefined>;
  all: <T = any>(sql: string, params?: any[]) => Promise<T[]>;
  exec: (sql: string) => Promise<void>;
  beginTransaction: () => Promise<void>;
  commit: () => Promise<void>;
  rollback: () => Promise<void>;
}

let dbWrapper: DatabaseWrapper;

export async function initDatabase(): Promise<DatabaseWrapper> {
  if (dbWrapper) return dbWrapper;

  return new Promise((resolve, reject) => {
    db = new sqlite3.Database('./developer-portal.db', (err) => {
      if (err) {
        reject(err);
        return;
      }

      dbWrapper = {
        run: promisify(db.run.bind(db)) as (sql: string, params?: any[]) => Promise<{ lastID: number; changes: number }>,
        get: promisify(db.get.bind(db)) as <T = any>(sql: string, params?: any[]) => Promise<T | undefined>,
        all: promisify(db.all.bind(db)) as <T = any>(sql: string, params?: any[]) => Promise<T[]>,
        exec: promisify(db.exec.bind(db)) as (sql: string) => Promise<void>,
        beginTransaction: () => new Promise<void>((res, rej) => db.run('BEGIN TRANSACTION', (e) => e ? rej(e) : res())),
        commit: () => new Promise<void>((res, rej) => db.run('COMMIT', (e) => e ? rej(e) : res())),
        rollback: () => new Promise<void>((res, rej) => db.run('ROLLBACK', (e) => e ? rej(e) : res()))
      };

      const tables = `
        CREATE TABLE IF NOT EXISTS developers (
          id TEXT PRIMARY KEY,
          name TEXT NOT NULL,
          email TEXT NOT NULL UNIQUE,
          company TEXT,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS apps (
          id TEXT PRIMARY KEY,
          developer_id TEXT NOT NULL,
          name TEXT NOT NULL,
          description TEXT,
          callback_url TEXT NOT NULL,
          status TEXT DEFAULT 'active',
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          FOREIGN KEY (developer_id) REFERENCES developers(id)
        );

        CREATE TABLE IF NOT EXISTS api_keys (
          id TEXT PRIMARY KEY,
          app_id TEXT NOT NULL,
          key_hash TEXT NOT NULL,
          status TEXT DEFAULT 'active',
          version INTEGER DEFAULT 1,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          transition_until DATETIME,
          FOREIGN KEY (app_id) REFERENCES apps(id)
        );

        CREATE TABLE IF NOT EXISTS key_resets (
          id TEXT PRIMARY KEY,
          app_id TEXT NOT NULL,
          reset_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          FOREIGN KEY (app_id) REFERENCES apps(id)
        );

        CREATE TABLE IF NOT EXISTS audit_logs (
          id TEXT PRIMARY KEY,
          developer_id TEXT,
          app_id TEXT,
          action TEXT NOT NULL,
          details TEXT,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS api_stats (
          id TEXT PRIMARY KEY,
          app_id TEXT NOT NULL,
          hour_start DATETIME NOT NULL,
          total_calls INTEGER DEFAULT 0,
          success_calls INTEGER DEFAULT 0,
          total_response_time INTEGER DEFAULT 0,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          UNIQUE(app_id, hour_start)
        );

        CREATE TABLE IF NOT EXISTS notifications (
          id TEXT PRIMARY KEY,
          app_id TEXT NOT NULL,
          type TEXT NOT NULL,
          message TEXT NOT NULL,
          sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          FOREIGN KEY (app_id) REFERENCES apps(id)
        );
      `;

      db.exec(tables, (err) => {
        if (err) reject(err);
        else resolve(dbWrapper);
      });
    });
  });
}

export function getDatabase(): DatabaseWrapper {
  return dbWrapper;
}
