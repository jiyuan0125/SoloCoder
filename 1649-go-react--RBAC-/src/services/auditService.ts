import db from '../database';
import { v4 as uuidv4 } from 'uuid';

export interface AuditLog {
  id: string;
  operation_type: string;
  entity_type: string;
  entity_id: string | null;
  details: string | null;
  created_at: string;
}

export function logOperation(
  operationType: string,
  entityType: string,
  entityId: string | null,
  details: Record<string, unknown> = {}
): void {
  const id = uuidv4();
  const stmt = db.prepare(`
    INSERT INTO audit_logs (id, operation_type, entity_type, entity_id, details)
    VALUES (?, ?, ?, ?, ?)
  `);
  stmt.run(id, operationType, entityType, entityId, JSON.stringify(details));
}

export function getAuditLogs(
  operationType?: string,
  startTime?: string,
  endTime?: string
): AuditLog[] {
  let query = 'SELECT * FROM audit_logs WHERE 1=1';
  const params: string[] = [];

  if (operationType) {
    query += ' AND operation_type = ?';
    params.push(operationType);
  }

  if (startTime) {
    query += ' AND created_at >= ?';
    params.push(startTime);
  }

  if (endTime) {
    query += ' AND created_at <= ?';
    params.push(endTime);
  }

  query += ' ORDER BY created_at DESC';
  return db.prepare(query).all(...params) as AuditLog[];
}
