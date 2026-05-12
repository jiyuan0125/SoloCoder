import db from '../database';
import { v4 as uuidv4 } from 'uuid';
import { logOperation } from './auditService';

export interface Permission {
  id: string;
  name: string;
  resource: string;
  action: string;
  created_at: string;
}

const permissionRegex = /^[a-zA-Z0-9_-]+:[a-zA-Z0-9_-]+$/;

export function parsePermissionName(name: string): { resource: string; action: string } | null {
  if (!permissionRegex.test(name)) {
    return null;
  }
  const [resource, action] = name.split(':');
  return { resource, action };
}

export function createPermission(name: string): Permission {
  const parsed = parsePermissionName(name);
  if (!parsed) {
    throw new Error('Invalid permission format. Expected format: "resource:action" (e.g., "user:read", "order:write")');
  }

  const id = uuidv4();
  const stmt = db.prepare(`
    INSERT INTO permissions (id, name, resource, action)
    VALUES (?, ?, ?, ?)
  `);
  
  try {
    stmt.run(id, name, parsed.resource, parsed.action);
  } catch (e: any) {
    if (e.code === 'SQLITE_CONSTRAINT_UNIQUE') {
      return db.prepare('SELECT * FROM permissions WHERE name = ?').get(name) as Permission;
    }
    throw e;
  }

  logOperation('CREATE', 'permission', id, { name, resource: parsed.resource, action: parsed.action });
  return db.prepare('SELECT * FROM permissions WHERE id = ?').get(id) as Permission;
}

export function getPermissionById(id: string): Permission | undefined {
  return db.prepare('SELECT * FROM permissions WHERE id = ?').get(id) as Permission | undefined;
}

export function getPermissionByName(name: string): Permission | undefined {
  return db.prepare('SELECT * FROM permissions WHERE name = ?').get(name) as Permission | undefined;
}

export function getAllPermissions(): Permission[] {
  return db.prepare('SELECT * FROM permissions ORDER BY created_at DESC').all() as Permission[];
}
