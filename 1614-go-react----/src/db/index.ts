import Database, { Database as DatabaseType } from 'better-sqlite3';
import { Plan, ErrorRecord, PlanStatus } from '../types';

const db: DatabaseType = new Database('./grayscale.db');

export { db };

db.exec(`
  CREATE TABLE IF NOT EXISTS plans (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    app_id TEXT NOT NULL,
    old_version TEXT NOT NULL,
    new_version TEXT NOT NULL,
    strategy TEXT NOT NULL,
    status TEXT NOT NULL,
    error_rate_config TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS error_records (
    id TEXT PRIMARY KEY,
    plan_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    is_error INTEGER NOT NULL,
    FOREIGN KEY (plan_id) REFERENCES plans(id)
  );

  CREATE INDEX IF NOT EXISTS idx_error_records_plan_time ON error_records(plan_id, timestamp);
  CREATE INDEX IF NOT EXISTS idx_plans_app_status ON plans(app_id, status);
`);

export const planRepository = {
  create: (plan: Plan) => {
    const stmt = db.prepare(`
      INSERT INTO plans (id, name, app_id, old_version, new_version, strategy, status, error_rate_config, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(
      plan.id,
      plan.name,
      plan.appId,
      plan.oldVersion,
      plan.newVersion,
      JSON.stringify(plan.strategy),
      plan.status,
      JSON.stringify(plan.errorRateConfig),
      plan.createdAt,
      plan.updatedAt
    );
  },

  findById: (id: string): Plan | null => {
    const stmt = db.prepare('SELECT * FROM plans WHERE id = ?');
    const row = stmt.get(id) as any;
    if (!row) return null;
    return {
      id: row.id,
      name: row.name,
      appId: row.app_id,
      oldVersion: row.old_version,
      newVersion: row.new_version,
      strategy: JSON.parse(row.strategy),
      status: row.status as PlanStatus,
      errorRateConfig: JSON.parse(row.error_rate_config),
      createdAt: row.created_at,
      updatedAt: row.updated_at
    };
  },

  findByAppIdAndStatus: (appId: string, status: PlanStatus): Plan | null => {
    const stmt = db.prepare('SELECT * FROM plans WHERE app_id = ? AND status = ?');
    const row = stmt.get(appId, status) as any;
    if (!row) return null;
    return {
      id: row.id,
      name: row.name,
      appId: row.app_id,
      oldVersion: row.old_version,
      newVersion: row.new_version,
      strategy: JSON.parse(row.strategy),
      status: row.status as PlanStatus,
      errorRateConfig: JSON.parse(row.error_rate_config),
      createdAt: row.created_at,
      updatedAt: row.updated_at
    };
  },

  findAll: (): Plan[] => {
    const stmt = db.prepare('SELECT * FROM plans ORDER BY created_at DESC');
    const rows = stmt.all() as any[];
    return rows.map(row => ({
      id: row.id,
      name: row.name,
      appId: row.app_id,
      oldVersion: row.old_version,
      newVersion: row.new_version,
      strategy: JSON.parse(row.strategy),
      status: row.status as PlanStatus,
      errorRateConfig: JSON.parse(row.error_rate_config),
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }));
  },

  update: (plan: Plan) => {
    const stmt = db.prepare(`
      UPDATE plans 
      SET name = ?, strategy = ?, status = ?, error_rate_config = ?, updated_at = ?
      WHERE id = ?
    `);
    stmt.run(
      plan.name,
      JSON.stringify(plan.strategy),
      plan.status,
      JSON.stringify(plan.errorRateConfig),
      plan.updatedAt,
      plan.id
    );
  },

  delete: (id: string) => {
    const stmt = db.prepare('DELETE FROM plans WHERE id = ?');
    stmt.run(id);
  }
};

export const errorRepository = {
  create: (record: ErrorRecord) => {
    const stmt = db.prepare(`
      INSERT INTO error_records (id, plan_id, timestamp, is_error)
      VALUES (?, ?, ?, ?)
    `);
    stmt.run(record.id, record.planId, record.timestamp, record.isError ? 1 : 0);
  },

  getRecordsInWindow: (planId: string, startTime: number, endTime: number): ErrorRecord[] => {
    const stmt = db.prepare(`
      SELECT * FROM error_records 
      WHERE plan_id = ? AND timestamp >= ? AND timestamp <= ?
      ORDER BY timestamp ASC
    `);
    const rows = stmt.all(planId, startTime, endTime) as any[];
    return rows.map(row => ({
      id: row.id,
      planId: row.plan_id,
      timestamp: row.timestamp,
      isError: row.is_error === 1
    }));
  },

  calculateErrorRate: (planId: string, startTime: number, endTime: number): { total: number; errors: number; rate: number } => {
    const stmt = db.prepare(`
      SELECT 
        COUNT(*) as total,
        SUM(CASE WHEN is_error = 1 THEN 1 ELSE 0 END) as errors
      FROM error_records 
      WHERE plan_id = ? AND timestamp >= ? AND timestamp <= ?
    `);
    const row = stmt.get(planId, startTime, endTime) as any;
    const total = row.total || 0;
    const errors = row.errors || 0;
    const rate = total === 0 ? 0 : (errors / total) * 100;
    return { total, errors, rate };
  }
};

export default db;
