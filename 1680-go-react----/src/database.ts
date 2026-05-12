import Database from 'better-sqlite3';
import type { Database as DatabaseType } from 'better-sqlite3';
import path from 'path';
import { UserRole, IssueStatus, VoteType, VoteOption } from './types';

const dbPath = path.join(process.cwd(), 'voting.db');
const db: DatabaseType = new Database(dbPath);

db.pragma('journal_mode = WAL');

const initDatabase = () => {
  db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      username TEXT UNIQUE NOT NULL,
      name TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'owner',
      is_registered INTEGER NOT NULL DEFAULT 0,
      is_verified INTEGER NOT NULL DEFAULT 0,
      house_area REAL NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS issues (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      title TEXT NOT NULL,
      content TEXT NOT NULL,
      category TEXT NOT NULL,
      attachments TEXT NOT NULL DEFAULT '',
      initiator_id INTEGER NOT NULL,
      status TEXT NOT NULL DEFAULT 'draft',
      publicity_start_at TEXT,
      publicity_end_at TEXT,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL,
      FOREIGN KEY (initiator_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS opinions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      issue_id INTEGER NOT NULL,
      user_id INTEGER NOT NULL,
      content TEXT NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (issue_id) REFERENCES issues(id),
      FOREIGN KEY (user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS vote_settings (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      issue_id INTEGER UNIQUE NOT NULL,
      start_at TEXT NOT NULL,
      end_at TEXT NOT NULL,
      vote_type TEXT NOT NULL,
      min_participation_rate REAL NOT NULL,
      created_at TEXT NOT NULL,
      FOREIGN KEY (issue_id) REFERENCES issues(id)
    );

    CREATE TABLE IF NOT EXISTS votes (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      issue_id INTEGER NOT NULL,
      user_id INTEGER NOT NULL,
      option TEXT NOT NULL,
      weight INTEGER NOT NULL,
      voted_at TEXT NOT NULL,
      UNIQUE(issue_id, user_id),
      FOREIGN KEY (issue_id) REFERENCES issues(id),
      FOREIGN KEY (user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS todos (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      issue_id INTEGER NOT NULL,
      user_id INTEGER NOT NULL,
      type TEXT NOT NULL,
      deadline_at TEXT NOT NULL,
      is_completed INTEGER NOT NULL DEFAULT 0,
      is_reminded INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL,
      FOREIGN KEY (issue_id) REFERENCES issues(id),
      FOREIGN KEY (user_id) REFERENCES users(id)
    );
  `);

  const userCount = db.prepare('SELECT COUNT(*) as count FROM users').get() as { count: number };
  if (userCount.count === 0) {
    const insertUser = db.prepare(`
      INSERT INTO users (username, name, role, is_registered, is_verified, house_area, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);

    const now = new Date().toISOString();
    
    insertUser.run('owner1', '业主张三', UserRole.OWNER, 1, 1, 120.5, now);
    insertUser.run('owner2', '业主李四', UserRole.OWNER, 1, 1, 80.0, now);
    insertUser.run('owner3', '业主王五', UserRole.OWNER, 1, 0, 95.5, now);
    insertUser.run('owner4', '业主赵六', UserRole.OWNER, 0, 0, 110.0, now);
    insertUser.run('committee1', '业委会主任', UserRole.COMMITTEE, 1, 1, 100.0, now);
    insertUser.run('executor1', '执行人', UserRole.EXECUTOR, 1, 1, 75.5, now);
  }
};

initDatabase();

export { db };
