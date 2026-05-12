import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '..', 'affiliate.db');
const db = new Database(dbPath);

// 启用外键约束
db.pragma('foreign_keys = ON');

// 初始化数据库表
export function initializeDatabase(): void {
  // 渠道表
  db.exec(`
    CREATE TABLE IF NOT EXISTS channels (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL UNIQUE,
      contact TEXT NOT NULL,
      settlement_method TEXT NOT NULL CHECK (settlement_method IN ('monthly', 'threshold')),
      commission_rate REAL NOT NULL CHECK (commission_rate >= 0 AND commission_rate <= 1),
      status TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now'))
    )
  `);

  // 推广链接表
  db.exec(`
    CREATE TABLE IF NOT EXISTS links (
      id TEXT PRIMARY KEY,
      channel_id TEXT NOT NULL,
      material_type TEXT NOT NULL CHECK (material_type IN ('homepage', 'campaign', 'product')),
      unique_code TEXT NOT NULL UNIQUE,
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      FOREIGN KEY (channel_id) REFERENCES channels(id)
    )
  `);

  // 用户归因记录表
  db.exec(`
    CREATE TABLE IF NOT EXISTS attributions (
      id TEXT PRIMARY KEY,
      user_id TEXT NOT NULL,
      channel_id TEXT NOT NULL,
      link_id TEXT NOT NULL,
      click_time TEXT NOT NULL DEFAULT (datetime('now')),
      attribution_window_end TEXT NOT NULL,
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      FOREIGN KEY (channel_id) REFERENCES channels(id),
      FOREIGN KEY (link_id) REFERENCES links(id)
    )
  `);

  // 订单表
  db.exec(`
    CREATE TABLE IF NOT EXISTS orders (
      id TEXT PRIMARY KEY,
      user_id TEXT NOT NULL,
      amount INTEGER NOT NULL,
      status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'refunded')),
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now'))
    )
  `);

  // 佣金表
  db.exec(`
    CREATE TABLE IF NOT EXISTS commissions (
      id TEXT PRIMARY KEY,
      order_id TEXT NOT NULL UNIQUE,
      channel_id TEXT NOT NULL,
      amount INTEGER NOT NULL,
      status TEXT NOT NULL CHECK (status IN ('pending', 'settled', 'cancelled')),
      settlement_month TEXT,
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now')),
      FOREIGN KEY (order_id) REFERENCES orders(id),
      FOREIGN KEY (channel_id) REFERENCES channels(id)
    )
  `);

  // 创建索引
  db.exec(`CREATE INDEX IF NOT EXISTS idx_attributions_user_id ON attributions(user_id)`);
  db.exec(`CREATE INDEX IF NOT EXISTS idx_attributions_click_time ON attributions(click_time)`);
  db.exec(`CREATE INDEX IF NOT EXISTS idx_commissions_channel_id ON commissions(channel_id)`);
  db.exec(`CREATE INDEX IF NOT EXISTS idx_commissions_settlement_month ON commissions(settlement_month)`);
}

export default db;
