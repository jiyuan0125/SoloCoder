export const MAX_RETRIES = 3;

export const TABLES = {
  TASKS: 'sync_tasks',
  REPORTS: 'sync_reports',
  RETRY_QUEUE: 'retry_queue',
};

export const STATUS_TRANSITIONS: Record<string, string[]> = {
  pending: ['syncing'],
  syncing: ['success', 'failed'],
  success: ['pending'],
  failed: ['pending'],
};
