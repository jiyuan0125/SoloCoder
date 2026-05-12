import { Database } from 'sqlite3';

export const db = new Database('./payment.db');

export const initDB = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    db.serialize(() => {
      db.run(`
        CREATE TABLE IF NOT EXISTS accounts (
          id TEXT PRIMARY KEY,
          currency TEXT NOT NULL,
          balance INTEGER NOT NULL DEFAULT 0,
          created_at INTEGER NOT NULL
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS exchange_rates (
          currency TEXT PRIMARY KEY,
          rate INTEGER NOT NULL,
          updated_at INTEGER NOT NULL
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS transactions (
          id TEXT PRIMARY KEY,
          source_account_id TEXT NOT NULL,
          target_account_id TEXT NOT NULL,
          source_currency TEXT NOT NULL,
          target_currency TEXT NOT NULL,
          source_amount INTEGER NOT NULL,
          target_amount INTEGER,
          exchange_rate_info TEXT,
          status TEXT NOT NULL,
          created_at INTEGER NOT NULL,
          updated_at INTEGER NOT NULL
        )
      `);

      db.run(`
        CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_currency ON accounts(id, currency)
      `);

      db.run(`
        CREATE INDEX IF NOT EXISTS idx_transactions_status_created ON transactions(status, created_at)
      `);

      resolve();
    });
  });
};
