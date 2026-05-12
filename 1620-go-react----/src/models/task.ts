import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { TABLES, STATUS_TRANSITIONS } from '../constants';
import {
  SyncTask,
  TaskStatus,
  SyncMode,
  ConflictStrategy,
  CreateTaskRequest,
  UpdateTaskRequest,
} from '../types';

function rowToTask(row: any): SyncTask {
  return {
    id: row.id,
    name: row.name,
    mode: row.mode as SyncMode,
    sourceSystem: row.source_system,
    targetSystem: row.target_system,
    watermark: row.watermark,
    conflictStrategy: row.conflict_strategy as ConflictStrategy,
    status: row.status as TaskStatus,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  };
}

export function getAllTasks(): SyncTask[] {
  const rows = db.prepare(`SELECT * FROM ${TABLES.TASKS} ORDER BY created_at DESC`).all();
  return rows.map(rowToTask);
}

export function getTaskById(id: string): SyncTask | undefined {
  const row = db.prepare(`SELECT * FROM ${TABLES.TASKS} WHERE id = ?`).get(id);
  return row ? rowToTask(row) : undefined;
}

export function createTask(req: CreateTaskRequest): SyncTask {
  const now = Date.now();
  const task: SyncTask = {
    id: uuidv4(),
    name: req.name,
    mode: req.mode,
    sourceSystem: req.sourceSystem,
    targetSystem: req.targetSystem,
    watermark: 0,
    conflictStrategy: req.conflictStrategy,
    status: TaskStatus.PENDING,
    createdAt: now,
    updatedAt: now,
  };

  const stmt = db.prepare(`
    INSERT INTO ${TABLES.TASKS} (
      id, name, mode, source_system, target_system, watermark,
      conflict_strategy, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);

  stmt.run(
    task.id,
    task.name,
    task.mode,
    task.sourceSystem,
    task.targetSystem,
    task.watermark,
    task.conflictStrategy,
    task.status,
    task.createdAt,
    task.updatedAt
  );

  return task;
}

export function updateTask(id: string, req: UpdateTaskRequest): SyncTask | undefined {
  const task = getTaskById(id);
  if (!task) return undefined;

  const now = Date.now();
  const updates: string[] = [];
  const values: any[] = [];

  if (req.name !== undefined) {
    updates.push('name = ?');
    values.push(req.name);
  }
  if (req.conflictStrategy !== undefined) {
    updates.push('conflict_strategy = ?');
    values.push(req.conflictStrategy);
  }

  if (updates.length === 0) return task;

  updates.push('updated_at = ?');
  values.push(now);
  values.push(id);

  const stmt = db.prepare(`
    UPDATE ${TABLES.TASKS}
    SET ${updates.join(', ')}
    WHERE id = ?
  `);

  stmt.run(...values);
  return getTaskById(id);
}

export function deleteTask(id: string): boolean {
  const result = db.prepare(`DELETE FROM ${TABLES.TASKS} WHERE id = ?`).run(id);
  return result.changes > 0;
}

export function canTransitionStatus(current: TaskStatus, next: TaskStatus): boolean {
  const allowed = STATUS_TRANSITIONS[current] || [];
  return allowed.includes(next);
}

export function updateTaskStatus(
  id: string,
  newStatus: TaskStatus
): { success: boolean; reason?: string; task?: SyncTask } {
  const task = getTaskById(id);
  if (!task) {
    return { success: false, reason: 'Task not found' };
  }

  if (!canTransitionStatus(task.status, newStatus)) {
    return {
      success: false,
      reason: `Invalid status transition: ${task.status} -> ${newStatus}. Only allowed: ${STATUS_TRANSITIONS[task.status]?.join(', ') || 'none'}`,
    };
  }

  const now = Date.now();
  const stmt = db.prepare(`
    UPDATE ${TABLES.TASKS}
    SET status = ?, updated_at = ?
    WHERE id = ?
  `);

  stmt.run(newStatus, now, id);
  const updated = getTaskById(id);

  return { success: true, task: updated };
}

export function updateTaskWatermark(id: string, watermark: number): void {
  const now = Date.now();
  db.prepare(`
    UPDATE ${TABLES.TASKS}
    SET watermark = ?, updated_at = ?
    WHERE id = ?
  `).run(watermark, now, id);
}
