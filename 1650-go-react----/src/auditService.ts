import {
  CreateAuditLogRequest,
  QueryAuditLogsRequest,
  AuditLog,
  PaginatedResult,
  CompareResult,
  FieldDiff,
  RetentionConfig,
  OperationType,
} from './types';
import {
  insertAuditLog,
  getAuditLogById,
  queryAuditLogs,
  cleanupExpiredLogs,
  setRetentionDays,
  getRetentionDays,
} from './database';

const VALID_OPERATION_TYPES: OperationType[] = ['CREATE', 'READ', 'UPDATE', 'DELETE'];

export interface CreateLogResult {
  success: boolean;
  logId?: number;
  error?: string;
  wasDuplicate?: boolean;
}

export interface CompareLogsResult {
  success: boolean;
  result?: CompareResult;
  error?: string;
  notFound?: boolean;
  resourceMismatch?: boolean;
}

export function validateCreateRequest(req: unknown): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  if (!req || typeof req !== 'object') {
    return { valid: false, errors: ['请求体必须是 JSON 对象'] };
  }

  const data = req as Record<string, unknown>;

  if (!('operatorId' in data) || data.operatorId === undefined || data.operatorId === null || data.operatorId === '') {
    errors.push('缺少必填字段: operatorId');
  } else if (typeof data.operatorId !== 'string') {
    errors.push('operatorId 必须是字符串');
  }

  if (!('operationType' in data) || data.operationType === undefined || data.operationType === null) {
    errors.push('缺少必填字段: operationType');
  } else if (typeof data.operationType !== 'string' || !VALID_OPERATION_TYPES.includes(data.operationType as OperationType)) {
    errors.push(`operationType 必须是以下值之一: ${VALID_OPERATION_TYPES.join(', ')}`);
  }

  if (!('resourceType' in data) || data.resourceType === undefined || data.resourceType === null || data.resourceType === '') {
    errors.push('缺少必填字段: resourceType');
  } else if (typeof data.resourceType !== 'string') {
    errors.push('resourceType 必须是字符串');
  }

  if ('resourceId' in data && data.resourceId !== undefined && data.resourceId !== null) {
    if (typeof data.resourceId !== 'string') {
      errors.push('resourceId 必须是字符串');
    }
  }

  if (!('beforeSnapshot' in data)) {
    errors.push('缺少必填字段: beforeSnapshot');
  }

  if (!('afterSnapshot' in data)) {
    errors.push('缺少必填字段: afterSnapshot');
  }

  return { valid: errors.length === 0, errors };
}

export function createAuditLog(req: CreateAuditLogRequest, sourceIp: string): CreateLogResult {
  const timestamp = Date.now();
  const logId = insertAuditLog(req, timestamp, sourceIp);

  if (logId === null) {
    return { success: true, wasDuplicate: true };
  }

  return { success: true, logId };
}

export function getLogById(id: number): AuditLog | null {
  return getAuditLogById(id);
}

export function listAuditLogs(req: QueryAuditLogsRequest): PaginatedResult<AuditLog> {
  return queryAuditLogs(req);
}

function normalizeSnapshot(snapshot: Record<string, unknown> | null): Record<string, unknown> {
  if (snapshot === null) return {};
  return snapshot;
}

function compareValues(a: unknown, b: unknown): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

function collectFields(obj: Record<string, unknown>, prefix = ''): Set<string> {
  const fields = new Set<string>();
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key;
    fields.add(fullKey);
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      const nested = collectFields(value as Record<string, unknown>, fullKey);
      nested.forEach(f => fields.add(f));
    }
  }
  return fields;
}

function getNestedValue(obj: Record<string, unknown>, path: string): unknown {
  const parts = path.split('.');
  let current: unknown = obj;
  for (const part of parts) {
    if (current === null || current === undefined || typeof current !== 'object') {
      return undefined;
    }
    current = (current as Record<string, unknown>)[part];
  }
  return current;
}

export function compareLogs(log1Id: number, log2Id: number): CompareLogsResult {
  const log1 = getAuditLogById(log1Id);
  const log2 = getAuditLogById(log2Id);

  if (!log1 || !log2) {
    return { success: false, error: '日志不存在', notFound: true };
  }

  if (log1.resourceType !== log2.resourceType || log1.resourceId !== log2.resourceId) {
    return { success: false, error: '资源不匹配', resourceMismatch: true };
  }

  const before1 = normalizeSnapshot(log1.beforeSnapshot);
  const after2 = normalizeSnapshot(log2.afterSnapshot);

  const allFields = new Set<string>();
  collectFields(before1).forEach(f => allFields.add(f));
  collectFields(after2).forEach(f => allFields.add(f));

  const differences: FieldDiff[] = [];
  for (const field of allFields) {
    const v1 = getNestedValue(before1, field);
    const v2 = getNestedValue(after2, field);
    if (!compareValues(v1, v2)) {
      differences.push({ field, before: v1, after: v2 });
    }
  }

  const result: CompareResult = {
    log1Id,
    log2Id,
    resourceType: log1.resourceType,
    resourceId: log1.resourceId,
    differences,
  };

  return { success: true, result };
}

export function getRetentionConfig(): RetentionConfig {
  return { retentionDays: getRetentionDays() };
}

export function updateRetentionConfig(days: number, operatorId: string, sourceIp: string): void {
  const oldDays = getRetentionDays();
  setRetentionDays(days);

  createAuditLog({
    operatorId,
    operationType: 'UPDATE',
    resourceType: 'AUDIT_RETENTION_CONFIG',
    resourceId: 'GLOBAL',
    beforeSnapshot: { retentionDays: oldDays },
    afterSnapshot: { retentionDays: days },
    isInternal: true,
  }, sourceIp);
}

export function runCleanup(): { deleted: number } {
  return cleanupExpiredLogs();
}
