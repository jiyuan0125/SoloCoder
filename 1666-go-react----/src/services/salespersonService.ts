import { v4 as uuidv4 } from 'uuid';
import { db } from '../db';
import { Salesperson, CreateSalespersonRequest } from '../types';

interface DbSalesperson {
  id: string;
  name: string;
  team_id: string;
  created_at: string;
}

function mapToSalesperson(dbRow: DbSalesperson): Salesperson {
  return {
    id: dbRow.id,
    name: dbRow.name,
    teamId: dbRow.team_id,
    createdAt: dbRow.created_at,
  };
}

export function getSalespeople(): Salesperson[] {
  const rows = db.prepare('SELECT * FROM salespeople ORDER BY created_at DESC').all() as DbSalesperson[];
  return rows.map(mapToSalesperson);
}

export function getSalespersonById(id: string): Salesperson | null {
  const row = db.prepare('SELECT * FROM salespeople WHERE id = ?').get(id) as DbSalesperson | undefined;
  return row ? mapToSalesperson(row) : null;
}

export function createSalesperson(data: CreateSalespersonRequest): Salesperson {
  const id = uuidv4();
  const stmt = db.prepare(`
    INSERT INTO salespeople (id, name, team_id)
    VALUES (?, ?, ?)
  `);
  stmt.run(id, data.name, data.teamId);
  const created = getSalespersonById(id);
  if (!created) {
    throw new Error('Failed to create salesperson');
  }
  return created;
}

export function updateSalesperson(id: string, data: Partial<CreateSalespersonRequest>): Salesperson | null {
  const existing = getSalespersonById(id);
  if (!existing) return null;

  const updates: string[] = [];
  const values: any[] = [];

  if (data.name !== undefined) {
    updates.push('name = ?');
    values.push(data.name);
  }
  if (data.teamId !== undefined) {
    updates.push('team_id = ?');
    values.push(data.teamId);
  }

  if (updates.length === 0) return existing;

  values.push(id);
  const stmt = db.prepare(`UPDATE salespeople SET ${updates.join(', ')} WHERE id = ?`);
  stmt.run(...values);

  return getSalespersonById(id);
}

export function deleteSalesperson(id: string): boolean {
  const stmt = db.prepare('DELETE FROM salespeople WHERE id = ?');
  const result = stmt.run(id);
  return (result.changes ?? 0) > 0;
}

export function salespersonExists(id: string): boolean {
  const row = db.prepare('SELECT id FROM salespeople WHERE id = ?').get(id);
  return !!row;
}

export function getTeamSalespeople(teamId: string): Salesperson[] {
  const rows = db.prepare('SELECT * FROM salespeople WHERE team_id = ? ORDER BY created_at DESC').all(teamId) as DbSalesperson[];
  return rows.map(mapToSalesperson);
}
