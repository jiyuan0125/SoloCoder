export type OperationType = 'CREATE' | 'READ' | 'UPDATE' | 'DELETE';

export interface AuditLog {
  id: number;
  operatorId: string;
  operationType: OperationType;
  resourceType: string;
  resourceId: string;
  timestamp: number;
  beforeSnapshot: Record<string, unknown> | null;
  afterSnapshot: Record<string, unknown> | null;
  sourceIp: string;
  isInternal: boolean;
}

export interface CreateAuditLogRequest {
  operatorId: string;
  operationType: OperationType;
  resourceType: string;
  resourceId?: string;
  beforeSnapshot?: Record<string, unknown> | null;
  afterSnapshot?: Record<string, unknown> | null;
  sourceIp?: string;
  isInternal?: boolean;
}

export interface QueryAuditLogsRequest {
  operatorId?: string;
  operationType?: OperationType;
  resourceType?: string;
  startTime?: number;
  endTime?: number;
  page?: number;
  pageSize?: number;
}

export interface PaginatedResult<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface FieldDiff {
  field: string;
  before: unknown;
  after: unknown;
}

export interface CompareResult {
  log1Id: number;
  log2Id: number;
  resourceType: string;
  resourceId: string;
  differences: FieldDiff[];
}

export interface RetentionConfig {
  retentionDays: number;
}
