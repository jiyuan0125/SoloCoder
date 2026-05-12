import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.resolve(process.cwd(), 'data.db');
const db = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

export function initDatabase() {
  db.exec(`
    CREATE TABLE IF NOT EXISTS communities (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL UNIQUE,
      type TEXT NOT NULL,
      max_members INTEGER NOT NULL,
      current_members INTEGER NOT NULL DEFAULT 0,
      created_at INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS community_members (
      id TEXT PRIMARY KEY,
      community_id TEXT NOT NULL,
      user_id TEXT NOT NULL,
      joined_at INTEGER NOT NULL,
      FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
      UNIQUE(community_id, user_id)
    );

    CREATE TABLE IF NOT EXISTS member_records (
      id TEXT PRIMARY KEY,
      community_id TEXT NOT NULL,
      user_id TEXT NOT NULL,
      action TEXT NOT NULL,
      timestamp INTEGER NOT NULL,
      FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS users (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      last_active_at INTEGER NOT NULL,
      created_at INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS tags (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      category TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      UNIQUE(name, category)
    );

    CREATE TABLE IF NOT EXISTS user_tags (
      id TEXT PRIMARY KEY,
      user_id TEXT NOT NULL,
      tag_id TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
      FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
      UNIQUE(user_id, tag_id)
    );

    CREATE TABLE IF NOT EXISTS broadcast_tasks (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      content TEXT NOT NULL,
      scheduled_at INTEGER NOT NULL,
      status TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      tier_filter TEXT,
      tag_filter_mode TEXT
    );

    CREATE TABLE IF NOT EXISTS task_target_communities (
      id TEXT PRIMARY KEY,
      task_id TEXT NOT NULL,
      community_id TEXT NOT NULL,
      send_result TEXT,
      failure_reason TEXT,
      sent_at INTEGER,
      FOREIGN KEY (task_id) REFERENCES broadcast_tasks(id) ON DELETE CASCADE,
      UNIQUE(task_id, community_id)
    );

    CREATE TABLE IF NOT EXISTS task_tag_filters (
      id TEXT PRIMARY KEY,
      task_id TEXT NOT NULL,
      tag_id TEXT NOT NULL,
      FOREIGN KEY (task_id) REFERENCES broadcast_tasks(id) ON DELETE CASCADE,
      FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
      UNIQUE(task_id, tag_id)
    );

    CREATE INDEX IF NOT EXISTS idx_community_members_user ON community_members(user_id);
    CREATE INDEX IF NOT EXISTS idx_member_records_user ON member_records(user_id);
    CREATE INDEX IF NOT EXISTS idx_user_tags_tag ON user_tags(tag_id);
    CREATE INDEX IF NOT EXISTS idx_broadcast_tasks_status ON broadcast_tasks(status);
    CREATE INDEX IF NOT EXISTS idx_broadcast_tasks_scheduled ON broadcast_tasks(scheduled_at);
    CREATE INDEX IF NOT EXISTS idx_task_targets_task ON task_target_communities(task_id);
  `);
}

export default db;
