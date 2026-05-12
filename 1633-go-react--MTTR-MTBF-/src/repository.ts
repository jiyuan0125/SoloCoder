import { v4 as uuidv4 } from 'uuid';
import { runQuery, runExec } from './db';
import { Incident, CreateIncidentRequest, UpdateIncidentRequest, ListIncidentsQuery, IncidentStatus } from './types';

interface IncidentRow {
  id: string;
  incident_number: string;
  service: string;
  start_time: string;
  discovered_time: string;
  recovered_time: string | null;
  severity: string;
  status: string;
  impact_scope: string;
  created_at: string;
  updated_at: string;
}

function rowToIncident(row: IncidentRow): Incident {
  return {
    id: row.id,
    incidentNumber: row.incident_number,
    service: row.service,
    startTime: new Date(row.start_time),
    discoveredTime: new Date(row.discovered_time),
    recoveredTime: row.recovered_time ? new Date(row.recovered_time) : null,
    severity: row.severity as any,
    status: row.status as any,
    impactScope: row.impact_scope as any,
    createdAt: new Date(row.created_at),
    updatedAt: new Date(row.updated_at),
  };
}

export async function createIncident(
  data: CreateIncidentRequest,
  severity: string,
  status: string
): Promise<Incident> {
  const id = uuidv4();
  const now = new Date().toISOString();

  await runExec(
    `INSERT INTO incidents (
      id, incident_number, service, start_time, discovered_time, recovered_time,
      severity, status, impact_scope, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [
      id,
      data.incidentNumber,
      data.service,
      data.startTime,
      data.discoveredTime,
      null,
      severity,
      status,
      data.impactScope,
      now,
      now,
    ]
  );

  const rows = await runQuery<IncidentRow>('SELECT * FROM incidents WHERE id = ?', [id]);
  return rowToIncident(rows[0]);
}

export async function findIncidentById(id: string): Promise<Incident | null> {
  const rows = await runQuery<IncidentRow>('SELECT * FROM incidents WHERE id = ?', [id]);
  return rows.length > 0 ? rowToIncident(rows[0]) : null;
}

export async function listIncidents(query: ListIncidentsQuery): Promise<Incident[]> {
  const conditions: string[] = [];
  const params: any[] = [];

  if (query.service) {
    conditions.push('service = ?');
    params.push(query.service);
  }

  if (query.startTimeFrom) {
    conditions.push('start_time >= ?');
    params.push(query.startTimeFrom);
  }

  if (query.startTimeTo) {
    conditions.push('start_time <= ?');
    params.push(query.startTimeTo);
  }

  let sql = 'SELECT * FROM incidents';
  if (conditions.length > 0) {
    sql += ' WHERE ' + conditions.join(' AND ');
  }
  sql += ' ORDER BY start_time DESC';

  const rows = await runQuery<IncidentRow>(sql, params);
  return rows.map(rowToIncident);
}

export async function updateIncident(
  id: string,
  data: UpdateIncidentRequest,
  severity?: string
): Promise<Incident | null> {
  const fields: string[] = [];
  const params: any[] = [];

  if (data.service !== undefined) {
    fields.push('service = ?');
    params.push(data.service);
  }

  if (data.startTime !== undefined) {
    fields.push('start_time = ?');
    params.push(data.startTime);
  }

  if (data.discoveredTime !== undefined) {
    fields.push('discovered_time = ?');
    params.push(data.discoveredTime);
  }

  if (data.recoveredTime !== undefined) {
    fields.push('recovered_time = ?');
    params.push(data.recoveredTime);
  }

  if (data.impactScope !== undefined) {
    fields.push('impact_scope = ?');
    params.push(data.impactScope);
  }

  if (data.severity !== undefined) {
    fields.push('severity = ?');
    params.push(data.severity);
  } else if (severity !== undefined) {
    fields.push('severity = ?');
    params.push(severity);
  }

  fields.push('updated_at = ?');
  params.push(new Date().toISOString());

  if (fields.length === 0) {
    return findIncidentById(id);
  }

  params.push(id);
  const sql = `UPDATE incidents SET ${fields.join(', ')} WHERE id = ?`;
  const result = await runExec(sql, params);

  if (result.changes === 0) {
    return null;
  }

  return findIncidentById(id);
}

export async function updateIncidentStatus(
  id: string,
  status: IncidentStatus,
  recoveredTime?: string
): Promise<Incident | null> {
  const fields: string[] = ['status = ?', 'updated_at = ?'];
  const params: any[] = [status, new Date().toISOString()];

  if (recoveredTime !== undefined) {
    fields.push('recovered_time = ?');
    params.push(recoveredTime);
  }

  params.push(id);
  const sql = `UPDATE incidents SET ${fields.join(', ')} WHERE id = ?`;
  const result = await runExec(sql, params);

  if (result.changes === 0) {
    return null;
  }

  return findIncidentById(id);
}
