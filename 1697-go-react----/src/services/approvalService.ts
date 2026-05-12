import db from '../database';
import { generateId, formatDate } from '../utils';

interface CreateRequestInput {
  reportId: string;
  requesterId: string;
  requesterName: string;
  changes: Record<string, any>;
  reason?: string;
}

interface ApproveRequestInput {
  requestId: string;
  approverId: string;
}

export function createModificationRequest(input: CreateRequestInput): {
  success: boolean;
  notFound?: boolean;
  requestId?: string;
} {
  const report = db.prepare('SELECT * FROM disaster_reports WHERE id = ?').get(input.reportId) as any;
  
  if (!report) {
    return { success: false, notFound: true };
  }

  const id = generateId();
  const now = formatDate(new Date());

  db.prepare(`
    INSERT INTO modification_requests (
      id, report_id, requester_id, requester_name, changes,
      reason, status, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    id,
    input.reportId,
    input.requesterId,
    input.requesterName,
    JSON.stringify(input.changes),
    input.reason || '',
    '待审批',
    now
  );

  return {
    success: true,
    requestId: id
  };
}

export function approveRequest(input: ApproveRequestInput): {
  success: boolean;
  notFound?: boolean;
} {
  const request = db.prepare('SELECT * FROM modification_requests WHERE id = ?').get(input.requestId) as any;
  
  if (!request) {
    return { success: false, notFound: true };
  }

  const now = formatDate(new Date());

  db.prepare(`
    UPDATE modification_requests 
    SET status = ?, approved_by = ?, approved_at = ?
    WHERE id = ?
  `).run('已通过', input.approverId, now, input.requestId);

  const changes = JSON.parse(request.changes);
  const reportId = request.report_id;
  const now2 = formatDate(new Date());

  const fields: string[] = [];
  const values: any[] = [];

  if (changes.affectedPopulation !== undefined) {
    fields.push('affected_population = ?');
    values.push(String(changes.affectedPopulation));
  }
  if (changes.evacuatedPopulation !== undefined) {
    fields.push('evacuated_population = ?');
    values.push(String(changes.evacuatedPopulation));
  }
  if (changes.deathMissingCount !== undefined) {
    fields.push('death_missing_count = ?');
    values.push(String(changes.deathMissingCount));
  }
  if (changes.cropAreaAffected !== undefined) {
    fields.push('crop_area_affected = ?');
    values.push(String(changes.cropAreaAffected));
  }
  if (changes.housesDamaged !== undefined) {
    fields.push('houses_damaged = ?');
    values.push(String(changes.housesDamaged));
  }
  if (changes.directEconomicLoss !== undefined) {
    fields.push('direct_economic_loss = ?');
    values.push(String(changes.directEconomicLoss));
  }

  if (fields.length > 0) {
    fields.push('updated_at = ?');
    values.push(now2);
    values.push(reportId);

    db.prepare(`UPDATE disaster_reports SET ${fields.join(', ')} WHERE id = ?`).run(...values);
  }

  return { success: true };
}

export function rejectRequest(requestId: string, approverId: string): {
  success: boolean;
  notFound?: boolean;
} {
  const request = db.prepare('SELECT * FROM modification_requests WHERE id = ?').get(requestId) as any;
  
  if (!request) {
    return { success: false, notFound: true };
  }

  const now = formatDate(new Date());

  db.prepare(`
    UPDATE modification_requests 
    SET status = ?, approved_by = ?, approved_at = ?
    WHERE id = ?
  `).run('已拒绝', approverId, now, requestId);

  return { success: true };
}

export function getPendingRequests(): any[] {
  const rows = db.prepare(`
    SELECT * FROM modification_requests WHERE status = '待审批' ORDER BY created_at DESC
  `).all() as any[];
  
  return rows.map(row => ({
    ...row,
    changes: JSON.parse(row.changes)
  }));
}
