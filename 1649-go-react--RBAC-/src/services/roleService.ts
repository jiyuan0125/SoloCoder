import db from '../database';
import { v4 as uuidv4 } from 'uuid';
import { logOperation } from './auditService';
import { getPermissionById, Permission } from './permissionService';

export interface Role {
  id: string;
  name: string;
  description: string | null;
  created_at: string;
}

export interface RoleWithPermissions extends Role {
  permissions: Permission[];
  parent_roles: string[];
}

export function createRole(name: string, description?: string): Role {
  const id = uuidv4();
  const stmt = db.prepare(`
    INSERT INTO roles (id, name, description)
    VALUES (?, ?, ?)
  `);
  
  try {
    stmt.run(id, name, description || null);
  } catch (e: any) {
    if (e.code === 'SQLITE_CONSTRAINT_UNIQUE') {
      throw new Error('Role with this name already exists');
    }
    throw e;
  }

  logOperation('CREATE', 'role', id, { name, description });
  return db.prepare('SELECT * FROM roles WHERE id = ?').get(id) as Role;
}

export function getRoleById(id: string): Role | undefined {
  return db.prepare('SELECT * FROM roles WHERE id = ?').get(id) as Role | undefined;
}

export function getRoleByName(name: string): Role | undefined {
  return db.prepare('SELECT * FROM roles WHERE name = ?').get(name) as Role | undefined;
}

export function getAllRoles(): Role[] {
  return db.prepare('SELECT * FROM roles ORDER BY created_at DESC').all() as Role[];
}

export function assignPermissionToRole(roleId: string, permissionId: string): void {
  const role = getRoleById(roleId);
  const permission = getPermissionById(permissionId);
  
  if (!role) {
    throw new Error('Role not found');
  }
  if (!permission) {
    throw new Error('Permission not found');
  }

  const existing = db.prepare('SELECT 1 FROM role_permissions WHERE role_id = ? AND permission_id = ?').get(roleId, permissionId);
  if (existing) {
    return;
  }

  const stmt = db.prepare(`
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (?, ?)
  `);
  stmt.run(roleId, permissionId);

  logOperation('ASSIGN_PERMISSION', 'role', roleId, { roleId, permissionId, permissionName: permission.name });
}

export function removePermissionFromRole(roleId: string, permissionId: string): void {
  const role = getRoleById(roleId);
  if (!role) {
    throw new Error('Role not found');
  }

  const stmt = db.prepare(`
    DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?
  `);
  stmt.run(roleId, permissionId);

  logOperation('REMOVE_PERMISSION', 'role', roleId, { roleId, permissionId });
}

export function getRolePermissions(roleId: string): Permission[] {
  return db.prepare(`
    SELECT p.* FROM permissions p
    JOIN role_permissions rp ON p.id = rp.permission_id
    WHERE rp.role_id = ?
    ORDER BY p.created_at DESC
  `).all(roleId) as Permission[];
}

function wouldCreateCycle(childId: string, parentId: string): boolean {
  const visited = new Set<string>();
  const queue: string[] = [parentId];

  while (queue.length > 0) {
    const current = queue.shift()!;
    
    if (current === childId) {
      return true;
    }
    
    if (visited.has(current)) {
      continue;
    }
    visited.add(current);

    const parents = db.prepare(`
      SELECT parent_role_id FROM role_inheritance
      WHERE child_role_id = ?
    `).all(current) as { parent_role_id: string }[];

    for (const row of parents) {
      if (!visited.has(row.parent_role_id)) {
        queue.push(row.parent_role_id);
      }
    }
  }

  return false;
}

export function setRoleInheritance(childRoleId: string, parentRoleId: string): void {
  if (childRoleId === parentRoleId) {
    throw new Error('Cannot inherit from self');
  }

  const child = getRoleById(childRoleId);
  const parent = getRoleById(parentRoleId);

  if (!child || !parent) {
    throw new Error('Role not found');
  }

  if (wouldCreateCycle(childRoleId, parentRoleId)) {
    throw new Error('Inheritance would create a cycle');
  }

  const existing = db.prepare('SELECT 1 FROM role_inheritance WHERE child_role_id = ? AND parent_role_id = ?').get(childRoleId, parentRoleId);
  if (existing) {
    return;
  }

  const stmt = db.prepare(`
    INSERT INTO role_inheritance (child_role_id, parent_role_id)
    VALUES (?, ?)
  `);
  stmt.run(childRoleId, parentRoleId);

  logOperation('SET_INHERITANCE', 'role', childRoleId, { childRoleId, parentRoleId, childName: child.name, parentName: parent.name });
}

export function removeRoleInheritance(childRoleId: string, parentRoleId: string): void {
  const stmt = db.prepare(`
    DELETE FROM role_inheritance WHERE child_role_id = ? AND parent_role_id = ?
  `);
  stmt.run(childRoleId, parentRoleId);

  logOperation('REMOVE_INHERITANCE', 'role', childRoleId, { childRoleId, parentRoleId });
}

export function getRoleParentIds(roleId: string): string[] {
  const rows = db.prepare(`
    SELECT parent_role_id FROM role_inheritance
    WHERE child_role_id = ?
  `).all(roleId) as { parent_role_id: string }[];
  
  return rows.map(r => r.parent_role_id);
}

export function getAllRoleAncestorIds(roleId: string): string[] {
  const result = new Set<string>();
  const queue: string[] = [roleId];

  while (queue.length > 0) {
    const current = queue.shift()!;
    const parents = getRoleParentIds(current);
    
    for (const parentId of parents) {
      if (!result.has(parentId)) {
        result.add(parentId);
        queue.push(parentId);
      }
    }
  }

  return Array.from(result);
}

export function getUsersAssignedToRole(roleId: string): { user_id: string }[] {
  return db.prepare(`
    SELECT user_id FROM user_roles WHERE role_id = ?
  `).all(roleId) as { user_id: string }[];
}

export function deleteRole(roleId: string): void {
  const role = getRoleById(roleId);
  if (!role) {
    throw new Error('Role not found');
  }

  const assignedUsers = getUsersAssignedToRole(roleId);
  if (assignedUsers.length > 0) {
    const userIds = assignedUsers.map(u => u.user_id);
    throw new Error(`Role is still assigned to users: ${userIds.join(', ')}`);
  }

  const tx = db.transaction(() => {
    db.prepare('DELETE FROM role_permissions WHERE role_id = ?').run(roleId);
    db.prepare('DELETE FROM role_inheritance WHERE child_role_id = ? OR parent_role_id = ?').run(roleId, roleId);
    db.prepare('DELETE FROM roles WHERE id = ?').run(roleId);
  });

  tx();
  logOperation('DELETE', 'role', roleId, { roleName: role.name });
}

export function getRoleWithPermissions(roleId: string): RoleWithPermissions | undefined {
  const role = getRoleById(roleId);
  if (!role) {
    return undefined;
  }

  return {
    ...role,
    permissions: getRolePermissions(roleId),
    parent_roles: getRoleParentIds(roleId)
  };
}
