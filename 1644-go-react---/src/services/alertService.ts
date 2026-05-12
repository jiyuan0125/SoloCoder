import { v4 as uuidv4 } from 'uuid';
import { run, get, all } from '../database';
import { Alert, AlertLevel, AlertStatus } from '../types';

const AGGREGATION_WINDOW_MS = 5 * 60 * 1000;
const ESCALATION_TIMEOUT_MS = 15 * 60 * 1000;

export const getAlertKey = (name: string, level: AlertLevel, sourceSystem: string): string => {
  return `${sourceSystem}:${level}:${name}`;
};

export const createAlert = async (
  name: string,
  level: AlertLevel,
  sourceSystem: string,
  description: string = '',
  metrics: Record<string, any> = {}
): Promise<Alert> => {
  const now = Date.now();
  const fiveMinutesAgo = now - AGGREGATION_WINDOW_MS;

  const existingAlert = await get<any>(
    `SELECT * FROM alerts 
     WHERE name = ? 
       AND level = ? 
       AND sourceSystem = ? 
       AND status IN ('open', 'acknowledged')
       AND lastOccurrenceAt >= ?
     ORDER BY lastOccurrenceAt DESC
     LIMIT 1`,
    [name, level, sourceSystem, fiveMinutesAgo]
  );

  if (existingAlert) {
    const updatedCount = existingAlert.count + 1;
    await run(
      `UPDATE alerts 
       SET count = ?, 
           lastOccurrenceAt = ?,
           description = ?,
           metrics = ?
       WHERE id = ?`,
      [updatedCount, now, description, JSON.stringify(metrics), existingAlert.id]
    );

    return {
      ...existingAlert,
      count: updatedCount,
      lastOccurrenceAt: now,
      description,
      metrics
    };
  }

  const id = uuidv4();
  const alert: Alert = {
    id,
    name,
    level,
    sourceSystem,
    description,
    metrics,
    status: 'open',
    originalLevel: level,
    count: 1,
    createdAt: now,
    lastOccurrenceAt: now,
    acknowledgedAt: null,
    resolvedAt: null,
    escalationTime: level === 'critical' ? now + ESCALATION_TIMEOUT_MS : now
  };

  await run(
    `INSERT INTO alerts (id, name, level, sourceSystem, description, metrics, status, originalLevel, count, createdAt, lastOccurrenceAt, acknowledgedAt, resolvedAt, escalationTime)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      id,
      name,
      level,
      sourceSystem,
      description,
      JSON.stringify(metrics),
      alert.status,
      alert.originalLevel,
      alert.count,
      alert.createdAt,
      alert.lastOccurrenceAt,
      alert.acknowledgedAt,
      alert.resolvedAt,
      alert.escalationTime
    ]
  );

  return alert;
};

export const acknowledgeAlert = async (id: string): Promise<Alert | null> => {
  const alert = await get<Alert>(`SELECT * FROM alerts WHERE id = ?`, [id]);
  if (!alert) return null;

  if (alert.status === 'acknowledged' || alert.status === 'resolved') {
    return alert;
  }

  const now = Date.now();
  await run(
    `UPDATE alerts SET status = 'acknowledged', acknowledgedAt = ? WHERE id = ?`,
    [now, id]
  );

  return { ...alert, status: 'acknowledged', acknowledgedAt: now };
};

export const resolveAlert = async (id: string): Promise<Alert | null> => {
  const alert = await get<Alert>(`SELECT * FROM alerts WHERE id = ?`, [id]);
  if (!alert) return null;

  if (alert.status === 'resolved') {
    return alert;
  }

  const now = Date.now();
  await run(
    `UPDATE alerts SET status = 'resolved', resolvedAt = ? WHERE id = ?`,
    [now, id]
  );

  return { ...alert, status: 'resolved', resolvedAt: now };
};

export const getAlerts = async (
  level?: AlertLevel,
  status?: AlertStatus
): Promise<Alert[]> => {
  let sql = `SELECT * FROM alerts WHERE 1=1`;
  const params: any[] = [];

  if (level) {
    sql += ` AND level = ?`;
    params.push(level);
  }

  if (status) {
    sql += ` AND status = ?`;
    params.push(status);
  }

  sql += ` ORDER BY createdAt DESC`;

  const rows = await all<any>(sql, params);
  return rows.map(row => ({
    ...row,
    metrics: JSON.parse(row.metrics)
  }));
};

export const getAlertById = async (id: string): Promise<Alert | null> => {
  const row = await get<any>(`SELECT * FROM alerts WHERE id = ?`, [id]);
  if (!row) return null;

  return {
    ...row,
    metrics: JSON.parse(row.metrics)
  };
};

export const checkAndEscalateAlerts = async (): Promise<string[]> => {
  const now = Date.now();
  const alertsToEscalate = await all<Alert>(
    `SELECT * FROM alerts 
     WHERE status = 'open' 
       AND level = 'critical' 
       AND escalationTime <= ?
       AND originalLevel = 'critical'`,
    [now]
  );

  const escalatedIds: string[] = [];

  for (const alert of alertsToEscalate) {
    await run(
      `UPDATE alerts SET level = 'emergency' WHERE id = ?`,
      [alert.id]
    );
    escalatedIds.push(alert.id);
  }

  return escalatedIds;
};

export const findOrCreateAggregation = async (
  name: string,
  level: AlertLevel,
  sourceSystem: string,
  description: string,
  metrics: Record<string, any>
): Promise<{ aggregated: any; isNewWindow: boolean }> => {
  const alertKey = getAlertKey(name, level, sourceSystem);
  const now = Date.now();

  const existingAgg = await get<any>(
    `SELECT * FROM aggregated_alerts WHERE alertKey = ?`,
    [alertKey]
  );

  if (existingAgg) {
    const descriptions = JSON.parse(existingAgg.descriptions);
    const metricsList = JSON.parse(existingAgg.metricsList);

    descriptions.push(description);
    metricsList.push(metrics);

    const newWindowEnd = now + AGGREGATION_WINDOW_MS;

    await run(
      `UPDATE aggregated_alerts 
       SET count = count + 1, 
           descriptions = ?, 
           metricsList = ?, 
           windowEndTime = ?
       WHERE alertKey = ?`,
      [JSON.stringify(descriptions), JSON.stringify(metricsList), newWindowEnd, alertKey]
    );

    return {
      aggregated: {
        ...existingAgg,
        count: existingAgg.count + 1,
        descriptions,
        metricsList,
        windowEndTime: newWindowEnd
      },
      isNewWindow: false
    };
  }

  const id = uuidv4();
  const newAgg = {
    id,
    alertKey,
    name,
    level,
    sourceSystem,
    count: 1,
    descriptions: [description],
    metricsList: [metrics],
    windowEndTime: now + AGGREGATION_WINDOW_MS,
    createdAt: now
  };

  await run(
    `INSERT INTO aggregated_alerts (id, alertKey, name, level, sourceSystem, count, descriptions, metricsList, windowEndTime, createdAt)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      id,
      alertKey,
      name,
      level,
      sourceSystem,
      1,
      JSON.stringify([description]),
      JSON.stringify([metrics]),
      newAgg.windowEndTime,
      now
    ]
  );

  return { aggregated: newAgg, isNewWindow: true };
};

export const getExpiredAggregations = async (): Promise<any[]> => {
  const now = Date.now();
  const rows = await all<any>(
    `SELECT * FROM aggregated_alerts WHERE windowEndTime <= ?`,
    [now]
  );

  return rows.map(row => ({
    ...row,
    descriptions: JSON.parse(row.descriptions),
    metricsList: JSON.parse(row.metricsList)
  }));
};

export const removeAggregation = async (alertKey: string): Promise<void> => {
  await run(`DELETE FROM aggregated_alerts WHERE alertKey = ?`, [alertKey]);
};
