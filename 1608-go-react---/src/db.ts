import * as sqlite3 from 'sqlite3';
import { promisify } from 'util';

const DB_PATH = process.env.DB_PATH || './recommendations.db';

const db = new sqlite3.Database(DB_PATH);

function runAsync(sql: string, ...params: any[]): Promise<{ lastID: number; changes: number }> {
  return new Promise((resolve, reject) => {
    db.run(sql, params, function(err) {
      if (err) {
        reject(err);
      } else {
        resolve({ lastID: this.lastID, changes: this.changes });
      }
    });
  });
}

const getAsync = promisify(db.get.bind(db)) as (sql: string, ...params: any[]) => Promise<any>;
const allAsync = promisify(db.all.bind(db)) as (sql: string, ...params: any[]) => Promise<any[]>;

async function initDb(): Promise<void> {
  await runAsync(`
    CREATE TABLE IF NOT EXISTS products (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      category TEXT NOT NULL,
      avg_score REAL DEFAULT 0,
      total_ratings INTEGER DEFAULT 0,
      created_at TEXT DEFAULT (datetime('now'))
    )
  `);

  await runAsync(`
    CREATE TABLE IF NOT EXISTS ratings (
      user_id INTEGER NOT NULL,
      product_id INTEGER NOT NULL,
      score INTEGER NOT NULL,
      updated_at TEXT DEFAULT (datetime('now')),
      PRIMARY KEY (user_id, product_id),
      FOREIGN KEY (product_id) REFERENCES products(id)
    )
  `);

  await runAsync(`
    CREATE TABLE IF NOT EXISTS recommend_history (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_id INTEGER NOT NULL,
      product_id INTEGER NOT NULL,
      exposed_at TEXT DEFAULT (datetime('now')),
      interacted_at TEXT NULL
    )
  `);

  await runAsync(`
    CREATE TABLE IF NOT EXISTS daily_stats (
      date TEXT PRIMARY KEY,
      coverage REAL DEFAULT 0,
      accuracy REAL DEFAULT 0,
      diversity REAL DEFAULT 0,
      total_exposures INTEGER DEFAULT 0,
      total_interactions INTEGER DEFAULT 0
    )
  `);

  await runAsync(`CREATE INDEX IF NOT EXISTS idx_ratings_user ON ratings(user_id)`);
  await runAsync(`CREATE INDEX IF NOT EXISTS idx_ratings_product ON ratings(product_id)`);
  await runAsync(`CREATE INDEX IF NOT EXISTS idx_history_user ON recommend_history(user_id)`);
  await runAsync(`CREATE INDEX IF NOT EXISTS idx_history_exposed ON recommend_history(exposed_at)`);
}

export { db, runAsync, getAsync, allAsync, initDb };
