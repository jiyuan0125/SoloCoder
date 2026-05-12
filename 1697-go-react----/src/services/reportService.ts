import db from '../database';
import { ReportStatus, DisasterReport } from '../types';
import { generateId, getMaxValue, isSameTimeRange, formatDate, parseNumeric } from '../utils';

interface ReportInput {
  disasterType: string;
  occurrenceTime: string;
  location: string;
  affectedPopulation?: string | number;
  evacuatedPopulation?: string | number;
  deathMissingCount?: string | number;
  cropAreaAffected?: string | number;
  housesDamaged?: string | number;
  directEconomicLoss?: string | number;
}

function findSimilarReport(input: ReportInput): any | null {
  const reports = db.prepare(`
    SELECT * FROM disaster_reports
    WHERE location = ? AND status IN ('待核查', '需补充')
  `).all(input.location) as any[];

  for (const report of reports) {
    if (report.disaster_type === input.disasterType &&
        isSameTimeRange(report.occurrence_time, input.occurrenceTime)) {
      return report;
    }
  }
  return null;
}

function mergeWithMax(existing: any, input: ReportInput): any {
  return {
    ...existing,
    affectedPopulation: getMaxValue(existing.affected_population, input.affectedPopulation ?? '待核实'),
    evacuatedPopulation: getMaxValue(existing.evacuated_population, input.evacuatedPopulation ?? '待核实'),
    deathMissingCount: getMaxValue(existing.death_missing_count, input.deathMissingCount ?? '待核实'),
    cropAreaAffected: getMaxValue(existing.crop_area_affected, input.cropAreaAffected ?? '待核实'),
    housesDamaged: getMaxValue(existing.houses_damaged, input.housesDamaged ?? '待核实'),
    directEconomicLoss: getMaxValue(existing.direct_economic_loss, input.directEconomicLoss ?? '待核实'),
    reportCount: (existing.report_count || 1) + 1
  };
}

export function createReport(input: ReportInput): { report: DisasterReport; merged: boolean } {
  const similarReport = findSimilarReport(input);

  if (similarReport) {
    const merged = mergeWithMax(similarReport, input);
    const now = formatDate(new Date());
    
    db.prepare(`
      UPDATE disaster_reports
      SET affected_population = ?,
          evacuated_population = ?,
          death_missing_count = ?,
          crop_area_affected = ?,
          houses_damaged = ?,
          direct_economic_loss = ?,
          report_count = ?,
          updated_at = ?,
          status = '待核查'
      WHERE id = ?
    `).run(
      String(merged.affectedPopulation),
      String(merged.evacuatedPopulation),
      String(merged.deathMissingCount),
      String(merged.cropAreaAffected),
      String(merged.housesDamaged),
      String(merged.directEconomicLoss),
      merged.reportCount,
      now,
      similarReport.id
    );

    const updatedReport = db.prepare('SELECT * FROM disaster_reports WHERE id = ?').get(similarReport.id) as any;
    return { report: mapToReport(updatedReport), merged: true };
  }

  const id = generateId();
  const now = formatDate(new Date());

  db.prepare(`
    INSERT INTO disaster_reports (
      id, disaster_type, occurrence_time, location,
      affected_population, evacuated_population, death_missing_count,
      crop_area_affected, houses_damaged, direct_economic_loss,
      status, created_at, updated_at, report_count
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    id,
    input.disasterType,
    input.occurrenceTime,
    input.location,
    String(input.affectedPopulation ?? '待核实'),
    String(input.evacuatedPopulation ?? '待核实'),
    String(input.deathMissingCount ?? '待核实'),
    String(input.cropAreaAffected ?? '待核实'),
    String(input.housesDamaged ?? '待核实'),
    String(input.directEconomicLoss ?? '待核实'),
    ReportStatus.PENDING,
    now,
    now,
    1
  );

  const newReport = db.prepare('SELECT * FROM disaster_reports WHERE id = ?').get(id) as any;
  return { report: mapToReport(newReport), merged: false };
}

export function getReportById(id: string): DisasterReport | null {
  const row = db.prepare('SELECT * FROM disaster_reports WHERE id = ?').get(id) as any;
  return row ? mapToReport(row) : null;
}

export function getAllReports(status?: string): DisasterReport[] {
  let query = 'SELECT * FROM disaster_reports';
  const params: any[] = [];
  
  if (status) {
    query += ' WHERE status = ?';
    params.push(status);
  }
  
  query += ' ORDER BY created_at DESC';
  
  const rows = db.prepare(query).all(...params) as any[];
  return rows.map(mapToReport);
}

export function updateReport(id: string, updates: Partial<ReportInput>, requiresApproval: boolean = false): { success: boolean; needsApproval?: boolean } {
  const report = db.prepare('SELECT * FROM disaster_reports WHERE id = ?').get(id) as any;
  
  if (!report) {
    return { success: false };
  }

  const isVerified = report.status === ReportStatus.VERIFIED || report.status === ReportStatus.PUBLISHED;
  
  if (isVerified && requiresApproval) {
    return { success: false, needsApproval: true };
  }

  const now = formatDate(new Date());
  const fields: string[] = [];
  const values: any[] = [];

  if (updates.affectedPopulation !== undefined) {
    fields.push('affected_population = ?');
    values.push(String(updates.affectedPopulation));
  }
  if (updates.evacuatedPopulation !== undefined) {
    fields.push('evacuated_population = ?');
    values.push(String(updates.evacuatedPopulation));
  }
  if (updates.deathMissingCount !== undefined) {
    fields.push('death_missing_count = ?');
    values.push(String(updates.deathMissingCount));
  }
  if (updates.cropAreaAffected !== undefined) {
    fields.push('crop_area_affected = ?');
    values.push(String(updates.cropAreaAffected));
  }
  if (updates.housesDamaged !== undefined) {
    fields.push('houses_damaged = ?');
    values.push(String(updates.housesDamaged));
  }
  if (updates.directEconomicLoss !== undefined) {
    fields.push('direct_economic_loss = ?');
    values.push(String(updates.directEconomicLoss));
  }

  if (fields.length === 0) {
    return { success: true };
  }

  fields.push('updated_at = ?');
  values.push(now);
  
  if (report.status === ReportStatus.NEEDS_SUPPLEMENT) {
    fields.push('status = ?');
    values.push(ReportStatus.PENDING);
  }

  values.push(id);

  db.prepare(`UPDATE disaster_reports SET ${fields.join(', ')} WHERE id = ?`).run(...values);
  return { success: true };
}

export function setReportStatus(id: string, status: ReportStatus): boolean {
  const now = formatDate(new Date());
  const result = db.prepare(`
    UPDATE disaster_reports SET status = ?, updated_at = ? WHERE id = ?
  `).run(status, now, id);
  
  return result.changes > 0;
}

function mapToReport(row: any): DisasterReport {
  return {
    id: row.id,
    disasterType: row.disaster_type,
    occurrenceTime: row.occurrence_time,
    location: row.location,
    affectedPopulation: parseNumeric(row.affected_population),
    evacuatedPopulation: parseNumeric(row.evacuated_population),
    deathMissingCount: parseNumeric(row.death_missing_count),
    cropAreaAffected: parseNumeric(row.crop_area_affected),
    housesDamaged: parseNumeric(row.houses_damaged),
    directEconomicLoss: parseNumeric(row.direct_economic_loss),
    status: row.status as ReportStatus,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
    mergedFrom: row.merged_from,
    reportCount: row.report_count
  };
}
