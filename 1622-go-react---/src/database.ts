import sqlite3 from 'sqlite3';
import path from 'path';

const DB_PATH = path.join(__dirname, '..', 'logs.db');

export const db = new sqlite3.Database(DB_PATH, (err) => {
  if (err) {
    console.error('Error opening database:', err.message);
    process.exit(1);
  }
  console.log('Connected to SQLite database');
});

db.serialize(() => {
  db.run(`
    CREATE TABLE IF NOT EXISTS logs (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      timestamp INTEGER NOT NULL,
      level TEXT NOT NULL,
      serviceName TEXT NOT NULL,
      traceId TEXT NOT NULL,
      message TEXT NOT NULL,
      tags TEXT NOT NULL
    )
  `);

  db.run('CREATE INDEX IF NOT EXISTS idx_timestamp ON logs(timestamp)');
  db.run('CREATE INDEX IF NOT EXISTS idx_level ON logs(level)');
  db.run('CREATE INDEX IF NOT EXISTS idx_serviceName ON logs(serviceName)');
  db.run('CREATE INDEX IF NOT EXISTS idx_traceId ON logs(traceId)');
});

export function closeDatabase(): void {
  db.close((err) => {
    if (err) {
      console.error('Error closing database:', err.message);
    } else {
      console.log('Database connection closed');
    }
  });
}
