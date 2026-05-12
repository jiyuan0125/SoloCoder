import getDb from '../db/database';
import { Metric, MetricData } from '../types';

interface MetricRow {
  id: number;
  name: string;
  description: string | null;
  created_at: string;
}

interface MetricDataRow {
  id: number;
  metric_id: number;
  value: number;
  timestamp: string;
}

const mapMetric = (row: MetricRow): Metric => ({
  id: row.id,
  name: row.name,
  description: row.description ?? undefined,
  createdAt: row.created_at,
});

const mapMetricData = (row: MetricDataRow): MetricData => ({
  id: row.id,
  metricId: row.metric_id,
  value: row.value,
  timestamp: row.timestamp,
});

export const createMetric = (
  name: string,
  description: string | undefined,
  data: Array<{ value: number; timestamp: string }>
): Metric => {
  const db = getDb();
  const insertMetric = db.prepare(`
    INSERT INTO metrics (name, description)
    VALUES (?, ?)
  `);
  const insertData = db.prepare(`
    INSERT INTO metric_data (metric_id, value, timestamp)
    VALUES (?, ?, ?)
  `);

  const result = insertMetric.run(name, description ?? null);
  const metricId = result.lastInsertRowid as number;

  for (const d of data) {
    insertData.run(metricId, d.value, d.timestamp);
  }

  const row = db.prepare(`SELECT * FROM metrics WHERE id = ?`).get(metricId) as MetricRow;
  return mapMetric(row);
};

export const getMetricById = (id: number): Metric | undefined => {
  const db = getDb();
  const row = db.prepare(`SELECT * FROM metrics WHERE id = ?`).get(id) as MetricRow | undefined;
  return row ? mapMetric(row) : undefined;
};

export const getMetricByName = (name: string): Metric | undefined => {
  const db = getDb();
  const row = db.prepare(`SELECT * FROM metrics WHERE name = ?`).get(name) as MetricRow | undefined;
  return row ? mapMetric(row) : undefined;
};

export const getAllMetrics = (): Metric[] => {
  const db = getDb();
  const rows = db.prepare(`SELECT * FROM metrics ORDER BY created_at DESC`).all() as MetricRow[];
  return rows.map(mapMetric);
};

export const updateMetric = (
  id: number,
  name?: string,
  description?: string
): Metric | undefined => {
  const db = getDb();
  const current = getMetricById(id);
  if (!current) return undefined;

  const newName = name ?? current.name;
  const newDesc = description !== undefined ? description : current.description;

  db.prepare(`
    UPDATE metrics
    SET name = ?, description = ?
    WHERE id = ?
  `).run(newName, newDesc ?? null, id);

  return getMetricById(id);
};

export const deleteMetric = (id: number): boolean => {
  const db = getDb();
  const metric = getMetricById(id);
  if (!metric) return false;

  const result = db.prepare(`DELETE FROM metrics WHERE id = ?`).run(id);
  return result.changes > 0;
};

export const getMetricData = (metricId: number, start: string, end: string): MetricData[] => {
  const db = getDb();
  const rows = db.prepare(`
    SELECT * FROM metric_data
    WHERE metric_id = ? AND timestamp >= ? AND timestamp <= ?
    ORDER BY timestamp ASC
  `).all(metricId, start, end) as MetricDataRow[];
  return rows.map(mapMetricData);
};

export const addMetricData = (
  metricId: number,
  data: Array<{ value: number; timestamp: string }>
): void => {
  const db = getDb();
  const insert = db.prepare(`
    INSERT INTO metric_data (metric_id, value, timestamp)
    VALUES (?, ?, ?)
  `);
  for (const d of data) {
    insert.run(metricId, d.value, d.timestamp);
  }
};
