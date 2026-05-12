import { v4 as uuidv4 } from 'uuid';
import { db } from '../database';
import type { Service } from '../types';

export function getAllServices(): Service[] {
  const stmt = db.prepare('SELECT * FROM services ORDER BY created_at DESC');
  return stmt.all() as Service[];
}

export function getServiceById(id: string): Service | undefined {
  const stmt = db.prepare('SELECT * FROM services WHERE id = ?');
  return stmt.get(id) as Service | undefined;
}

export function createService(data: Omit<Service, 'id' | 'created_at' | 'updated_at'>): Service {
  const existing = db
    .prepare('SELECT * FROM services WHERE route_prefix = ?')
    .get(data.route_prefix) as Service | undefined;
  
  if (existing) {
    const error: Error & { statusCode?: number } = new Error(
      `Service with route prefix "${data.route_prefix}" already exists`
    );
    error.statusCode = 409;
    throw error;
  }

  const now = Date.now();
  const id = uuidv4();
  const stmt = db.prepare(`
    INSERT INTO services (id, name, upstream_url, route_prefix, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?)
  `);
  stmt.run(id, data.name, data.upstream_url, data.route_prefix, now, now);
  
  return getServiceById(id)!;
}

export function updateService(
  id: string,
  data: Partial<Pick<Service, 'name' | 'upstream_url' | 'route_prefix'>>
): Service | undefined {
  const existing = getServiceById(id);
  if (!existing) return undefined;

  if (data.route_prefix && data.route_prefix !== existing.route_prefix) {
    const prefixConflict = db
      .prepare('SELECT * FROM services WHERE route_prefix = ? AND id != ?')
      .get(data.route_prefix, id) as Service | undefined;
    
    if (prefixConflict) {
      const error: Error & { statusCode?: number } = new Error(
        `Service with route prefix "${data.route_prefix}" already exists`
      );
      error.statusCode = 409;
      throw error;
    }
  }

  const now = Date.now();
  const updates: string[] = [];
  const values: unknown[] = [];

  if (data.name !== undefined) {
    updates.push('name = ?');
    values.push(data.name);
  }
  if (data.upstream_url !== undefined) {
    updates.push('upstream_url = ?');
    values.push(data.upstream_url);
  }
  if (data.route_prefix !== undefined) {
    updates.push('route_prefix = ?');
    values.push(data.route_prefix);
  }
  
  updates.push('updated_at = ?');
  values.push(now);
  values.push(id);

  const stmt = db.prepare(`UPDATE services SET ${updates.join(', ')} WHERE id = ?`);
  stmt.run(...values);
  
  return getServiceById(id);
}

export function deleteService(id: string): boolean {
  const stmt = db.prepare('DELETE FROM services WHERE id = ?');
  const result = stmt.run(id);
  return result.changes > 0;
}

export function getAllRoutePrefixes(): Service[] {
  const stmt = db.prepare('SELECT * FROM services ORDER BY length(route_prefix) DESC');
  return stmt.all() as Service[];
}
