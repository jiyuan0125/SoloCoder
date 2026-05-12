import { v4 as uuidv4 } from 'uuid';
import db from '../db';
import { Community, CommunityType } from '../types';

function rowToCommunity(row: any): Community {
  return {
    id: row.id,
    name: row.name,
    type: row.type as CommunityType,
    maxMembers: row.max_members,
    currentMembers: row.current_members,
    createdAt: row.created_at
  };
}

export function createCommunity(data: {
  name: string;
  type: CommunityType;
  maxMembers: number;
}): Community {
  if (data.maxMembers <= 0) {
    const err: any = new Error('maxMembers must be greater than 0');
    err.status = 400;
    throw err;
  }

  const existing = db.prepare('SELECT id FROM communities WHERE name = ?').get(data.name);
  if (existing) {
    const err: any = new Error('Community name already exists');
    err.status = 409;
    throw err;
  }

  const id = uuidv4();
  const now = Date.now();

  db.prepare(`
    INSERT INTO communities (id, name, type, max_members, current_members, created_at)
    VALUES (?, ?, ?, ?, 0, ?)
  `).run(id, data.name, data.type, data.maxMembers, now);

  return {
    id,
    name: data.name,
    type: data.type,
    maxMembers: data.maxMembers,
    currentMembers: 0,
    createdAt: now
  };
}

export function getAllCommunities(): Community[] {
  const rows = db.prepare('SELECT * FROM communities ORDER BY created_at DESC').all();
  return rows.map(rowToCommunity);
}

export function getCommunityById(id: string): Community | undefined {
  const row = db.prepare('SELECT * FROM communities WHERE id = ?').get(id);
  return row ? rowToCommunity(row) : undefined;
}

export function deleteCommunity(id: string): void {
  db.prepare('DELETE FROM communities WHERE id = ?').run(id);
}

export function joinCommunity(communityId: string, userId: string): void {
  const community = db.prepare('SELECT * FROM communities WHERE id = ?').get(communityId) as any;
  if (!community) {
    const err: any = new Error('Community not found');
    err.status = 404;
    throw err;
  }

  if (community.current_members >= community.max_members) {
    const err: any = new Error('Community is full');
    err.status = 400;
    throw err;
  }

  const existingMember = db.prepare(
    'SELECT id FROM community_members WHERE community_id = ? AND user_id = ?'
  ).get(communityId, userId);

  if (existingMember) {
    return;
  }

  const userExists = db.prepare('SELECT id FROM users WHERE id = ?').get(userId);
  if (!userExists) {
    const now = Date.now();
    db.prepare(
      'INSERT INTO users (id, name, last_active_at, created_at) VALUES (?, ?, ?, ?)'
    ).run(userId, `user_${userId.slice(0, 8)}`, now, now);
  }

  const tx = db.transaction(() => {
    db.prepare(`
      INSERT INTO community_members (id, community_id, user_id, joined_at)
      VALUES (?, ?, ?, ?)
    `).run(uuidv4(), communityId, userId, Date.now());

    db.prepare(
      'UPDATE communities SET current_members = current_members + 1 WHERE id = ?'
    ).run(communityId);

    db.prepare(`
      INSERT INTO member_records (id, community_id, user_id, action, timestamp)
      VALUES (?, ?, ?, 'join', ?)
    `).run(uuidv4(), communityId, userId, Date.now());
  });

  tx();
}

export function leaveCommunity(communityId: string, userId: string): void {
  const existingMember = db.prepare(
    'SELECT id FROM community_members WHERE community_id = ? AND user_id = ?'
  ).get(communityId, userId);

  if (!existingMember) {
    const err: any = new Error('User is not a member of this community');
    err.status = 404;
    throw err;
  }

  const tx = db.transaction(() => {
    db.prepare(
      'DELETE FROM community_members WHERE community_id = ? AND user_id = ?'
    ).run(communityId, userId);

    db.prepare(
      'UPDATE communities SET current_members = current_members - 1 WHERE id = ?'
    ).run(communityId);

    db.prepare(`
      INSERT INTO member_records (id, community_id, user_id, action, timestamp)
      VALUES (?, ?, ?, 'leave', ?)
    `).run(uuidv4(), communityId, userId, Date.now());
  });

  tx();
}

export function getCommunityMembers(communityId: string): string[] {
  const rows = db.prepare(
    'SELECT user_id FROM community_members WHERE community_id = ?'
  ).all(communityId) as any[];
  return rows.map(r => r.user_id);
}
