export enum SyncMode {
  FULL = 'full',
  INCREMENTAL = 'incremental',
}

export enum TaskStatus {
  PENDING = 'pending',
  SYNCING = 'syncing',
  SUCCESS = 'success',
  FAILED = 'failed',
}

export enum ConflictStrategy {
  SOURCE_PRIORITY = 'source_priority',
  TARGET_PRIORITY = 'target_priority',
  MANUAL_RESOLVE = 'manual_resolve',
}

export enum DataOperation {
  INSERT = 'insert',
  UPDATE = 'update',
  DELETE = 'delete',
  CONFLICT = 'conflict',
}

export interface SyncTask {
  id: string;
  name: string;
  mode: SyncMode;
  sourceSystem: string;
  targetSystem: string;
  watermark: number;
  conflictStrategy: ConflictStrategy;
  status: TaskStatus;
  createdAt: number;
  updatedAt: number;
}

export interface SyncReport {
  id: string;
  taskId: string;
  createdAt: number;
  insertCount: number;
  updateCount: number;
  deleteCount: number;
  conflictCount: number;
  manualInterventionItems: string[];
  details: string;
}

export interface RetryQueueItem {
  id: string;
  taskId: string;
  dataId: string;
  data: string;
  retryCount: number;
  maxRetries: number;
  lastError: string;
  status: 'pending' | 'needs_manual';
  createdAt: number;
  updatedAt: number;
}

export interface DataRecord {
  id: string;
  updatedAt: number;
  [key: string]: any;
}

export interface SyncResult {
  insertCount: number;
  updateCount: number;
  deleteCount: number;
  conflictCount: number;
  manualInterventionItems: string[];
  conflicts: ConflictItem[];
}

export interface ConflictItem {
  dataId: string;
  sourceUpdatedAt: number;
  targetUpdatedAt: number;
  resolved: boolean;
  resolvedBy?: ConflictStrategy;
}

export interface CreateTaskRequest {
  name: string;
  mode: SyncMode;
  sourceSystem: string;
  targetSystem: string;
  conflictStrategy: ConflictStrategy;
}

export interface UpdateTaskRequest {
  name?: string;
  conflictStrategy?: ConflictStrategy;
}

export interface StartTaskResponse {
  success: boolean;
  message?: string;
  report?: SyncReport;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}
