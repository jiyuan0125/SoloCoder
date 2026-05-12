import * as sqlite3 from 'sqlite3';
import * as path from 'path';

const dbPath = path.join(process.cwd(), 'incidents.db');

export const db = new sqlite3.Database(dbPath, (err) => {
  if (err) {
    console.error('Failed to open database:', err.message);
  } else {
    console.log('Connected to SQLite database:', dbPath);
  }
});

db.serialize(() => {
  db.run(`
    CREATE TABLE IF NOT EXISTS incidents (
      id TEXT PRIMARY KEY,
      incident_number TEXT NOT NULL UNIQUE,
      service TEXT NOT NULL,
      start_time TEXT NOT NULL,
      discovered_time TEXT NOT NULL,
      recovered_time TEXT,
      severity TEXT NOT NULL,
      status TEXT NOT NULL,
      impact_scope TEXT NOT NULL,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    )
  `);

  db.run(`CREATE INDEX IF NOT EXISTS idx_incidents_service ON incidents(service)`);
  db.run(`CREATE INDEX IF NOT EXISTS idx_incidents_start_time ON incidents(start_time)`);
  db.run(`CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status)`);
});

export function runQuery<T = any>(sql: string, params: any[] = []): Promise<T[]> {
  return new Promise((resolve, reject) => {
    db.all(sql, params, (err, rows) => {
      if (err) reject(err);
      else resolve(rows as T[]);
    });
  });
}

export function runExec(sql: string, params: any[] = []): Promise<{ lastID: number; changes: number }> {
  return new Promise((resolve, reject) => {
    db.run(sql, params, function (err) {
      if (err) reject(err);
      else resolve({ lastID: this.lastID, changes: this.changes });
    });
  });
}
