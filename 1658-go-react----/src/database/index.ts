import Database from 'better-sqlite3';
import path from 'path';

const DB_PATH = path.join(__dirname, '../../data/fraud.db');

let db: Database.Database;

export function getDatabase(): Database.Database {
  if (!db) {
    db = new Database(DB_PATH);
    db.pragma('journal_mode = WAL');
    db.pragma('foreign_keys = ON');
    initializeTables(db);
  }
  return db;
}

function initializeTables(database: Database.Database): void {
  database.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      email TEXT NOT NULL UNIQUE,
      phone TEXT NOT NULL,
      riskTags TEXT NOT NULL DEFAULT '[]',
      riskScore INTEGER NOT NULL DEFAULT 0,
      isBlacklisted INTEGER NOT NULL DEFAULT 0,
      createdAt INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS devices (
      id TEXT PRIMARY KEY,
      fingerprintId TEXT NOT NULL UNIQUE,
      isBlacklisted INTEGER NOT NULL DEFAULT 0,
      createdAt INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS device_user_associations (
      deviceId TEXT NOT NULL,
      userId TEXT NOT NULL,
      createdAt INTEGER NOT NULL,
      PRIMARY KEY (deviceId, userId),
      FOREIGN KEY (deviceId) REFERENCES devices(id) ON DELETE CASCADE,
      FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS orders (
      id TEXT PRIMARY KEY,
      userId TEXT NOT NULL,
      deviceId TEXT NOT NULL,
      productInfo TEXT NOT NULL,
      amount REAL NOT NULL,
      shippingAddress TEXT NOT NULL,
      paymentMethod TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT 'pending',
      riskLevel TEXT NOT NULL DEFAULT 'low',
      createdAt INTEGER NOT NULL,
      FOREIGN KEY (userId) REFERENCES users(id),
      FOREIGN KEY (deviceId) REFERENCES devices(id)
    );

    CREATE INDEX IF NOT EXISTS idx_orders_userId ON orders(userId);
    CREATE INDEX IF NOT EXISTS idx_orders_deviceId ON orders(deviceId);
    CREATE INDEX IF NOT EXISTS idx_device_user_associations_userId ON device_user_associations(userId);
    CREATE INDEX IF NOT EXISTS idx_device_user_associations_deviceId ON device_user_associations(deviceId);
    CREATE INDEX IF NOT EXISTS idx_users_isBlacklisted ON users(isBlacklisted);
    CREATE INDEX IF NOT EXISTS idx_devices_isBlacklisted ON devices(isBlacklisted);
  `);
}

export function closeDatabase(): void {
  if (db) {
    db.close();
  }
}

export { Database };
