import db from '../database';
import type { User, BlacklistEntry, MarkBlacklistResult } from '../types';

export function ensureUserExists(userId: string): void {
  db.prepare('INSERT OR IGNORE INTO users (id) VALUES (?)').run(userId);
}

export function getUser(userId: string): User | null {
  return db.prepare('SELECT * FROM users WHERE id = ?').get(userId) as User | null;
}

export function getBlacklistEntry(userId: string): BlacklistEntry | null {
  const entry = db.prepare('SELECT * FROM blacklist WHERE user_id = ?').get(userId) as BlacklistEntry | null;
  if (!entry) return null;
  if (entry.expires_at !== null && Date.now() > entry.expires_at) {
    db.prepare('DELETE FROM blacklist WHERE user_id = ?').run(userId);
    return null;
  }
  return entry;
}

export function isUserBlacklisted(userId: string): boolean {
  return getBlacklistEntry(userId) !== null;
}

export function isUserWhitelisted(userId: string): boolean {
  const user = getUser(userId);
  return user !== null && user.is_whitelist === 1;
}

export function setWhitelist(userId: string, isWhitelist: boolean): boolean {
  ensureUserExists(userId);
  const result = db.prepare('UPDATE users SET is_whitelist = ? WHERE id = ?').run(isWhitelist ? 1 : 0, userId);
  return result.changes > 0;
}

export function markAsBlacklist(
  userId: string,
  durationMs: number | null,
  reason: string | null = null
): MarkBlacklistResult {
  ensureUserExists(userId);
  const now = Date.now();
  const expiresAt = durationMs !== null ? now + durationMs : null;

  db.prepare(`
    INSERT INTO blacklist (user_id, added_at, expires_at, reason)
    VALUES (?, ?, ?, ?)
    ON CONFLICT(user_id) DO UPDATE SET
      added_at = excluded.added_at,
      expires_at = excluded.expires_at,
      reason = excluded.reason
  `).run(userId, now, expiresAt, reason);

  const pendingPosts = db.prepare(`
    SELECT id FROM posts WHERE user_id = ? AND status IN ('pending_review', 'suspicious')
  `).all(userId) as Array<{ id: string }>;

  let successCount = 0;
  let failCount = 0;

  for (const post of pendingPosts) {
    try {
      const result = db.prepare("UPDATE posts SET status = 'publisher_blocked' WHERE id = ?").run(post.id);
      if (result.changes > 0) {
        successCount++;
      } else {
        failCount++;
      }
    } catch {
      failCount++;
    }
  }

  const allSuccess = failCount === 0;
  return {
    success: allSuccess,
    totalCleaned: pendingPosts.length,
    successCleaned: successCount,
    failedCleaned: failCount,
  };
}

export function removeFromBlacklist(userId: string): boolean {
  const result = db.prepare('DELETE FROM blacklist WHERE user_id = ?').run(userId);
  return result.changes > 0;
}