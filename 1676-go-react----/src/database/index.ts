import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '../../crowdfunding.db');
export const db = new Database(dbPath);

db.exec(`
  CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    category TEXT NOT NULL,
    target_amount INTEGER NOT NULL,
    raised_amount INTEGER NOT NULL DEFAULT 0,
    deadline TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    succeeded_at TEXT
  );

  CREATE TABLE IF NOT EXISTS reward_tiers (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS support_records (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    reward_tier_id TEXT NOT NULL,
    supporter_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    refunded_at TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (reward_tier_id) REFERENCES reward_tiers(id) ON DELETE CASCADE,
    UNIQUE(supporter_id, reward_tier_id)
  );
`);

export default db;
