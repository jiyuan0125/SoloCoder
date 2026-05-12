import { v4 as uuidv4 } from 'uuid';
import db from '../db';
import { BroadcastTask, TaskStatus, UserTier, SendResult } from '../types';

function rowToTask(row: any): BroadcastTask {
  return {
    id: row.id,
    name: row.name,
    content: row.content,
    scheduledAt: row.scheduled_at,
    status: row.status as TaskStatus,
    createdAt: row.created_at,
    tierFilter: row.tier_filter || undefined,
    tagFilterMode: row.tag_filter_mode || undefined
  };
}

export interface CreateTaskInput {
  name: string;
  content: string;
  scheduledAt: number;
  communityIds: string[];
  tierFilter?: UserTier;
  tagFilterMode?: 'AND' | 'OR';
  tagFilterIds?: string[];
}

export interface TaskDetail {
  task: BroadcastTask;
  communities: {
    communityId: string;
    communityName?: string;
    sendResult?: SendResult;
    failureReason?: string;
    sentAt?: number;
  }[];
  tagFilterIds: string[];
}

export function createTask(input: CreateTaskInput): BroadcastTask {
  if (!input.communityIds || input.communityIds.length === 0) {
    const err: any = new Error('At least one community is required');
    err.status = 400;
    throw err;
  }

  const tx = db.transaction(() => {
    const id = uuidv4();
    const now = Date.now();

    db.prepare(`
      INSERT INTO broadcast_tasks (id, name, content, scheduled_at, status, created_at, tier_filter, tag_filter_mode)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `).run(
      id,
      input.name,
      input.content,
      input.scheduledAt,
      TaskStatus.PENDING,
      now,
      input.tierFilter || null,
      input.tagFilterMode || null
    );

    for (const communityId of input.communityIds) {
      try {
        db.prepare(`
          INSERT INTO task_target_communities (id, task_id, community_id)
          VALUES (?, ?, ?)
        `).run(uuidv4(), id, communityId);
      } catch (e: any) {
        if (e.code === 'SQLITE_CONSTRAINT_UNIQUE') {
          const err: any = new Error(`Cannot add duplicate community ${communityId} to the same task`);
          err.status = 409;
          throw err;
        }
        throw e;
      }
    }

    if (input.tagFilterIds && input.tagFilterIds.length > 0) {
      for (const tagId of input.tagFilterIds) {
        db.prepare(`
          INSERT INTO task_tag_filters (id, task_id, tag_id)
          VALUES (?, ?, ?)
        `).run(uuidv4(), id, tagId);
      }
    }

    return {
      id,
      name: input.name,
      content: input.content,
      scheduledAt: input.scheduledAt,
      status: TaskStatus.PENDING,
      createdAt: now,
      tierFilter: input.tierFilter,
      tagFilterMode: input.tagFilterMode
    };
  });

  return tx();
}

export function getAllTasks(): BroadcastTask[] {
  const rows = db.prepare('SELECT * FROM broadcast_tasks ORDER BY created_at DESC').all();
  return rows.map(rowToTask);
}

export function getTaskDetail(taskId: string): TaskDetail | undefined {
  const taskRow = db.prepare('SELECT * FROM broadcast_tasks WHERE id = ?').get(taskId) as any;
  if (!taskRow) {
    return undefined;
  }

  const communities = db.prepare(`
    SELECT ttc.community_id, c.name as community_name, ttc.send_result, ttc.failure_reason, ttc.sent_at
    FROM task_target_communities ttc
    LEFT JOIN communities c ON c.id = ttc.community_id
    WHERE ttc.task_id = ?
  `).all(taskId) as any[];

  const tagFilters = db.prepare(
    'SELECT tag_id FROM task_tag_filters WHERE task_id = ?'
  ).all(taskId) as any[];

  return {
    task: rowToTask(taskRow),
    communities: communities.map(c => ({
      communityId: c.community_id,
      communityName: c.community_name,
      sendResult: c.send_result,
      failureReason: c.failure_reason,
      sentAt: c.sent_at
    })),
    tagFilterIds: tagFilters.map(t => t.tag_id)
  };
}

export function cancelTask(taskId: string): void {
  const task = db.prepare('SELECT * FROM broadcast_tasks WHERE id = ?').get(taskId) as any;
  if (!task) {
    const err: any = new Error('Task not found');
    err.status = 404;
    throw err;
  }

  if (task.status === TaskStatus.RUNNING) {
    const err: any = new Error('任务执行中');
    err.status = 409;
    throw err;
  }

  if (task.status === TaskStatus.COMPLETED || task.status === TaskStatus.CANCELLED) {
    const err: any = new Error('Task is already completed or cancelled');
    err.status = 400;
    throw err;
  }

  db.prepare(
    "UPDATE broadcast_tasks SET status = ? WHERE id = ?"
  ).run(TaskStatus.CANCELLED, taskId);
}

export function getPendingTasks(): BroadcastTask[] {
  const now = Date.now();
  const rows = db.prepare(`
    SELECT * FROM broadcast_tasks
    WHERE status = ? AND scheduled_at <= ?
    ORDER BY scheduled_at ASC
  `).all(TaskStatus.PENDING, now) as any[];
  return rows.map(rowToTask);
}

export function markTaskRunning(taskId: string): boolean {
  const result = db.prepare(`
    UPDATE broadcast_tasks SET status = ?
    WHERE id = ? AND status = ?
  `).run(TaskStatus.RUNNING, taskId, TaskStatus.PENDING);

  return result.changes > 0;
}

export function markTaskCompleted(taskId: string): void {
  db.prepare(
    "UPDATE broadcast_tasks SET status = ? WHERE id = ?"
  ).run(TaskStatus.COMPLETED, taskId);
}

export function updateCommunitySendResult(
  taskId: string,
  communityId: string,
  result: SendResult,
  failureReason?: string
): void {
  db.prepare(`
    UPDATE task_target_communities
    SET send_result = ?, failure_reason = ?, sent_at = ?
    WHERE task_id = ? AND community_id = ?
  `).run(result, failureReason || null, Date.now(), taskId, communityId);
}

export function getTaskTargetCommunities(taskId: string): string[] {
  const rows = db.prepare(
    'SELECT community_id FROM task_target_communities WHERE task_id = ?'
  ).all(taskId) as any[];
  return rows.map(r => r.community_id);
}
