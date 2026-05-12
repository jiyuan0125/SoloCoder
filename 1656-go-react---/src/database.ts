import sqlite3 from 'sqlite3';
import path from 'path';
import { TrustRecord, TrustStatus } from './types';

const DB_PATH = path.join(__dirname, '../device-trust.db');

export class Database {
  private static instance: sqlite3.Database;

  static init(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.instance = new sqlite3.Database(DB_PATH, (err) => {
        if (err) {
          reject(err);
          return;
        }
        this.initTables()
          .then(resolve)
          .catch(reject);
      });
    });
  }

  private static initTables(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.instance.run(
        `CREATE TABLE IF NOT EXISTS trust_records (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          user_id TEXT NOT NULL,
          device_fingerprint TEXT NOT NULL,
          first_trusted_at TEXT NOT NULL,
          last_used_at TEXT NOT NULL,
          status TEXT NOT NULL DEFAULT 'trusted',
          UNIQUE(user_id, device_fingerprint)
        )`,
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  }

  static close(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.instance.close((err) => {
        if (err) reject(err);
        else resolve();
      });
    });
  }

  static getTrustRecord(userId: string, deviceFingerprint: string): Promise<TrustRecord | null> {
    return new Promise((resolve, reject) => {
      this.instance.get(
        'SELECT * FROM trust_records WHERE user_id = ? AND device_fingerprint = ?',
        [userId, deviceFingerprint],
        (err, row: any) => {
          if (err) reject(err);
          else if (!row) resolve(null);
          else resolve({
            id: row.id,
            userId: row.user_id,
            deviceFingerprint: row.device_fingerprint,
            firstTrustedAt: row.first_trusted_at,
            lastUsedAt: row.last_used_at,
            status: row.status as TrustStatus,
          });
        }
      );
    });
  }

  static getAllTrustRecords(): Promise<TrustRecord[]> {
    return new Promise((resolve, reject) => {
      this.instance.all(
        'SELECT * FROM trust_records',
        (err, rows: any[]) => {
          if (err) reject(err);
          else resolve(rows.map((row) => ({
            id: row.id,
            userId: row.user_id,
            deviceFingerprint: row.device_fingerprint,
            firstTrustedAt: row.first_trusted_at,
            lastUsedAt: row.last_used_at,
            status: row.status as TrustStatus,
          })));
        }
      );
    });
  }

  static createTrustRecord(userId: string, deviceFingerprint: string): Promise<void> {
    const now = new Date().toISOString();
    return new Promise((resolve, reject) => {
      this.instance.run(
        `INSERT INTO trust_records (user_id, device_fingerprint, first_trusted_at, last_used_at, status)
         VALUES (?, ?, ?, ?, ?)`,
        [userId, deviceFingerprint, now, now, TrustStatus.TRUSTED],
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  }

  static updateTrustRecordStatus(userId: string, deviceFingerprint: string, status: TrustStatus): Promise<void> {
    const now = new Date().toISOString();
    return new Promise((resolve, reject) => {
      this.instance.run(
        `UPDATE trust_records SET last_used_at = ?, status = ?
         WHERE user_id = ? AND device_fingerprint = ?`,
        [now, status, userId, deviceFingerprint],
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  }

  static revokeTrustRecordById(id: number): Promise<void> {
    return new Promise((resolve, reject) => {
      this.instance.run(
        'UPDATE trust_records SET status = ? WHERE id = ?',
        [TrustStatus.REVOKED, id],
        function (this: any, err) {
          if (err) reject(err);
          else if (this.changes === 0) reject(new Error('RecordNotFound'));
          else resolve();
        }
      );
    });
  }

  static getTrustRecordById(id: number): Promise<TrustRecord | null> {
    return new Promise((resolve, reject) => {
      this.instance.get(
        'SELECT * FROM trust_records WHERE id = ?',
        [id],
        (err, row: any) => {
          if (err) reject(err);
          else if (!row) resolve(null);
          else resolve({
            id: row.id,
            userId: row.user_id,
            deviceFingerprint: row.device_fingerprint,
            firstTrustedAt: row.first_trusted_at,
            lastUsedAt: row.last_used_at,
            status: row.status as TrustStatus,
          });
        }
      );
    });
  }
}
