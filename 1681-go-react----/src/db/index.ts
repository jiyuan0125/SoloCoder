import sqlite3 from 'sqlite3';
import { RequestStatus } from '../types';
import { getCurrentTime } from '../utils/id';

const DATABASE_PATH = process.env.DB_PATH || './data.db';

const db = new sqlite3.Database(DATABASE_PATH, (err) => {
  if (err) {
    console.error('Database connection error:', err.message);
  } else {
    console.log('Connected to SQLite database');
    initializeTables();
    scheduleCleanup();
  }
});

function runSQL(sql: string, params: any[] = []): Promise<void> {
  return new Promise((resolve, reject) => {
    db.run(sql, params, (err) => {
      if (err) reject(err);
      else resolve();
    });
  });
}

function getOne<T>(sql: string, params: any[] = []): Promise<T | undefined> {
  return new Promise((resolve, reject) => {
    db.get(sql, params, (err, row: T) => {
      if (err) reject(err);
      else resolve(row);
    });
  });
}

function getAll<T>(sql: string, params: any[] = []): Promise<T[]> {
  return new Promise((resolve, reject) => {
    db.all(sql, params, (err, rows: T[]) => {
      if (err) reject(err);
      else resolve(rows);
    });
  });
}

function initializeTables(): void {
  const tables = [
    `CREATE TABLE IF NOT EXISTS users (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      building TEXT NOT NULL,
      avg_rating REAL DEFAULT 0,
      rating_count INTEGER DEFAULT 0,
      cool_down_until INTEGER,
      created_at INTEGER NOT NULL
    )`,
    `CREATE TABLE IF NOT EXISTS requests (
      id TEXT PRIMARY KEY,
      publisher_id TEXT NOT NULL,
      title TEXT NOT NULL,
      category TEXT NOT NULL,
      description TEXT,
      expected_time INTEGER NOT NULL,
      willing_to_pay INTEGER NOT NULL DEFAULT 0,
      status TEXT NOT NULL DEFAULT '待接单',
      assignee_id TEXT,
      completed_at INTEGER,
      created_at INTEGER NOT NULL,
      expired_at INTEGER,
      archived_at INTEGER,
      FOREIGN KEY (publisher_id) REFERENCES users(id),
      FOREIGN KEY (assignee_id) REFERENCES users(id)
    )`,
    `CREATE TABLE IF NOT EXISTS ratings (
      id TEXT PRIMARY KEY,
      request_id TEXT NOT NULL,
      from_user_id TEXT NOT NULL,
      to_user_id TEXT NOT NULL,
      score INTEGER NOT NULL,
      created_at INTEGER NOT NULL,
      FOREIGN KEY (request_id) REFERENCES requests(id),
      FOREIGN KEY (from_user_id) REFERENCES users(id),
      FOREIGN KEY (to_user_id) REFERENCES users(id)
    )`,
    `CREATE TABLE IF NOT EXISTS month_stats (
      id TEXT PRIMARY KEY,
      year INTEGER NOT NULL,
      month INTEGER NOT NULL,
      publish_count INTEGER DEFAULT 0,
      match_success_count INTEGER DEFAULT 0,
      total_complete_time INTEGER DEFAULT 0,
      complete_count INTEGER DEFAULT 0,
      UNIQUE(year, month)
    )`,
    `CREATE TABLE IF NOT EXISTS building_activities (
      id TEXT PRIMARY KEY,
      building TEXT NOT NULL,
      year INTEGER NOT NULL,
      month INTEGER NOT NULL,
      publish_count INTEGER DEFAULT 0,
      accept_count INTEGER DEFAULT 0,
      score REAL DEFAULT 0,
      UNIQUE(building, year, month)
    )`,
    `CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(status)`,
    `CREATE INDEX IF NOT EXISTS idx_requests_expected_time ON requests(expected_time)`,
    `CREATE INDEX IF NOT EXISTS idx_ratings_to_user ON ratings(to_user_id)`,
  ];

  tables.forEach((sql) => {
    db.run(sql, (err) => {
      if (err) console.error('Table creation error:', err.message);
    });
  });
}

function scheduleCleanup(): void {
  setInterval(() => {
    checkExpiredRequests();
    autoArchiveExpiredRequests();
    autoCompleteRatings();
  }, 60 * 1000);
}

async function checkExpiredRequests(): Promise<void> {
  const now = getCurrentTime();
  const sql = `
    UPDATE requests 
    SET status = ?, expired_at = ?
    WHERE status = ? AND expected_time < ?
  `;
  await runSQL(sql, [RequestStatus.已过期, now, RequestStatus.待接单, now]);
}

async function autoArchiveExpiredRequests(): Promise<void> {
  const now = getCurrentTime();
  const fortyEightHours = 48 * 60 * 60 * 1000;
  const sql = `
    UPDATE requests 
    SET status = ?, archived_at = ?
    WHERE status = ? AND expired_at < ?
  `;
  await runSQL(sql, [RequestStatus.已归档, now, RequestStatus.已过期, now - fortyEightHours]);
}

async function autoCompleteRatings(): Promise<void> {
  const now = getCurrentTime();
  const seventyTwoHours = 72 * 60 * 60 * 1000;
  
  const completedRequests = await getAll<{
    id: string;
    publisher_id: string;
    assignee_id: string;
    completed_at: number;
  }>(`
    SELECT id, publisher_id, assignee_id, completed_at
    FROM requests
    WHERE status = ? AND completed_at < ?
    AND completed_at IS NOT NULL
  `, [RequestStatus.已完成, now - seventyTwoHours]);

  for (const req of completedRequests) {
    await processAutoRating(req);
  }
}

async function processAutoRating(req: {
  id: string;
  publisher_id: string;
  assignee_id: string;
  completed_at: number;
}): Promise<void> {
  const publisherRating = await getOne(
    'SELECT id FROM ratings WHERE request_id = ? AND from_user_id = ?',
    [req.id, req.assignee_id]
  );
  
  const assigneeRating = await getOne(
    'SELECT id FROM ratings WHERE request_id = ? AND from_user_id = ?',
    [req.id, req.publisher_id]
  );

  const { generateId } = await import('../utils/id');
  const now = getCurrentTime();

  if (!publisherRating) {
    await runSQL(
      'INSERT INTO ratings (id, request_id, from_user_id, to_user_id, score, created_at) VALUES (?, ?, ?, ?, ?, ?)',
      [generateId(), req.id, req.assignee_id, req.publisher_id, 5, now]
    );
    await updateUserAvgRating(req.publisher_id);
  }

  if (!assigneeRating) {
    await runSQL(
      'INSERT INTO ratings (id, request_id, from_user_id, to_user_id, score, created_at) VALUES (?, ?, ?, ?, ?, ?)',
      [generateId(), req.id, req.publisher_id, req.assignee_id, 5, now]
    );
    await updateUserAvgRating(req.assignee_id);
  }
}

async function updateUserAvgRating(userId: string): Promise<void> {
  const result = await getOne<{ avg_score: number; count: number }>(
    'SELECT AVG(score) as avg_score, COUNT(*) as count FROM ratings WHERE to_user_id = ?',
    [userId]
  );
  if (result) {
    await runSQL(
      'UPDATE users SET avg_rating = ?, rating_count = ? WHERE id = ?',
      [result.avg_score || 0, result.count || 0, userId]
    );
  }
}

export {
  db,
  runSQL,
  getOne,
  getAll,
  updateUserAvgRating,
};
