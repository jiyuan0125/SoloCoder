import Database from 'better-sqlite3';
import { v4 as uuidv4 } from 'uuid';
import { PasswordPolicy, User, PasswordHistory } from './types';

const db = new Database('./password-service.db');

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS policies (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL UNIQUE,
      minLength INTEGER NOT NULL DEFAULT 8,
      requireUppercase INTEGER NOT NULL DEFAULT 0,
      requireLowercase INTEGER NOT NULL DEFAULT 0,
      requireNumber INTEGER NOT NULL DEFAULT 0,
      requireSpecial INTEGER NOT NULL DEFAULT 0,
      forbidUsername INTEGER NOT NULL DEFAULT 0,
      historyCount INTEGER NOT NULL DEFAULT 3,
      createdAt INTEGER NOT NULL,
      updatedAt INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS users (
      id TEXT PRIMARY KEY,
      username TEXT NOT NULL UNIQUE,
      passwordHash TEXT NOT NULL,
      policyId TEXT NOT NULL,
      createdAt INTEGER NOT NULL,
      updatedAt INTEGER NOT NULL,
      FOREIGN KEY (policyId) REFERENCES policies(id)
    );

    CREATE TABLE IF NOT EXISTS passwordHistory (
      id TEXT PRIMARY KEY,
      userId TEXT NOT NULL,
      passwordHash TEXT NOT NULL,
      createdAt INTEGER NOT NULL,
      FOREIGN KEY (userId) REFERENCES users(id)
    );

    CREATE INDEX IF NOT EXISTS idx_passwordHistory_user ON passwordHistory(userId);
  `);

  const defaultPolicy = db.prepare('SELECT * FROM policies WHERE name = ?').get('default');
  if (!defaultPolicy) {
    const now = Date.now();
    db.prepare(`
      INSERT INTO policies (
        id, name, minLength, requireUppercase, requireLowercase, 
        requireNumber, requireSpecial, forbidUsername, historyCount,
        createdAt, updatedAt
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `).run(
      uuidv4(),
      'default',
      8,
      0,
      1,
      0,
      0,
      0,
      3,
      now,
      now
    );
  }
}

export function getPolicyById(id: string): PasswordPolicy | undefined {
  const row = db.prepare('SELECT * FROM policies WHERE id = ?').get(id);
  return row ? mapPolicy(row) : undefined;
}

export function getPolicyByName(name: string): PasswordPolicy | undefined {
  const row = db.prepare('SELECT * FROM policies WHERE name = ?').get(name);
  return row ? mapPolicy(row) : undefined;
}

export function createPolicy(policy: Omit<PasswordPolicy, 'id' | 'createdAt' | 'updatedAt'>): PasswordPolicy {
  const id = uuidv4();
  const now = Date.now();
  db.prepare(`
    INSERT INTO policies (
      id, name, minLength, requireUppercase, requireLowercase,
      requireNumber, requireSpecial, forbidUsername, historyCount,
      createdAt, updatedAt
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    id,
    policy.name,
    policy.minLength,
    policy.requireUppercase ? 1 : 0,
    policy.requireLowercase ? 1 : 0,
    policy.requireNumber ? 1 : 0,
    policy.requireSpecial ? 1 : 0,
    policy.forbidUsername ? 1 : 0,
    policy.historyCount,
    now,
    now
  );
  return { ...policy, id, createdAt: now, updatedAt: now };
}

export function updatePolicy(id: string, updates: Partial<Omit<PasswordPolicy, 'id' | 'createdAt' | 'updatedAt'>>): void {
  const fields: string[] = [];
  const values: any[] = [];
  
  if (updates.name !== undefined) {
    fields.push('name = ?');
    values.push(updates.name);
  }
  if (updates.minLength !== undefined) {
    fields.push('minLength = ?');
    values.push(updates.minLength);
  }
  if (updates.requireUppercase !== undefined) {
    fields.push('requireUppercase = ?');
    values.push(updates.requireUppercase ? 1 : 0);
  }
  if (updates.requireLowercase !== undefined) {
    fields.push('requireLowercase = ?');
    values.push(updates.requireLowercase ? 1 : 0);
  }
  if (updates.requireNumber !== undefined) {
    fields.push('requireNumber = ?');
    values.push(updates.requireNumber ? 1 : 0);
  }
  if (updates.requireSpecial !== undefined) {
    fields.push('requireSpecial = ?');
    values.push(updates.requireSpecial ? 1 : 0);
  }
  if (updates.forbidUsername !== undefined) {
    fields.push('forbidUsername = ?');
    values.push(updates.forbidUsername ? 1 : 0);
  }
  if (updates.historyCount !== undefined) {
    fields.push('historyCount = ?');
    values.push(updates.historyCount);
  }
  
  fields.push('updatedAt = ?');
  values.push(Date.now());
  values.push(id);
  
  db.prepare(`UPDATE policies SET ${fields.join(', ')} WHERE id = ?`).run(...values);
}

export function getUserById(id: string): User | undefined {
  const row = db.prepare('SELECT * FROM users WHERE id = ?').get(id);
  return row ? mapUser(row) : undefined;
}

export function getUserByUsername(username: string): User | undefined {
  const row = db.prepare('SELECT * FROM users WHERE username = ?').get(username);
  return row ? mapUser(row) : undefined;
}

export function createUser(user: Omit<User, 'id' | 'createdAt' | 'updatedAt'>): User {
  const id = uuidv4();
  const now = Date.now();
  db.prepare(`
    INSERT INTO users (id, username, passwordHash, policyId, createdAt, updatedAt)
    VALUES (?, ?, ?, ?, ?, ?)
  `).run(id, user.username, user.passwordHash, user.policyId, now, now);
  return { ...user, id, createdAt: now, updatedAt: now };
}

export function getUserPasswordHistory(userId: string): PasswordHistory[] {
  const rows = db.prepare(
    'SELECT * FROM passwordHistory WHERE userId = ? ORDER BY createdAt DESC'
  ).all(userId);
  return (rows as any[]).map(mapHistory);
}

export function changePasswordAtomically(
  userId: string,
  newPasswordHash: string,
  historyLimit: number
): void {
  const transaction = db.transaction(() => {
    const now = Date.now();
    
    db.prepare(`
      INSERT INTO passwordHistory (id, userId, passwordHash, createdAt)
      VALUES (?, ?, ?, ?)
    `).run(uuidv4(), userId, newPasswordHash, now);
    
    db.prepare(`
      UPDATE users SET passwordHash = ?, updatedAt = ? WHERE id = ?
    `).run(newPasswordHash, now, userId);
    
    const history = db.prepare(
      'SELECT id FROM passwordHistory WHERE userId = ? ORDER BY createdAt DESC'
    ).all(userId) as { id: string }[];
    
    if (history.length > historyLimit) {
      const toDelete = history.slice(historyLimit);
      const deleteStmt = db.prepare('DELETE FROM passwordHistory WHERE id = ?');
      for (const item of toDelete) {
        deleteStmt.run(item.id);
      }
    }
  });
  
  transaction();
}

function mapPolicy(row: any): PasswordPolicy {
  return {
    id: row.id,
    name: row.name,
    minLength: row.minLength,
    requireUppercase: row.requireUppercase === 1,
    requireLowercase: row.requireLowercase === 1,
    requireNumber: row.requireNumber === 1,
    requireSpecial: row.requireSpecial === 1,
    forbidUsername: row.forbidUsername === 1,
    historyCount: row.historyCount,
    createdAt: row.createdAt,
    updatedAt: row.updatedAt
  };
}

function mapUser(row: any): User {
  return {
    id: row.id,
    username: row.username,
    passwordHash: row.passwordHash,
    policyId: row.policyId,
    createdAt: row.createdAt,
    updatedAt: row.updatedAt
  };
}

function mapHistory(row: any): PasswordHistory {
  return {
    id: row.id,
    userId: row.userId,
    passwordHash: row.passwordHash,
    createdAt: row.createdAt
  };
}

export { db };
