import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import type { ErrorStatus, ErrorEventPayload, ErrorGroup, ErrorGroupFilters, TrendPoint } from '../types';

const STATUS_TRANSITIONS: Record<ErrorStatus, ErrorStatus[]> = {
  unhandled: ['acknowledged'],
  acknowledged: ['resolved', 'ignored'],
  resolved: [],
  ignored: [],
};

function extractTopFrame(stackTrace?: string): string | null {
  if (!stackTrace) return null;
  const lines = stackTrace.split('\n').filter(l => l.trim().length > 0);
  for (const line of lines) {
    if (line.includes('at ') || line.match(/^\s*at\s/)) {
      return line.trim();
    }
  }
  return lines[0]?.trim() ?? null;
}

function generateGroupHash(serviceName: string, errorType: string, errorMessage: string, topFrame: string | null): string {
  const base = `${serviceName}:${errorType}:${errorMessage}`;
  const framePart = topFrame ? `:${topFrame}` : '';
  return base + framePart;
}

function serializeContext(context?: Record<string, unknown>): string | null {
  if (!context || Object.keys(context).length === 0) return null;
  return JSON.stringify(context);
}

function extractUserId(context?: Record<string, unknown>): string | null {
  if (!context) return null;
  const userId = context.userId ?? context.user_id ?? context.user;
  if (typeof userId === 'string' || typeof userId === 'number') {
    return String(userId);
  }
  return null;
}

export function processErrorEvent(payload: ErrorEventPayload): ErrorGroup {
  const { serviceName, errorType, errorMessage, stackTrace, occurredAt, context } = payload;

  const missingFields: string[] = [];
  if (!serviceName) missingFields.push('serviceName');
  if (!errorType) missingFields.push('errorType');
  if (!errorMessage) missingFields.push('errorMessage');

  if (missingFields.length > 0) {
    const err = new Error(`缺少必填字段: ${missingFields.join(', ')}`);
    (err as any).statusCode = 400;
    throw err;
  }

  const parsedOccurredAt = occurredAt ? new Date(occurredAt).getTime() : Date.now();
  if (isNaN(parsedOccurredAt)) {
    const err = new Error('无效的 occurredAt 时间格式');
    (err as any).statusCode = 400;
    throw err;
  }

  const topFrame = extractTopFrame(stackTrace);
  const hash = generateGroupHash(serviceName, errorType, errorMessage, topFrame);
  const userId = extractUserId(context);
  const serializedContext = serializeContext(context);

  const tx = db.transaction(() => {
    let existingGroup = db.prepare('SELECT * FROM error_groups WHERE hash = ?').get(hash) as any;

    if (existingGroup) {
      let newStatus = existingGroup.status;
      let newCount = existingGroup.occurrence_count + 1;
      let newAffectedUsers = existingGroup.affected_users;
      let userIdSet = JSON.parse(existingGroup.created_user_id_set || '[]') as string[];

      if (existingGroup.status === 'resolved') {
        newStatus = 'unhandled';
      }

      if (userId && !userIdSet.includes(userId)) {
        userIdSet.push(userId);
        newAffectedUsers = userIdSet.length;
      }

      const twentyFourHoursAgo = Date.now() - 24 * 60 * 60 * 1000;
      const recentCount = (db.prepare(`
        SELECT COUNT(*) as count FROM error_events 
        WHERE group_id = ? AND occurred_at >= ?
      `).get(existingGroup.id, twentyFourHoursAgo) as any).count + 1;

      const isHighFrequency = recentCount > 100 || newAffectedUsers > 50;

      db.prepare(`
        UPDATE error_groups 
        SET status = ?, 
            occurrence_count = ?, 
            last_occurred_at = ?, 
            affected_users = ?, 
            created_user_id_set = ?,
            is_high_frequency = ?
        WHERE id = ?
      `).run(
        newStatus,
        newCount,
        parsedOccurredAt,
        newAffectedUsers,
        JSON.stringify(userIdSet),
        isHighFrequency ? 1 : 0,
        existingGroup.id
      );

      const eventId = uuidv4();
      db.prepare(`
        INSERT INTO error_events (id, group_id, service_name, error_type, error_message, stack_trace, top_frame, occurred_at, context, user_id)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `).run(
        eventId,
        existingGroup.id,
        serviceName,
        errorType,
        errorMessage,
        stackTrace || null,
        topFrame,
        parsedOccurredAt,
        serializedContext,
        userId
      );

      return db.prepare('SELECT * FROM error_groups WHERE id = ?').get(existingGroup.id) as any;
    } else {
      const groupId = uuidv4();
      const eventId = uuidv4();
      const userIdSet = userId ? [userId] : [];
      const affectedUsers = userIdSet.length;

      db.prepare(`
        INSERT INTO error_groups (id, service_name, error_type, error_message, top_frame, status, occurrence_count, last_occurred_at, first_occurred_at, is_high_frequency, affected_users, created_user_id_set, hash)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `).run(
        groupId,
        serviceName,
        errorType,
        errorMessage,
        topFrame,
        'unhandled',
        1,
        parsedOccurredAt,
        parsedOccurredAt,
        0,
        affectedUsers,
        JSON.stringify(userIdSet),
        hash
      );

      db.prepare(`
        INSERT INTO error_events (id, group_id, service_name, error_type, error_message, stack_trace, top_frame, occurred_at, context, user_id)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `).run(
        eventId,
        groupId,
        serviceName,
        errorType,
        errorMessage,
        stackTrace || null,
        topFrame,
        parsedOccurredAt,
        serializedContext,
        userId
      );

      return db.prepare('SELECT * FROM error_groups WHERE id = ?').get(groupId) as any;
    }
  });

  const raw = tx();
  return mapRowToErrorGroup(raw);
}

export function updateErrorGroupStatus(groupId: string, newStatus: ErrorStatus): ErrorGroup {
  const validStatuses: ErrorStatus[] = ['unhandled', 'acknowledged', 'resolved', 'ignored'];
  if (!validStatuses.includes(newStatus)) {
    const err = new Error(`无效的状态值: ${newStatus}`);
    (err as any).statusCode = 400;
    throw err;
  }

  const group = db.prepare('SELECT * FROM error_groups WHERE id = ?').get(groupId) as any;
  if (!group) {
    const err = new Error('错误组不存在');
    (err as any).statusCode = 404;
    throw err;
  }

  const currentStatus = group.status as ErrorStatus;
  const allowedTransitions = STATUS_TRANSITIONS[currentStatus];

  if (!allowedTransitions.includes(newStatus)) {
    const err = new Error(`非法状态转换: ${currentStatus} -> ${newStatus}`);
    (err as any).statusCode = 400;
    throw err;
  }

  db.prepare('UPDATE error_groups SET status = ? WHERE id = ?').run(newStatus, groupId);
  const updated = db.prepare('SELECT * FROM error_groups WHERE id = ?').get(groupId) as any;
  return mapRowToErrorGroup(updated);
}

export function listErrorGroups(filters: ErrorGroupFilters): ErrorGroup[] {
  let sql = 'SELECT * FROM error_groups WHERE 1=1';
  const params: any[] = [];

  if (filters.status) {
    sql += ' AND status = ?';
    params.push(filters.status);
  }

  if (filters.isHighFrequency !== undefined) {
    sql += ' AND is_high_frequency = ?';
    params.push(filters.isHighFrequency ? 1 : 0);
  }

  sql += ' ORDER BY last_occurred_at DESC';

  const rows = db.prepare(sql).all(...params) as any[];
  return rows.map(mapRowToErrorGroup);
}

export function getErrorGroupTrend(groupId: string): TrendPoint[] {
  const group = db.prepare('SELECT id FROM error_groups WHERE id = ?').get(groupId);
  if (!group) {
    const err = new Error('错误组不存在');
    (err as any).statusCode = 404;
    throw err;
  }

  const now = Date.now();
  const twentyFourHoursAgo = now - 24 * 60 * 60 * 1000;

  const rows = db.prepare(`
    SELECT 
      strftime('%Y-%m-%dT%H:00:00Z', datetime(occurred_at / 1000, 'unixepoch')) as hour,
      COUNT(*) as count
    FROM error_events
    WHERE group_id = ? AND occurred_at >= ?
    GROUP BY hour
    ORDER BY hour ASC
  `).all(groupId, twentyFourHoursAgo) as any[];

  return rows.map(row => ({
    hour: row.hour,
    count: row.count,
  }));
}

export function cleanupOldResolvedGroups(): number {
  const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000;

  const result = db.prepare(`
    DELETE FROM error_groups 
    WHERE status = 'resolved' AND last_occurred_at < ?
  `).run(thirtyDaysAgo);

  return result.changes ?? 0;
}

function mapRowToErrorGroup(row: any): ErrorGroup {
  return {
    id: row.id,
    serviceName: row.service_name,
    errorType: row.error_type,
    errorMessage: row.error_message,
    topFrame: row.top_frame,
    status: row.status as ErrorStatus,
    occurrenceCount: row.occurrence_count,
    lastOccurredAt: row.last_occurred_at,
    firstOccurredAt: row.first_occurred_at,
    isHighFrequency: !!row.is_high_frequency,
    affectedUsers: row.affected_users,
    createdUserIdSet: row.created_user_id_set,
  };
}

export function refreshHighFrequencyMarkers(): void {
  const twentyFourHoursAgo = Date.now() - 24 * 60 * 60 * 1000;

  const groups = db.prepare('SELECT * FROM error_groups').all() as any[];

  for (const group of groups) {
    const recentCount = (db.prepare(`
      SELECT COUNT(*) as count FROM error_events 
      WHERE group_id = ? AND occurred_at >= ?
    `).get(group.id, twentyFourHoursAgo) as any).count;

    const affectedUsers = group.affected_users;
    const isHighFrequency = recentCount > 100 || affectedUsers > 50;

    if (isHighFrequency !== !!group.is_high_frequency) {
      db.prepare('UPDATE error_groups SET is_high_frequency = ? WHERE id = ?').run(
        isHighFrequency ? 1 : 0,
        group.id
      );
    }
  }
}
