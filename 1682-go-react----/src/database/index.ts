import Database from 'better-sqlite3';
import type { Database as DBType } from 'better-sqlite3';

const db: DBType = new Database('proposal-voting.db');

db.exec(`
CREATE TABLE IF NOT EXISTS proposals (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  type TEXT NOT NULL,
  attachment_description TEXT NOT NULL,
  stage TEXT NOT NULL,
  stage_start_time INTEGER NOT NULL,
  created_by TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  is_rerun INTEGER DEFAULT 0,
  original_id TEXT,
  notice_end_time INTEGER,
  voting_end_time INTEGER,
  execution_end_time INTEGER,
  is_suspended INTEGER DEFAULT 0,
  public_notice_end_time INTEGER
);

CREATE TABLE IF NOT EXISTS owners (
  id TEXT PRIMARY KEY,
  phone TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS votes (
  id TEXT PRIMARY KEY,
  proposal_id TEXT NOT NULL,
  phone TEXT NOT NULL,
  option TEXT NOT NULL,
  voted_at INTEGER NOT NULL,
  UNIQUE(proposal_id, phone)
);

CREATE TABLE IF NOT EXISTS todos (
  id TEXT PRIMARY KEY,
  proposal_id TEXT NOT NULL,
  type TEXT NOT NULL,
  assignee TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  handled_at INTEGER,
  escalated_at INTEGER
);

CREATE TABLE IF NOT EXISTS suggestions (
  id TEXT PRIMARY KEY,
  proposal_id TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS objections (
  id TEXT PRIMARY KEY,
  proposal_id TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  reason TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  UNIQUE(proposal_id, owner_id)
);

CREATE INDEX IF NOT EXISTS idx_proposals_stage ON proposals(stage);
CREATE INDEX IF NOT EXISTS idx_votes_proposal ON votes(proposal_id);
CREATE INDEX IF NOT EXISTS idx_todos_assignee ON todos(assignee, status);
CREATE INDEX IF NOT EXISTS idx_todos_proposal ON todos(proposal_id);
`);

export default db;
