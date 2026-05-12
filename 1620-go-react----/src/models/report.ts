import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { TABLES } from '../constants';
import { SyncReport, SyncResult } from '../types';

function rowToReport(row: any): SyncReport {
  return {
    id: row.id,
    taskId: row.task_id,
    createdAt: row.created_at,
    insertCount: row.insert_count,
    updateCount: row.update_count,
    deleteCount: row.delete_count,
    conflictCount: row.conflict_count,
    manualInterventionItems: row.manual_intervention_items
      ? JSON.parse(row.manual_intervention_items)
      : [],
    details: row.details || '',
  };
}

export function createReport(taskId: string, result: SyncResult): SyncReport {
  const now = Date.now();
  const report: SyncReport = {
    id: uuidv4(),
    taskId,
    createdAt: now,
    insertCount: result.insertCount,
    updateCount: result.updateCount,
    deleteCount: result.deleteCount,
    conflictCount: result.conflictCount,
    manualInterventionItems: result.manualInterventionItems,
    details: JSON.stringify({
      conflicts: result.conflicts,
      manualInterventionItems: result.manualInterventionItems,
    }),
  };

  const stmt = db.prepare(`
    INSERT INTO ${TABLES.REPORTS} (
      id, task_id, created_at, insert_count, update_count,
      delete_count, conflict_count, manual_intervention_items, details
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);

  stmt.run(
    report.id,
    report.taskId,
    report.createdAt,
    report.insertCount,
    report.updateCount,
    report.deleteCount,
    report.conflictCount,
    JSON.stringify(report.manualInterventionItems),
    report.details
  );

  return report;
}

export function getReportsByTaskId(taskId: string): SyncReport[] {
  const rows = db
    .prepare(
      `SELECT * FROM ${TABLES.REPORTS} WHERE task_id = ? ORDER BY created_at DESC`
    )
    .all(taskId);
  return rows.map(rowToReport);
}

export function getLatestReport(taskId: string): SyncReport | undefined {
  const row = db
    .prepare(
      `SELECT * FROM ${TABLES.REPORTS} WHERE task_id = ? ORDER BY created_at DESC LIMIT 1`
    )
    .get(taskId);
  return row ? rowToReport(row) : undefined;
}

export function getReportById(id: string): SyncReport | undefined {
  const row = db.prepare(`SELECT * FROM ${TABLES.REPORTS} WHERE id = ?`).get(id);
  return row ? rowToReport(row) : undefined;
}
