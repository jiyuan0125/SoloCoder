import { v4 as uuidv4 } from 'uuid';
import db from '../db';
import { Tag, User, UserTier } from '../types';

function rowToTag(row: any): Tag {
  return {
    id: row.id,
    name: row.name,
    category: row.category,
    createdAt: row.created_at
  };
}

function rowToUser(row: any): User {
  return {
    id: row.id,
    name: row.name,
    lastActiveAt: row.last_active_at,
    createdAt: row.created_at
  };
}

export function createTag(data: { name: string; category: string }): Tag {
  const existing = db.prepare(
    'SELECT id FROM tags WHERE name = ? AND category = ?'
  ).get(data.name, data.category);

  if (existing) {
    return rowToTag(existing);
  }

  const id = uuidv4();
  const now = Date.now();

  db.prepare(`
    INSERT INTO tags (id, name, category, created_at)
    VALUES (?, ?, ?, ?)
  `).run(id, data.name, data.category, now);

  return {
    id,
    name: data.name,
    category: data.category,
    createdAt: now
  };
}

export function getAllTags(): Tag[] {
  const rows = db.prepare('SELECT * FROM tags ORDER BY created_at DESC').all();
  return rows.map(rowToTag);
}

export function assignTagToUser(userId: string, tagId: string): void {
  const tagExists = db.prepare('SELECT id FROM tags WHERE id = ?').get(tagId);
  if (!tagExists) {
    const err: any = new Error('Tag not found');
    err.status = 404;
    throw err;
  }

  const userExists = db.prepare('SELECT id FROM users WHERE id = ?').get(userId);
  if (!userExists) {
    const now = Date.now();
    db.prepare(
      'INSERT INTO users (id, name, last_active_at, created_at) VALUES (?, ?, ?, ?)'
    ).run(userId, `user_${userId.slice(0, 8)}`, now, now);
  }

  const existing = db.prepare(
    'SELECT id FROM user_tags WHERE user_id = ? AND tag_id = ?'
  ).get(userId, tagId);

  if (existing) {
    return;
  }

  db.prepare(`
    INSERT INTO user_tags (id, user_id, tag_id, created_at)
    VALUES (?, ?, ?, ?)
  `).run(uuidv4(), userId, tagId, Date.now());
}

export function removeTagFromUser(userId: string, tagId: string): void {
  db.prepare(
    'DELETE FROM user_tags WHERE user_id = ? AND tag_id = ?'
  ).run(userId, tagId);
}

export function getUserTags(userId: string): Tag[] {
  const rows = db.prepare(`
    SELECT t.* FROM tags t
    JOIN user_tags ut ON ut.tag_id = t.id
    WHERE ut.user_id = ?
  `).all(userId) as any[];
  return rows.map(rowToTag);
}

export function getUsersByTags(tagIds: string[], mode: 'AND' | 'OR'): User[] {
  if (tagIds.length === 0) {
    return [];
  }

  const placeholders = tagIds.map(() => '?').join(',');

  if (mode === 'AND') {
    const rows = db.prepare(`
      SELECT u.* FROM users u
      WHERE id IN (
        SELECT user_id FROM user_tags
        WHERE tag_id IN (${placeholders})
        GROUP BY user_id
        HAVING COUNT(DISTINCT tag_id) = ?
      )
    `).all(...tagIds, tagIds.length) as any[];
    return rows.map(rowToUser);
  } else {
    const rows = db.prepare(`
      SELECT DISTINCT u.* FROM users u
      JOIN user_tags ut ON ut.user_id = u.id
      WHERE ut.tag_id IN (${placeholders})
    `).all(...tagIds) as any[];
    return rows.map(rowToUser);
  }
}

export function getUserTier(userId: string): UserTier {
  const user = db.prepare('SELECT last_active_at FROM users WHERE id = ?').get(userId) as any;
  if (!user) {
    return UserTier.CHURNED;
  }

  const now = Date.now();
  const daysSinceActive = (now - user.last_active_at) / (1000 * 60 * 60 * 24);

  if (daysSinceActive <= 7) {
    return UserTier.ACTIVE;
  } else if (daysSinceActive <= 30) {
    return UserTier.SILENT;
  } else {
    return UserTier.CHURNED;
  }
}

export function updateUserActivity(userId: string): void {
  db.prepare('UPDATE users SET last_active_at = ? WHERE id = ?').run(Date.now(), userId);
}

export function getUsersByTier(tier: UserTier): User[] {
  const now = Date.now();
  const sevenDays = 7 * 24 * 60 * 60 * 1000;
  const thirtyDays = 30 * 24 * 60 * 60 * 1000;

  let query = 'SELECT * FROM users';
  let params: any[] = [];

  if (tier === UserTier.ACTIVE) {
    query += ' WHERE last_active_at >= ?';
    params = [now - sevenDays];
  } else if (tier === UserTier.SILENT) {
    query += ' WHERE last_active_at < ? AND last_active_at >= ?';
    params = [now - sevenDays, now - thirtyDays];
  } else {
    query += ' WHERE last_active_at < ?';
    params = [now - thirtyDays];
  }

  const rows = db.prepare(query).all(...params) as any[];
  return rows.map(rowToUser);
}
