import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.resolve(process.cwd(), 'auth-center.db');
export const db = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

db.exec(`
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  mfa_enabled INTEGER DEFAULT 0,
  mfa_secret TEXT,
  mfa_failed_attempts INTEGER DEFAULT 0,
  mfa_frozen_until INTEGER DEFAULT 0,
  login_failed_attempts INTEGER DEFAULT 0,
  login_frozen_until INTEGER DEFAULT 0,
  created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL,
  app_id TEXT NOT NULL,
  callback_url TEXT,
  token TEXT,
  created_at INTEGER DEFAULT (strftime('%s', 'now')),
  expires_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS apps (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  callback_url TEXT NOT NULL,
  secret TEXT NOT NULL,
  created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE TABLE IF NOT EXISTS auth_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER,
  username TEXT,
  action TEXT NOT NULL,
  success INTEGER NOT NULL,
  ip TEXT,
  user_agent TEXT,
  created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_app ON sessions(user_id, app_id);
CREATE INDEX IF NOT EXISTS idx_auth_logs_user ON auth_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
`);

export const defaultApps = [
  {
    id: 'app1',
    name: 'Test App 1',
    callback_url: 'http://localhost:3001/callback',
    secret: 'app1-secret-key-12345'
  },
  {
    id: 'app2',
    name: 'Test App 2',
    callback_url: 'http://localhost:3002/callback',
    secret: 'app2-secret-key-67890'
  }
];

for (const app of defaultApps) {
  const existing = db.prepare('SELECT id FROM apps WHERE id = ?').get(app.id);
  if (!existing) {
    db.prepare(`
      INSERT INTO apps (id, name, callback_url, secret)
      VALUES (?, ?, ?, ?)
    `).run(app.id, app.name, app.callback_url, app.secret);
  }
}
