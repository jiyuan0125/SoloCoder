import Database from 'better-sqlite3';
import { Change, Approval, Fault, ChangeStatus, ChangeType } from './types';

const db = new Database('./changes.db');

db.exec(`
  CREATE TABLE IF NOT EXISTS changes (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('release', 'config', 'scale', 'migrate')),
    affected_scope TEXT NOT NULL,
    plan_description TEXT NOT NULL,
    rollback_plan TEXT NOT NULL,
    scheduled_at TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'approved', 'executing', 'completed', 'rolled_back')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    executed_at TEXT,
    completed_at TEXT,
    rolled_back_at TEXT,
    rollback_reason TEXT
  );

  CREATE TABLE IF NOT EXISTS approvals (
    id TEXT PRIMARY KEY,
    change_id TEXT NOT NULL,
    approver TEXT NOT NULL,
    approved INTEGER NOT NULL,
    comment TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(change_id) REFERENCES changes(id)
  );

  CREATE TABLE IF NOT EXISTS faults (
    id TEXT PRIMARY KEY,
    change_id TEXT NOT NULL,
    description TEXT NOT NULL,
    is_false_positive INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    FOREIGN KEY(change_id) REFERENCES changes(id)
  );
`);

export const dbInstance = db;

export const insertChange = (change: Change) => {
  const stmt = db.prepare(`
    INSERT INTO changes (
      id, title, type, affected_scope, plan_description, rollback_plan,
      scheduled_at, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  stmt.run(
    change.id,
    change.title,
    change.type,
    change.affectedScope,
    change.planDescription,
    change.rollbackPlan,
    change.scheduledAt,
    change.status,
    change.createdAt,
    change.updatedAt
  );
};

export const updateChange = (id: string, updates: Partial<Change>) => {
  const fields: string[] = [];
  const values: any[] = [];

  const mapping: Record<string, string> = {
    title: 'title',
    type: 'type',
    affectedScope: 'affected_scope',
    planDescription: 'plan_description',
    rollbackPlan: 'rollback_plan',
    scheduledAt: 'scheduled_at',
    status: 'status',
    updatedAt: 'updated_at',
    executedAt: 'executed_at',
    completedAt: 'completed_at',
    rolledBackAt: 'rolled_back_at',
    rollbackReason: 'rollback_reason'
  };

  for (const [key, value] of Object.entries(updates)) {
    const dbField = mapping[key];
    if (dbField && value !== undefined) {
      fields.push(`${dbField} = ?`);
      values.push(value);
    }
  }

  if (fields.length === 0) return;

  values.push(id);
  const stmt = db.prepare(`UPDATE changes SET ${fields.join(', ')} WHERE id = ?`);
  stmt.run(...values);
};

export const getChangeById = (id: string): Change | undefined => {
  const stmt = db.prepare(`
    SELECT 
      id, title, type, affected_scope as affectedScope, 
      plan_description as planDescription, rollback_plan as rollbackPlan,
      scheduled_at as scheduledAt, status, created_at as createdAt,
      updated_at as updatedAt, executed_at as executedAt,
      completed_at as completedAt, rolled_back_at as rolledBackAt,
      rollback_reason as rollbackReason
    FROM changes WHERE id = ?
  `);
  const row = stmt.get(id) as any;
  if (!row) return undefined;
  return row as Change;
};

export const getAllChanges = (): Change[] => {
  const stmt = db.prepare(`
    SELECT 
      id, title, type, affected_scope as affectedScope, 
      plan_description as planDescription, rollback_plan as rollbackPlan,
      scheduled_at as scheduledAt, status, created_at as createdAt,
      updated_at as updatedAt, executed_at as executedAt,
      completed_at as completedAt, rolled_back_at as rolledBackAt,
      rollback_reason as rollbackReason
    FROM changes ORDER BY created_at DESC
  `);
  return stmt.all() as Change[];
};

export const insertApproval = (approval: Approval) => {
  const stmt = db.prepare(`
    INSERT INTO approvals (id, change_id, approver, approved, comment, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
  `);
  stmt.run(
    approval.id,
    approval.changeId,
    approval.approver,
    approval.approved ? 1 : 0,
    approval.comment || null,
    approval.createdAt
  );
};

export const getApprovalsByChangeId = (changeId: string): Approval[] => {
  const stmt = db.prepare(`
    SELECT id, change_id as changeId, approver, approved, comment, created_at as createdAt
    FROM approvals WHERE change_id = ? ORDER BY created_at DESC
  `);
  return stmt.all(changeId).map((row: any) => ({
    ...row,
    approved: row.approved === 1
  })) as Approval[];
};

export const getExecutingChange = (now: string): Change | undefined => {
  const stmt = db.prepare(`
    SELECT 
      id, title, type, affected_scope as affectedScope, 
      plan_description as planDescription, rollback_plan as rollbackPlan,
      scheduled_at as scheduledAt, status, created_at as createdAt,
      updated_at as updatedAt, executed_at as executedAt,
      completed_at as completedAt, rolled_back_at as rolledBackAt,
      rollback_reason as rollbackReason
    FROM changes 
    WHERE status = 'executing' 
      AND datetime(executed_at, '+2 hours') > ?
    ORDER BY executed_at DESC
    LIMIT 1
  `);
  const row = stmt.get(now) as any;
  if (!row) return undefined;
  return row as Change;
};

export const getFaultsByChangeId = (changeId: string): Fault[] => {
  const stmt = db.prepare(`
    SELECT id, change_id as changeId, description, is_false_positive as isFalsePositive, created_at as createdAt
    FROM faults WHERE change_id = ?
  `);
  return stmt.all(changeId).map((row: any) => ({
    ...row,
    isFalsePositive: row.isFalsePositive === 1
  })) as Fault[];
};
