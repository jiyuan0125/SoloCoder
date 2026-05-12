import Database from 'better-sqlite3';
import path from 'path';

const dbPath = process.env.DB_PATH ?? path.resolve(__dirname, '..', 'data.db');
const db = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

const schema = `
CREATE TABLE IF NOT EXISTS warehouses (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL CHECK(type IN ('central', 'area')),
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS material_batches (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  warehouse_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  specification TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  quantity INTEGER NOT NULL CHECK(quantity >= 0),
  inbound_date TEXT NOT NULL,
  expiry_date TEXT,
  supplier TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
);

CREATE TABLE IF NOT EXISTS allocations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  from_warehouse_id INTEGER NOT NULL,
  to_warehouse_id INTEGER NOT NULL,
  material_name TEXT NOT NULL,
  specification TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  requested_quantity INTEGER NOT NULL CHECK(requested_quantity > 0),
  approved_quantity INTEGER CHECK(approved_quantity IS NULL OR approved_quantity >= 0),
  status TEXT NOT NULL CHECK(status IN (
    'pending_approval',
    'approved',
    'in_transit',
    'arrived',
    'received'
  )),
  supplier TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (from_warehouse_id) REFERENCES warehouses(id),
  FOREIGN KEY (to_warehouse_id) REFERENCES warehouses(id)
);

CREATE TABLE IF NOT EXISTS allocation_in_transit (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  allocation_id INTEGER NOT NULL UNIQUE,
  warehouse_id INTEGER NOT NULL,
  material_name TEXT NOT NULL,
  specification TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  quantity INTEGER NOT NULL CHECK(quantity > 0),
  supplier TEXT NOT NULL DEFAULT '',
  inbound_date TEXT NOT NULL,
  expiry_date TEXT,
  FOREIGN KEY (allocation_id) REFERENCES allocations(id),
  FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
);

CREATE TABLE IF NOT EXISTS issue_records (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  warehouse_id INTEGER NOT NULL,
  material_name TEXT NOT NULL,
  specification TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  quantity INTEGER NOT NULL CHECK(quantity > 0),
  issued_at TEXT NOT NULL,
  voided INTEGER NOT NULL DEFAULT 0,
  voided_at TEXT,
  FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
);

CREATE TABLE IF NOT EXISTS issue_lines (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  issue_id INTEGER NOT NULL,
  batch_id INTEGER NOT NULL,
  quantity INTEGER NOT NULL CHECK(quantity > 0),
  FOREIGN KEY (issue_id) REFERENCES issue_records(id),
  FOREIGN KEY (batch_id) REFERENCES material_batches(id)
);

CREATE TABLE IF NOT EXISTS daily_ledger (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ledger_date TEXT NOT NULL,
  warehouse_id INTEGER NOT NULL,
  material_name TEXT NOT NULL,
  specification TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  opening_quantity INTEGER NOT NULL DEFAULT 0,
  inbound_quantity INTEGER NOT NULL DEFAULT 0,
  outbound_quantity INTEGER NOT NULL DEFAULT 0,
  closing_quantity INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(ledger_date, warehouse_id, material_name, specification, category)
);

CREATE TABLE IF NOT EXISTS inventory_differences (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  warehouse_id INTEGER NOT NULL,
  material_name TEXT NOT NULL,
  specification TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  expected_quantity INTEGER NOT NULL,
  actual_quantity INTEGER NOT NULL,
  difference_quantity INTEGER NOT NULL,
  recorded_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
);

CREATE TABLE IF NOT EXISTS check_tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  warehouse_id INTEGER NOT NULL,
  material_name TEXT NOT NULL,
  specification TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  expected_quantity INTEGER NOT NULL,
  actual_quantity INTEGER NOT NULL,
  difference_quantity INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  resolved INTEGER NOT NULL DEFAULT 0,
  FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
);
`;

db.exec(schema);

export default db;
