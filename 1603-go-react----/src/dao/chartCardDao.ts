import getDb from '../db/database';
import { ChartCard, ChartType, AggregationType } from '../types';

interface ChartCardRow {
  id: number;
  dashboard_id: number;
  title: string;
  chart_type: ChartType;
  metric_name: string;
  aggregation: AggregationType;
  position: string | null;
  data_status: 'available' | 'missing';
  created_at: string;
}

const mapChartCard = (row: ChartCardRow): ChartCard => ({
  id: row.id,
  dashboardId: row.dashboard_id,
  title: row.title,
  chartType: row.chart_type,
  metricName: row.metric_name,
  aggregation: row.aggregation,
  position: row.position ?? undefined,
  dataStatus: row.data_status,
  createdAt: row.created_at,
});

export const createChartCard = (
  dashboardId: number,
  title: string,
  chartType: ChartType,
  metricName: string,
  aggregation: AggregationType,
  position?: string,
  dataStatus: 'available' | 'missing' = 'available'
): ChartCard => {
  const db = getDb();
  const result = db.prepare(`
    INSERT INTO chart_cards (dashboard_id, title, chart_type, metric_name, aggregation, position, data_status)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `).run(dashboardId, title, chartType, metricName, aggregation, position ?? null, dataStatus);

  const id = result.lastInsertRowid as number;
  const row = db.prepare(`SELECT * FROM chart_cards WHERE id = ?`).get(id) as ChartCardRow;
  return mapChartCard(row);
};

export const getChartCardById = (id: number): ChartCard | undefined => {
  const db = getDb();
  const row = db.prepare(`SELECT * FROM chart_cards WHERE id = ?`).get(id) as ChartCardRow | undefined;
  return row ? mapChartCard(row) : undefined;
};

export const getChartCardsByDashboardId = (dashboardId: number): ChartCard[] => {
  const db = getDb();
  const rows = db.prepare(`
    SELECT * FROM chart_cards
    WHERE dashboard_id = ?
    ORDER BY created_at ASC
  `).all(dashboardId) as ChartCardRow[];
  return rows.map(mapChartCard);
};

export const updateChartCard = (
  id: number,
  updates: {
    title?: string;
    chartType?: ChartType;
    metricName?: string;
    aggregation?: AggregationType;
    position?: string;
    dataStatus?: 'available' | 'missing';
  }
): ChartCard | undefined => {
  const db = getDb();
  const current = getChartCardById(id);
  if (!current) return undefined;

  const newTitle = updates.title ?? current.title;
  const newChartType = updates.chartType ?? current.chartType;
  const newMetricName = updates.metricName ?? current.metricName;
  const newAggregation = updates.aggregation ?? current.aggregation;
  const newPosition = updates.position !== undefined ? updates.position : current.position;
  const newDataStatus = updates.dataStatus ?? current.dataStatus;

  db.prepare(`
    UPDATE chart_cards
    SET title = ?, chart_type = ?, metric_name = ?, aggregation = ?, position = ?, data_status = ?
    WHERE id = ?
  `).run(newTitle, newChartType, newMetricName, newAggregation, newPosition ?? null, newDataStatus, id);

  return getChartCardById(id);
};

export const deleteChartCard = (id: number): boolean => {
  const db = getDb();
  const result = db.prepare(`DELETE FROM chart_cards WHERE id = ?`).run(id);
  return result.changes > 0;
};

export const markCardsMissingByMetricName = (metricName: string): void => {
  const db = getDb();
  db.prepare(`
    UPDATE chart_cards
    SET data_status = 'missing'
    WHERE metric_name = ?
  `).run(metricName);
};

export const markCardsAvailableByMetricName = (metricName: string): void => {
  const db = getDb();
  db.prepare(`
    UPDATE chart_cards
    SET data_status = 'available'
    WHERE metric_name = ?
  `).run(metricName);
};
