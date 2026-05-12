import { db } from '../database';
import { generateId, getCurrentTime } from '../utils';

export function createAuditLog(
  operator: string,
  operationType: string,
  content: string,
  beforeSnapshot: object | null,
  afterSnapshot: object | null
): string {
  const id = generateId();
  const now = getCurrentTime();

  const stmt = db.prepare(`
    INSERT INTO audit_logs (
      id, operator, operation_time, operation_type, content,
      before_snapshot, after_snapshot, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `);

  stmt.run(
    id,
    operator,
    now,
    operationType,
    content,
    beforeSnapshot ? JSON.stringify(beforeSnapshot) : null,
    afterSnapshot ? JSON.stringify(afterSnapshot) : null,
    now
  );

  return id;
}

export function getAuditLogs(): any[] {
  return db.prepare(`
    SELECT * FROM audit_logs ORDER BY operation_time DESC
  `).all();
}

export function getAuditLogById(id: string): any {
  return db.prepare(`
    SELECT * FROM audit_logs WHERE id = ?
  `).get(id);
}

export function isAuditLogImmutable(): boolean {
  return true;
}
