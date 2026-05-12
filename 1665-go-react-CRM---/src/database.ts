import Database from 'better-sqlite3';
import { v4 as uuidv4 } from 'uuid';

const db = new Database('./crm.db');

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

db.exec(`
  CREATE TABLE IF NOT EXISTS sales_persons (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS leads (
    id TEXT PRIMARY KEY,
    source_channel TEXT NOT NULL,
    contact_info TEXT NOT NULL,
    company_name TEXT,
    contact_person TEXT,
    requirements TEXT,
    status TEXT NOT NULL DEFAULT 'new',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    assigned_sales_id TEXT,
    is_in_pool INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (assigned_sales_id) REFERENCES sales_persons(id)
  );

  CREATE TABLE IF NOT EXISTS lead_status_history (
    id TEXT PRIMARY KEY,
    lead_id TEXT NOT NULL,
    from_status TEXT,
    to_status TEXT NOT NULL,
    changed_by TEXT,
    reason TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (lead_id) REFERENCES leads(id),
    FOREIGN KEY (changed_by) REFERENCES sales_persons(id)
  );

  CREATE TABLE IF NOT EXISTS opportunities (
    id TEXT PRIMARY KEY,
    lead_id TEXT NOT NULL,
    expected_amount REAL NOT NULL,
    assigned_sales_id TEXT NOT NULL,
    stage TEXT NOT NULL DEFAULT 'initial_contact',
    expected_close_date TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (lead_id) REFERENCES leads(id),
    FOREIGN KEY (assigned_sales_id) REFERENCES sales_persons(id)
  );

  CREATE TABLE IF NOT EXISTS opportunity_stage_history (
    id TEXT PRIMARY KEY,
    opportunity_id TEXT NOT NULL,
    from_stage TEXT,
    to_stage TEXT NOT NULL,
    changed_by TEXT,
    reason TEXT,
    lost_to TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (opportunity_id) REFERENCES opportunities(id),
    FOREIGN KEY (changed_by) REFERENCES sales_persons(id)
  );

  CREATE TABLE IF NOT EXISTS sales_pool (
    lead_id TEXT PRIMARY KEY,
    entered_at INTEGER NOT NULL,
    reason TEXT,
    FOREIGN KEY (lead_id) REFERENCES leads(id)
  );

  CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
  CREATE INDEX IF NOT EXISTS idx_leads_assigned ON leads(assigned_sales_id);
  CREATE INDEX IF NOT EXISTS idx_leads_pool ON leads(is_in_pool);
  CREATE INDEX IF NOT EXISTS idx_opportunities_stage ON opportunities(stage);
  CREATE INDEX IF NOT EXISTS idx_opportunities_sales ON opportunities(assigned_sales_id);
`);

const insertSalesPerson = db.prepare(`
  INSERT INTO sales_persons (id, name, email, created_at)
  VALUES (?, ?, ?, ?)
`);

const existingSales = db.prepare('SELECT COUNT(*) as count FROM sales_persons').get() as { count: number };
if (existingSales.count === 0) {
  const now = Date.now();
  insertSalesPerson.run(uuidv4(), '销售张三', 'zhangsan@example.com', now);
  insertSalesPerson.run(uuidv4(), '销售李四', 'lisi@example.com', now);
  insertSalesPerson.run(uuidv4(), '销售王五', 'wangwu@example.com', now);
}

export { db };
