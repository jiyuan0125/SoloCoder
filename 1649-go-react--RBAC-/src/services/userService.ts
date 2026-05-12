import db from '../database';
import { v4 as uuidv4 } from 'uuid';
import { logOperation } from './auditService';
import { getRoleById, getAllRoleAncestorIds } from './roleService';
import { Permission } from './permissionService';

export interface User {
  id: string;
  name: string;
  created_at: string;
}

export function createUser(name: string, id?: string): User {
  const userId = id || uuidv4();
  
  const existing = db.prepare('SELECT * FROM users WHERE id = ?').get(userId);
  if (existing) {
    return existing as User;
  }

  const stmt = db.prepare(`
    INSERT INTO users (id, name)
    VALUES (?, ?)
  `);
  stmt.run(userId, name);

  logOperation('CREATE', 'user', userId, { name });
  return db.prepare('SELECT * FROM users WHERE id = ?').get(userId) as User;
}

export function getUserById(id: string): User | undefined {
  return db.prepare('SELECT * FROM users WHERE id = ?').get(id) as User | undefined;
}

export function getAllUsers(): User[] {
  return db.prepare('SELECT * FROM users ORDER BY created_at DESC').all() as User[];
}

export function assignRoleToUser(userId: string, roleId: string): void {
  let user = getUserById(userId);
  if (!user) {
    user = createUser(`user_${userId}`, userId);
  }

  const role = getRoleById(roleId);
  if (!role) {
    throw new Error('Role not found');
  }

  const existing = db.prepare('SELECT 1 FROM user_roles WHERE user_id = ? AND role_id = ?').get(userId, roleId);
  if (existing) {
    throw new Error('User already has this role');
  }

  const stmt = db.prepare(`
    INSERT INTO user_roles (user_id, role_id)
    VALUES (?, ?)
  `);
  stmt.run(userId, roleId);

  logOperation('ASSIGN_ROLE', 'user', userId, { userId, roleId, roleName: role.name });
}

export function removeRoleFromUser(userId: string, roleId: string): void {
  const role = getRoleById(roleId);
  
  const stmt = db.prepare(`
    DELETE FROM user_roles WHERE user_id = ? AND role_id = ?
  `);
  stmt.run(userId, roleId);

  logOperation('REMOVE_ROLE', 'user', userId, { userId, roleId, roleName: role?.name });
}

export function getUserDirectRoleIds(userId: string): string[] {
  const rows = db.prepare(`
    SELECT role_id FROM user_roles WHERE user_id = ?
  `).all(userId) as { role_id: string }[];
  
  return rows.map(r => r.role_id);
}

export function getUserAllRoleIds(userId: string): string[] {
  const directRoleIds = getUserDirectRoleIds(userId);
  const allRoleIds = new Set<string>(directRoleIds);

  for (const roleId of directRoleIds) {
    const ancestors = getAllRoleAncestorIds(roleId);
    for (const ancestorId of ancestors) {
      allRoleIds.add(ancestorId);
    }
  }

  return Array.from(allRoleIds);
}

export function getUserRoles(userId: string): Array<{ role_id: string; role_name: string | null }> {
  const roleIds = getUserAllRoleIds(userId);
  
  if (roleIds.length === 0) {
    return [];
  }

  const placeholders = roleIds.map(() => '?').join(',');
  return db.prepare(`
    SELECT id as role_id, name as role_name FROM roles
    WHERE id IN (${placeholders})
  `).all(...roleIds) as Array<{ role_id: string; role_name: string | null }>;
}

export function getUserPermissions(userId: string): Permission[] {
  const roleIds = getUserAllRoleIds(userId);
  
  if (roleIds.length === 0) {
    return [];
  }

  const placeholders = roleIds.map(() => '?').join(',');
  return db.prepare(`
    SELECT DISTINCT p.* FROM permissions p
    JOIN role_permissions rp ON p.id = rp.permission_id
    WHERE rp.role_id IN (${placeholders})
    ORDER BY p.created_at DESC
  `).all(...roleIds) as Permission[];
}

export function checkUserPermission(userId: string, resource: string, action: string): boolean {
  const roleIds = getUserAllRoleIds(userId);
  
  if (roleIds.length === 0) {
    return false;
  }

  const placeholders = roleIds.map(() => '?').join(',');
  const params = [resource, action, ...roleIds];

  const result = db.prepare(`
    SELECT 1 FROM permissions p
    JOIN role_permissions rp ON p.id = rp.permission_id
    WHERE p.resource = ? AND p.action = ? AND rp.role_id IN (${placeholders})
    LIMIT 1
  `).get(...params);

  return !!result;
}
