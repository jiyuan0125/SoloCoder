import Database, { Database as DatabaseType } from 'better-sqlite3';

const db: DatabaseType = new Database('./risk-engine.db');

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS rules (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      condition TEXT NOT NULL,
      weight INTEGER NOT NULL,
      priority INTEGER NOT NULL,
      enabled INTEGER NOT NULL DEFAULT 1,
      createdAt INTEGER NOT NULL,
      updatedAt INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS assessments (
      id TEXT PRIMARY KEY,
      transactionId TEXT NOT NULL,
      score INTEGER NOT NULL,
      status TEXT NOT NULL,
      matchedRuleIds TEXT NOT NULL,
      assessedAt INTEGER NOT NULL,
      accountFrozen INTEGER NOT NULL DEFAULT 0,
      frozenAttempted INTEGER NOT NULL DEFAULT 0,
      freezeFailed INTEGER NOT NULL DEFAULT 0,
      UNIQUE(transactionId)
    );

    CREATE TABLE IF NOT EXISTS reviews (
      id TEXT PRIMARY KEY,
      assessmentId TEXT NOT NULL,
      transactionId TEXT NOT NULL,
      reviewerId TEXT,
      decision TEXT NOT NULL DEFAULT 'PENDING',
      decidedAt INTEGER,
      createdAt INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS accounts (
      id TEXT PRIMARY KEY,
      userId TEXT NOT NULL,
      frozen INTEGER NOT NULL DEFAULT 0,
      frozenAt INTEGER,
      UNIQUE(userId)
    );

    CREATE INDEX IF NOT EXISTS idx_rules_priority ON rules(priority);
    CREATE INDEX IF NOT EXISTS idx_assessments_transaction ON assessments(transactionId);
    CREATE INDEX IF NOT EXISTS idx_reviews_assessment ON reviews(assessmentId);
    CREATE INDEX IF NOT EXISTS idx_reviews_decision ON reviews(decision);
  `);
}

initDatabase();

export default db;
