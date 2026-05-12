import getDb from '../db/database';
import { Dashboard } from '../types';

interface DashboardRow {
  id: number;
  name: string;
  description: string | null;
  created_at: string;
}

const mapDashboard = (row: DashboardRow): Dashboard => ({
  id: row.id,
  name: row.name,
  description: row.description ?? undefined,
  createdAt: row.created_at,
});

export const createDashboard = (name: string, description?: string): Dashboard => {
  const db = getDb();
  const result = db.prepare(`
    INSERT INTO dashboards (name, description)
    VALUES (?, ?)
  `).run(name, description ?? null);

  const id = result.lastInsertRowid as number;
  const row = db.prepare(`SELECT * FROM dashboards WHERE id = ?`).get(id) as DashboardRow;
  return mapDashboard(row);
};

export const getDashboardById = (id: number): Dashboard | undefined => {
  const db = getDb();
  const row = db.prepare(`SELECT * FROM dashboards WHERE id = ?`).get(id) as DashboardRow | undefined;
  return row ? mapDashboard(row) : undefined;
};

export const getAllDashboards = (): Dashboard[] => {
  const db = getDb();
  const rows = db.prepare(`SELECT * FROM dashboards ORDER BY created_at DESC`).all() as DashboardRow[];
  return rows.map(mapDashboard);
};

export const updateDashboard = (
  id: number,
  name?: string,
  description?: string
): Dashboard | undefined => {
  const db = getDb();
  const current = getDashboardById(id);
  if (!current) return undefined;

  const newName = name ?? current.name;
  const newDesc = description !== undefined ? description : current.description;

  db.prepare(`
    UPDATE dashboards
    SET name = ?, description = ?
    WHERE id = ?
  `).run(newName, newDesc ?? null, id);

  return getDashboardById(id);
};

export const deleteDashboard = (id: number): boolean => {
  const db = getDb();
  const result = db.prepare(`DELETE FROM dashboards WHERE id = ?`).run(id);
  return result.changes > 0;
};

export const countDashboardCards = (dashboardId: number): number => {
  const db = getDb();
  const row = db.prepare(`
    SELECT COUNT(*) as count FROM chart_cards WHERE dashboard_id = ?
  `).get(dashboardId) as { count: number };
  return row.count;
};
