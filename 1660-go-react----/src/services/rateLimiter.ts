import db from '../database';
import type { ContentType, WindowType, RateLimitResult } from '../types';

const WINDOW_LIMITS: Record<WindowType, number> = {
  minute: 3,
  hour: 20,
  day: 100,
};

const WINDOW_MS: Record<WindowType, number> = {
  minute: 60 * 1000,
  hour: 60 * 60 * 1000,
  day: 24 * 60 * 60 * 1000,
};

function cleanupExpiredWindows(now: number): void {
  const minTimestamp = now - WINDOW_MS.day;
  db.prepare('DELETE FROM rate_limit_windows WHERE timestamp < ?').run(minTimestamp);
}

function countInWindow(userId: string, contentType: ContentType, windowType: WindowType, now: number): number {
  const windowStart = now - WINDOW_MS[windowType];
  const row = db.prepare(`
    SELECT COUNT(*) as count
    FROM rate_limit_windows
    WHERE user_id = ? AND content_type = ? AND window_type = ? AND timestamp >= ?
  `).get(userId, contentType, windowType, windowStart) as { count: number };
  return row.count;
}

function recordAction(userId: string, contentType: ContentType, now: number): void {
  const insert = db.prepare(`
    INSERT INTO rate_limit_windows (user_id, content_type, window_type, timestamp)
    VALUES (?, ?, ?, ?)
  `);
  insert.run(userId, contentType, 'minute', now);
  insert.run(userId, contentType, 'hour', now);
  insert.run(userId, contentType, 'day', now);
}

export function checkAndRecordRateLimit(userId: string, contentType: ContentType): RateLimitResult {
  const now = Date.now();
  cleanupExpiredWindows(now);

  const windows: WindowType[] = ['minute', 'hour', 'day'];
  for (const window of windows) {
    const count = countInWindow(userId, contentType, window, now);
    if (count >= WINDOW_LIMITS[window]) {
      return { allowed: false, exceededDimension: window };
    }
  }

  recordAction(userId, contentType, now);
  return { allowed: true };
}

export function getRateLimitStatus(userId: string, contentType: ContentType): Record<WindowType, { count: number; limit: number }> {
  const now = Date.now();
  cleanupExpiredWindows(now);
  return {
    minute: { count: countInWindow(userId, contentType, 'minute', now), limit: WINDOW_LIMITS.minute },
    hour: { count: countInWindow(userId, contentType, 'hour', now), limit: WINDOW_LIMITS.hour },
    day: { count: countInWindow(userId, contentType, 'day', now), limit: WINDOW_LIMITS.day },
  };
}