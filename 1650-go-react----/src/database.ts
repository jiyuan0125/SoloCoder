import Database from 'better-sqlite3';
import { AuditLog, CreateAuditLogRequest, QueryAuditLogsRequest, PaginatedResult } from './types';
import dayjs from 'dayjs';

const db = new Database('./audit-logs.db');
db.pragma('journal_mode = WAL');
db.pragma('synchronous = NORMAL');
db.pragma('foreign_keys = ON');

db.exec(`
  CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operator_id TEXT NOT NULL,
    operation_type TEXT NOT NULL CHECK(operation_type IN ('CREATE', 'READ', 'UPDATE', 'DELETE')),
    resource_type TEXT NOT NULL,
    resource_id TEXT DEFAULT '',
    timestamp INTEGER NOT NULL,
    before_snapshot TEXT,
    after_snapshot TEXT,
    source_ip TEXT DEFAULT '',
    is_internal INTEGER DEFAULT 0
  );

  CREATE INDEX IF NOT EXISTS idx_operator ON audit_logs(operator_id);
  CREATE INDEX IF NOT EXISTS idx_operation ON audit_logs(operation_type);
  CREATE INDEX IF NOT EXISTS idx_resource ON audit_logs(resource_type, resource_id);
  CREATE INDEX IF NOT EXISTS idx_timestamp ON audit_logs(timestamp DESC);
  CREATE INDEX IF NOT EXISTS idx_dedup ON audit_logs(operator_id, operation_type, resource_type, resource_id, timestamp);

  CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
  );
`);

const DEFAULT_RETENTION_DAYS = 0;

const insertLogStmt = db.prepare(`
  INSERT INTO audit_logs 
    (operator_id, operation_type, resource_type, resource_id, timestamp, before_snapshot, after_snapshot, source_ip, is_internal)
  VALUES 
    (?, ?, ?, ?, ?, ?, ?, ?, ?)
`);

const findDuplicateStmt = db.prepare(`
  SELECT id FROM audit_logs 
  WHERE operator_id = ? 
    AND operation_type = ? 
    AND resource_type = ? 
    AND resource_id = ?
    AND timestamp >= ?
  ORDER BY id DESC
  LIMIT 1
`);

const getLogByIdStmt = db.prepare(`
  SELECT * FROM audit_logs WHERE id = ?
`);

const countLogsStmt = db.prepare(`
  SELECT COUNT(*) as count FROM audit_logs
  WHERE 1=1
    AND (operator_id = ? OR ? IS NULL)
    AND (operation_type = ? OR ? IS NULL)
    AND (resource_type = ? OR ? IS NULL)
    AND (timestamp >= ? OR ? IS NULL)
    AND (timestamp <= ? OR ? IS NULL)
`);

const queryLogsStmt = db.prepare(`
  SELECT * FROM audit_logs
  WHERE 1=1
    AND (operator_id = ? OR ? IS NULL)
    AND (operation_type = ? OR ? IS NULL)
    AND (resource_type = ? OR ? IS NULL)
    AND (timestamp >= ? OR ? IS NULL)
    AND (timestamp <= ? OR ? IS NULL)
  ORDER BY timestamp DESC
  LIMIT ? OFFSET ?
`);

const deleteExpiredLogsStmt = db.prepare(`
  DELETE FROM audit_logs WHERE timestamp < ?
`);

const getConfigStmt = db.prepare('SELECT value FROM config WHERE key = ?');
const setConfigStmt = db.prepare('INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)');

function serializeSnapshot(data: Record<string, unknown> | null | undefined): string | null {
  if (data === undefined || data === null) return null;
  return JSON.stringify(data);
}

function deserializeSnapshot(data: string | null): Record<string, unknown> | null {
  if (data === null) return null;
  try {
    return JSON.parse(data);
  } catch {
    return null;
  }
}

function mapRowToLog(row: any): AuditLog {
  return {
    id: row.id,
    operatorId: row.operator_id,
    operationType: row.operation_type,
    resourceType: row.resource_type,
    resourceId: row.resource_id,
    timestamp: row.timestamp,
    beforeSnapshot: deserializeSnapshot(row.before_snapshot),
    afterSnapshot: deserializeSnapshot(row.after_snapshot),
    sourceIp: row.source_ip,
    isInternal: row.is_internal === 1,
  };
}

function getRetentionDays(): number {
  const row = getConfigStmt.get('retention_days') as any;
  if (!row) return DEFAULT_RETENTION_DAYS;
  const days = parseInt(row.value, 10);
  return isNaN(days) ? DEFAULT_RETENTION_DAYS : days;
}

export function setRetentionDays(days: number): void {
  setConfigStmt.run('retention_days', String(days));
}

export function insertAuditLog(req: CreateAuditLogRequest, timestamp: number, sourceIp: string): number | null {
  const resourceId = req.resourceId || '';
  const deduplicateWindow = 5000;
  const windowStart = timestamp - deduplicateWindow;

  const existing = findDuplicateStmt.get(
    req.operatorId,
    req.operationType,
    req.resourceType,
    resourceId,
    windowStart
  ) as any;

  if (existing) {
    return null;
  }

  const info = insertLogStmt.run(
    req.operatorId,
    req.operationType,
    req.resourceType,
    resourceId,
    timestamp,
    serializeSnapshot(req.beforeSnapshot ?? null),
    serializeSnapshot(req.afterSnapshot ?? null),
    sourceIp,
    req.isInternal ? 1 : 0
  );

  return info.lastInsertRowid as number;
}

export function getAuditLogById(id: number): AuditLog | null {
  const row = getLogByIdStmt.get(id) as any;
  if (!row) return null;
  return mapRowToLog(row);
}

export function queryAuditLogs(req: QueryAuditLogsRequest): PaginatedResult<AuditLog> {
  const page = req.page ?? 1;
  const pageSize = req.pageSize ?? 50;
  const offset = (page - 1) * pageSize;

  const params = [
    req.operatorId ?? null,
    req.operatorId ?? null,
    req.operationType ?? null,
    req.operationType ?? null,
    req.resourceType ?? null,
    req.resourceType ?? null,
    req.startTime ?? null,
    req.startTime ?? null,
    req.endTime ?? null,
    req.endTime ?? null,
  ];

  const countRow = countLogsStmt.get(...params) as any;
  const total = countRow.count;

  const rows = queryLogsStmt.all(...params, pageSize, offset) as any[];
  const data = rows.map(mapRowToLog);

  return { data, total, page, pageSize };
}

export function cleanupExpiredLogs(): { deleted: number } {
  const retentionDays = getRetentionDays();
  if (retentionDays <= 0) return { deleted: 0 };

  const cutoffTimestamp = dayjs().subtract(retentionDays, 'day').valueOf();
  const info = deleteExpiredLogsStmt.run(cutoffTimestamp);
  return { deleted: info.changes };
}

export function getDatabase(): Database.Database {
  return db;
}

export { getRetentionDays };
