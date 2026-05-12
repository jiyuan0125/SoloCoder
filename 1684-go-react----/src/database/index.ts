import Database from 'better-sqlite3';
import { SCHEMA_STATEMENTS, INITIAL_SETTINGS } from './schema';
import path from 'path';

const DB_PATH = process.env.DB_PATH || path.join(process.cwd(), 'nursing-home.db');

let dbInstance: Database.Database | null = null;

export function getDb(): Database.Database {
  if (!dbInstance) {
    dbInstance = new Database(DB_PATH);
    dbInstance.pragma('journal_mode = WAL');
    dbInstance.pragma('foreign_keys = ON');
    initializeSchema(dbInstance);
    initializeSettings(dbInstance);
  }
  return dbInstance;
}

function initializeSchema(database: Database.Database): void {
  const transaction = database.transaction(() => {
    for (const statement of SCHEMA_STATEMENTS) {
      database.exec(statement);
    }
  });
  transaction();
}

function initializeSettings(database: Database.Database): void {
  const transaction = database.transaction(() => {
    const insertSetting = database.prepare(
      'INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)'
    );
    for (const setting of INITIAL_SETTINGS) {
      insertSetting.run(setting.key, setting.value);
    }
  });
  transaction();
}

export function closeDb(): void {
  if (dbInstance) {
    dbInstance.close();
    dbInstance = null;
  }
}

export function runInTransaction<T>(fn: () => T): T {
  const database = getDb();
  const transaction = database.transaction(fn);
  return transaction();
}
