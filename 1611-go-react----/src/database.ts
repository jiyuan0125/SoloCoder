import Database from 'better-sqlite3';
import { CategoryType, FeedbackStatus, Priority } from './types';

const db = new Database('feedback.db');

db.pragma('journal_mode = WAL');

db.exec(`
  CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS feedbacks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    priority TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT '新建',
    assignee_id INTEGER,
    internal_note TEXT NOT NULL DEFAULT '',
    is_timeout INTEGER NOT NULL DEFAULT 0,
    resolved_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories (id)
  );

  CREATE TABLE IF NOT EXISTS status_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feedback_id INTEGER NOT NULL,
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL,
    operator_id INTEGER NOT NULL,
    operator_role TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (feedback_id) REFERENCES feedbacks (id)
  );

  CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feedback_id INTEGER NOT NULL,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (feedback_id) REFERENCES feedbacks (id)
  );

  CREATE INDEX IF NOT EXISTS idx_feedbacks_status ON feedbacks(status);
  CREATE INDEX IF NOT EXISTS idx_feedbacks_priority ON feedbacks(priority);
  CREATE INDEX IF NOT EXISTS idx_feedbacks_user_id ON feedbacks(user_id);
  CREATE INDEX IF NOT EXISTS idx_status_logs_feedback_id ON status_logs(feedback_id);
`);

const now = new Date().toISOString();
const defaultCategories = [
  CategoryType.FEATURE_SUGGESTION,
  CategoryType.BUG_REPORT,
  CategoryType.EXPERIENCE_COMPLAINT,
  CategoryType.OTHER
];

const insertCategory = db.prepare('INSERT OR IGNORE INTO categories (name, created_at) VALUES (?, ?)');
const transaction = db.transaction(() => {
  for (const name of defaultCategories) {
    insertCategory.run(name, now);
  }
});
transaction();

export { db };
