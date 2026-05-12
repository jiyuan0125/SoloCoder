import sqlite3 from 'sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '../../alerts.db');

const db = new sqlite3.Database(dbPath, (err) => {
  if (err) {
    console.error('数据库连接失败:', err.message);
  } else {
    console.log('已连接到 SQLite 数据库');
  }
});

export const initDatabase = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    db.serialize(() => {
      db.run(`
        CREATE TABLE IF NOT EXISTS alerts (
          id TEXT PRIMARY KEY,
          name TEXT NOT NULL,
          level TEXT NOT NULL,
          sourceSystem TEXT NOT NULL,
          description TEXT,
          metrics TEXT,
          status TEXT NOT NULL,
          originalLevel TEXT NOT NULL,
          count INTEGER NOT NULL,
          createdAt INTEGER NOT NULL,
          lastOccurrenceAt INTEGER NOT NULL,
          acknowledgedAt INTEGER,
          resolvedAt INTEGER,
          escalationTime INTEGER NOT NULL
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS aggregated_alerts (
          id TEXT PRIMARY KEY,
          alertKey TEXT NOT NULL UNIQUE,
          name TEXT NOT NULL,
          level TEXT NOT NULL,
          sourceSystem TEXT NOT NULL,
          count INTEGER NOT NULL,
          descriptions TEXT NOT NULL,
          metricsList TEXT NOT NULL,
          windowEndTime INTEGER NOT NULL,
          createdAt INTEGER NOT NULL
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS channel_configs (
          type TEXT PRIMARY KEY,
          enabled INTEGER NOT NULL,
          rateLimitPerMinute INTEGER NOT NULL,
          emailRecipients TEXT,
          smsNumbers TEXT,
          webhookUrl TEXT
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS notification_queue (
          id TEXT PRIMARY KEY,
          channelType TEXT NOT NULL,
          content TEXT NOT NULL,
          priority INTEGER NOT NULL,
          scheduledTime INTEGER NOT NULL,
          createdAt INTEGER NOT NULL,
          sent INTEGER NOT NULL,
          alertId TEXT
        )
      `);

      const defaultConfigs = [
        { type: 'email', enabled: 1, rateLimitPerMinute: 10, emailRecipients: JSON.stringify(['admin@example.com']) },
        { type: 'sms', enabled: 1, rateLimitPerMinute: 5, smsNumbers: JSON.stringify(['+1234567890']) },
        { type: 'webhook', enabled: 0, rateLimitPerMinute: 20, webhookUrl: 'http://localhost:8080/webhook' }
      ];

      const insertStmt = db.prepare(
        `INSERT OR IGNORE INTO channel_configs (type, enabled, rateLimitPerMinute, emailRecipients, smsNumbers, webhookUrl)
         VALUES (?, ?, ?, ?, ?, ?)`
      );

      defaultConfigs.forEach(config => {
        insertStmt.run(
          config.type,
          config.enabled,
          config.rateLimitPerMinute,
          config.emailRecipients || null,
          config.smsNumbers || null,
          config.webhookUrl || null
        );
      });

      insertStmt.finalize((err) => {
        if (err) {
          reject(err);
        } else {
          resolve();
        }
      });
    });
  });
};

export const run = (sql: string, params: any[] = []): Promise<{ lastID: number; changes: number }> => {
  return new Promise((resolve, reject) => {
    db.run(sql, params, function (err) {
      if (err) {
        reject(err);
      } else {
        resolve({ lastID: this.lastID, changes: this.changes });
      }
    });
  });
};

export const get = <T>(sql: string, params: any[] = []): Promise<T | null> => {
  return new Promise((resolve, reject) => {
    db.get(sql, params, (err, row) => {
      if (err) {
        reject(err);
      } else {
        resolve(row as T || null);
      }
    });
  });
};

export const all = <T>(sql: string, params: any[] = []): Promise<T[]> => {
  return new Promise((resolve, reject) => {
    db.all(sql, params, (err, rows) => {
      if (err) {
        reject(err);
      } else {
        resolve(rows as T[]);
      }
    });
  });
};

export default db;
