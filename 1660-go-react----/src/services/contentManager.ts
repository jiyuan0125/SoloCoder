import db from '../database';
import type { Post, PostStatus, ContentType } from '../types';

function generateId(): string {
  return 'post_' + Date.now() + '_' + Math.random().toString(36).substring(2, 9);
}

export function createPost(
  userId: string,
  content: string,
  contentType: ContentType,
  status: PostStatus
): Post {
  const id = generateId();
  const now = Date.now();
  db.prepare(`
    INSERT INTO posts (id, user_id, content, content_type, status, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
  `).run(id, userId, content, contentType, status, now);
  return { id, user_id: userId, content, content_type: contentType, status, created_at: now };
}

export function getPost(postId: string): Post | null {
  return db.prepare('SELECT * FROM posts WHERE id = ?').get(postId) as Post | null;
}

export function updatePostStatus(postId: string, status: PostStatus): boolean {
  const result = db.prepare('UPDATE posts SET status = ? WHERE id = ?').run(status, postId);
  return result.changes > 0;
}

export function getAuditQueue(limit: number = 100, offset: number = 0): Post[] {
  return db.prepare(`
    SELECT * FROM posts
    WHERE status IN ('pending_review', 'suspicious')
    ORDER BY created_at ASC
    LIMIT ? OFFSET ?
  `).all(limit, offset) as Post[];
}

export function getAuditQueueCount(): number {
  const row = db.prepare(`
    SELECT COUNT(*) as count FROM posts
    WHERE status IN ('pending_review', 'suspicious')
  `).get() as { count: number };
  return row.count;
}