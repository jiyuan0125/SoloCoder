import Database from 'better-sqlite3';
import path from 'path';

const DB_PATH = path.join(__dirname, '..', 'poverty_alleviation.db');

export const db = new Database(DB_PATH);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

export const POVERTY_LEVELS = ['general', 'low_income', 'special_difficult'] as const;
export const POVERTY_CAUSES = [
  '因病', '因残', '因学', '因灾', '缺劳动力', '缺资金',
  '缺技术', '缺土地', '缺水', '交通不便', '自身发展动力不足', '其他'
] as const;
export const MEASURE_TYPES = ['industry', 'employment', 'education', 'health', 'support'] as const;
export const EVALUATION_RESULTS = ['effective', 'partially_effective', 'ineffective'] as const;
export const HOUSEHOLD_STATUSES = ['poverty', 'out_of_poverty', 'return_monitoring'] as const;

export const POVERTY_LINE = 4000;

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS households (
      id TEXT PRIMARY KEY,
      household_id TEXT UNIQUE NOT NULL,
      county TEXT NOT NULL,
      township TEXT NOT NULL,
      village TEXT NOT NULL,
      head_of_household TEXT NOT NULL,
      contact_phone TEXT,
      family_size INTEGER NOT NULL,
      family_members TEXT NOT NULL,
      poverty_causes TEXT NOT NULL,
      main_poverty_cause TEXT NOT NULL,
      poverty_level TEXT NOT NULL,
      assets TEXT NOT NULL,
      labor_force TEXT NOT NULL,
      yearly_income_per_capita REAL NOT NULL DEFAULT 0,
      two_worries_three_guarantees TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT 'poverty',
      out_of_poverty_date TEXT,
      monitoring_end_date TEXT,
      helper_name TEXT NOT NULL,
      helper_unit TEXT NOT NULL,
      helper_phone TEXT,
      archive_year INTEGER NOT NULL,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS assistance_plans (
      id TEXT PRIMARY KEY,
      household_id TEXT NOT NULL UNIQUE,
      measures TEXT NOT NULL,
      adjustment_reason TEXT,
      approved_by_township INTEGER DEFAULT 0,
      approval_date TEXT,
      status TEXT NOT NULL DEFAULT 'draft',
      created_by TEXT NOT NULL,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL,
      FOREIGN KEY (household_id) REFERENCES households(household_id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS household_histories (
      id TEXT PRIMARY KEY,
      household_id TEXT NOT NULL,
      event_type TEXT NOT NULL,
      previous_status TEXT,
      new_status TEXT,
      description TEXT NOT NULL,
      operator TEXT NOT NULL,
      operation_date TEXT NOT NULL,
      FOREIGN KEY (household_id) REFERENCES households(household_id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS plan_histories (
      id TEXT PRIMARY KEY,
      plan_id TEXT NOT NULL,
      household_id TEXT NOT NULL,
      event_type TEXT NOT NULL,
      previous_status TEXT,
      new_status TEXT,
      reason TEXT,
      operator TEXT NOT NULL,
      operation_date TEXT NOT NULL,
      FOREIGN KEY (plan_id) REFERENCES assistance_plans(id) ON DELETE CASCADE,
      FOREIGN KEY (household_id) REFERENCES households(household_id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_households_township ON households(township);
    CREATE INDEX IF NOT EXISTS idx_households_village ON households(village);
    CREATE INDEX IF NOT EXISTS idx_households_status ON households(status);
    CREATE INDEX IF NOT EXISTS idx_histories_household ON household_histories(household_id);
    CREATE INDEX IF NOT EXISTS idx_plan_histories_household ON plan_histories(household_id);
  `);
}
