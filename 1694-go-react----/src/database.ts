import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '..', 'project_fund.db');
const db = new Database(dbPath);

db.exec(`
  CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL,
    location TEXT NOT NULL,
    total_budget INTEGER NOT NULL,
    implementation_period TEXT NOT NULL,
    status TEXT NOT NULL,
    resubmission_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
  );

  CREATE TABLE IF NOT EXISTS project_budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    construction INTEGER NOT NULL,
    equipment INTEGER NOT NULL,
    labor INTEGER NOT NULL,
    other INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS project_expenditures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    category TEXT NOT NULL,
    amount INTEGER NOT NULL,
    description TEXT,
    expenditure_date TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS fund_allocations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    stage TEXT NOT NULL,
    amount INTEGER NOT NULL,
    percentage INTEGER NOT NULL,
    allocation_date TEXT NOT NULL,
    payment_method TEXT NOT NULL,
    receiving_account TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS review_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    review_level TEXT NOT NULL,
    result TEXT NOT NULL,
    comments TEXT,
    reviewer TEXT NOT NULL,
    reviewed_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS inspection_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    stage TEXT NOT NULL,
    result TEXT NOT NULL,
    comments TEXT,
    rectification_deadline TEXT,
    inspector TEXT NOT NULL,
    inspected_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS budget_adjustment_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    from_category TEXT NOT NULL,
    to_category TEXT NOT NULL,
    amount INTEGER NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL,
    approver TEXT,
    approved_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
  );

  CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
  CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
  CREATE INDEX IF NOT EXISTS idx_budgets_project ON project_budgets(project_id);
  CREATE INDEX IF NOT EXISTS idx_allocations_project ON fund_allocations(project_id);
  CREATE INDEX IF NOT EXISTS idx_reviews_project ON review_records(project_id);
  CREATE INDEX IF NOT EXISTS idx_inspections_project ON inspection_records(project_id);
  CREATE INDEX IF NOT EXISTS idx_adjustments_project ON budget_adjustment_requests(project_id);
`);

export default db;
