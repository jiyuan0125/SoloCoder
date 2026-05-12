import { v4 as uuidv4 } from 'uuid';
import { getDatabase } from '../database';
import { User, RiskTag } from '../types';
import { calculateUserRiskScore } from '../risk';

const db = getDatabase();

export function getUserById(id: string): User | null {
  const row = db.prepare('SELECT * FROM users WHERE id = ?').get(id) as any;
  if (!row) return null;
  return {
    ...row,
    riskTags: JSON.parse(row.riskTags),
    isBlacklisted: Boolean(row.isBlacklisted)
  };
}

export function getUserByEmail(email: string): User | null {
  const row = db.prepare('SELECT * FROM users WHERE email = ?').get(email) as any;
  if (!row) return null;
  return {
    ...row,
    riskTags: JSON.parse(row.riskTags),
    isBlacklisted: Boolean(row.isBlacklisted)
  };
}

export function createUser(data: { name: string; email: string; phone: string }): User {
  const id = uuidv4();
  const now = Date.now();
  const existing = getUserByEmail(data.email);
  if (existing) {
    throw new Error('User with this email already exists');
  }
  db.prepare(`
    INSERT INTO users (id, name, email, phone, riskTags, riskScore, isBlacklisted, createdAt)
    VALUES (?, ?, ?, ?, '[]', 0, 0, ?)
  `).run(id, data.name, data.email, data.phone, now);
  
  return getUserById(id)!;
}

export function updateUser(id: string, data: Partial<{ name: string; phone: string }>): User | null {
  const user = getUserById(id);
  if (!user) return null;
  
  const updates: string[] = [];
  const values: any[] = [];
  
  if (data.name !== undefined) {
    updates.push('name = ?');
    values.push(data.name);
  }
  if (data.phone !== undefined) {
    updates.push('phone = ?');
    values.push(data.phone);
  }
  
  if (updates.length > 0) {
    values.push(id);
    db.prepare(`UPDATE users SET ${updates.join(', ')} WHERE id = ?`).run(...values);
  }
  
  return getUserById(id);
}

export function addRiskTag(userId: string, tag: RiskTag): User | null {
  const user = getUserById(userId);
  if (!user) return null;
  
  if (!user.riskTags.includes(tag)) {
    const newTags = [...user.riskTags, tag];
    db.prepare('UPDATE users SET riskTags = ? WHERE id = ?').run(
      JSON.stringify(newTags),
      userId
    );
  }
  
  return getUserById(userId);
}

export function removeRiskTag(userId: string, tag: RiskTag): User | null {
  const user = getUserById(userId);
  if (!user) return null;
  
  const newTags = user.riskTags.filter(t => t !== tag);
  if (newTags.length !== user.riskTags.length) {
    db.prepare('UPDATE users SET riskTags = ? WHERE id = ?').run(
      JSON.stringify(newTags),
      userId
    );
  }
  
  return getUserById(userId);
}

export function setUserBlacklisted(userId: string, blacklisted: boolean, transaction?: any): User | null {
  const database = transaction || db;
  const user = getUserById(userId);
  if (!user) return null;
  
  database.prepare('UPDATE users SET isBlacklisted = ? WHERE id = ?').run(
    blacklisted ? 1 : 0,
    userId
  );
  
  const newScore = calculateUserRiskScore(userId, database);
  database.prepare('UPDATE users SET riskScore = ? WHERE id = ?').run(newScore, userId);
  
  return getUserById(userId);
}

export function listUsers(): User[] {
  const rows = db.prepare('SELECT * FROM users ORDER BY createdAt DESC').all() as any[];
  return rows.map(row => ({
    ...row,
    riskTags: JSON.parse(row.riskTags),
    isBlacklisted: Boolean(row.isBlacklisted)
  }));
}

export function decrementUserRiskScore(userId: string): User | null {
  const user = getUserById(userId);
  if (!user) return null;
  
  const newScore = Math.max(0, user.riskScore - 10);
  db.prepare('UPDATE users SET riskScore = ? WHERE id = ?').run(newScore, userId);
  
  return getUserById(userId);
}

export function updateUserRiskScore(userId: string, database?: any): void {
  const dbInstance = database || db;
  const score = calculateUserRiskScore(userId, dbInstance);
  dbInstance.prepare('UPDATE users SET riskScore = ? WHERE id = ?').run(score, userId);
}
