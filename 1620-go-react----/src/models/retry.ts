import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { TABLES, MAX_RETRIES } from '../constants';
import { RetryQueueItem, DataRecord } from '../types';

function rowToRetryItem(row: any): RetryQueueItem {
  return {
    id: row.id,
    taskId: row.task_id,
    dataId: row.data_id,
    data: row.data,
    retryCount: row.retry_count,
    maxRetries: row.max_retries,
    lastError: row.last_error,
    status: row.status as 'pending' | 'needs_manual',
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  };
}

export function addToRetryQueue(
  taskId: string,
  record: DataRecord,
  error: string,
  maxRetries: number = MAX_RETRIES
): RetryQueueItem {
  const now = Date.now();
  const item: RetryQueueItem = {
    id: uuidv4(),
    taskId,
    dataId: record.id,
    data: JSON.stringify(record),
    retryCount: 0,
    maxRetries,
    lastError: error,
    status: 'pending',
    createdAt: now,
    updatedAt: now,
  };

  const existing = db
    .prepare(
      `SELECT id FROM ${TABLES.RETRY_QUEUE} 
       WHERE task_id = ? AND data_id = ? AND status = 'pending'`
    )
    .get(taskId, record.id) as { id: string } | undefined;

  if (existing) {
    return incrementRetry(existing.id, error);
  }

  const stmt = db.prepare(`
    INSERT INTO ${TABLES.RETRY_QUEUE} (
      id, task_id, data_id, data, retry_count, max_retries,
      last_error, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);

  stmt.run(
    item.id,
    item.taskId,
    item.dataId,
    item.data,
    item.retryCount,
    item.maxRetries,
    item.lastError,
    item.status,
    item.createdAt,
    item.updatedAt
  );

  return item;
}

export function incrementRetry(id: string, error: string): RetryQueueItem {
  const row = db
    .prepare(`SELECT * FROM ${TABLES.RETRY_QUEUE} WHERE id = ?`)
    .get(id) as any;

  if (!row) throw new Error('Retry item not found');

  const newCount = row.retry_count + 1;
  const now = Date.now();
  const newStatus = newCount >= row.max_retries ? 'needs_manual' : 'pending';

  db.prepare(`
    UPDATE ${TABLES.RETRY_QUEUE}
    SET retry_count = ?, last_error = ?, status = ?, updated_at = ?
    WHERE id = ?
  `).run(newCount, error, newStatus, now, id);

  const updated = db
    .prepare(`SELECT * FROM ${TABLES.RETRY_QUEUE} WHERE id = ?`)
    .get(id);

  return rowToRetryItem(updated);
}

export function getPendingRetryItems(taskId: string): RetryQueueItem[] {
  const rows = db
    .prepare(
      `SELECT * FROM ${TABLES.RETRY_QUEUE} 
       WHERE task_id = ? AND status = 'pending'
       ORDER BY created_at ASC`
    )
    .all(taskId);
  return rows.map(rowToRetryItem);
}

export function getManualInterventionItems(taskId: string): RetryQueueItem[] {
  const rows = db
    .prepare(
      `SELECT * FROM ${TABLES.RETRY_QUEUE} 
       WHERE task_id = ? AND status = 'needs_manual'
       ORDER BY created_at DESC`
    )
    .all(taskId);
  return rows.map(rowToRetryItem);
}

export function removeFromRetryQueue(id: string): void {
  db.prepare(`DELETE FROM ${TABLES.RETRY_QUEUE} WHERE id = ?`).run(id);
}

export function removeFromRetryQueueByDataId(taskId: string, dataId: string): void {
  db
    .prepare(
      `DELETE FROM ${TABLES.RETRY_QUEUE} WHERE task_id = ? AND data_id = ?`
    )
    .run(taskId, dataId);
}

export function markAsResolved(taskId: string, dataId: string): void {
  const now = Date.now();
  db
    .prepare(
      `UPDATE ${TABLES.RETRY_QUEUE} 
       SET status = 'pending', retry_count = 0, updated_at = ?
       WHERE task_id = ? AND data_id = ? AND status = 'needs_manual'`
    )
    .run(now, taskId, dataId);
}
