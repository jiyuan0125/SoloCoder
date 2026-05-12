import db from '../database';
import { ReportStatus, VerificationResult, VerificationRecord } from '../types';
import { generateId, formatDate, isMajorDisaster, getVerificationDeadline, parseNumeric } from '../utils';

interface StartVerificationInput {
  reportId: string;
  verifierId: string;
  verifierName: string;
}

interface CompleteVerificationInput {
  reportId: string;
  verifierId: string;
  result: VerificationResult;
  comments?: string;
}

export function startVerification(input: StartVerificationInput): {
  success: boolean;
  conflict?: boolean;
  notFound?: boolean;
  record?: VerificationRecord;
} {
  const report = db.prepare('SELECT * FROM disaster_reports WHERE id = ?').get(input.reportId) as any;
  
  if (!report) {
    return { success: false, notFound: true };
  }

  const existingVerification = db.prepare(`
    SELECT * FROM verification_records WHERE report_id = ? ORDER BY verified_at DESC LIMIT 1
  `).get(input.reportId) as any;

  if (existingVerification && !existingVerification.result) {
    return { success: false, conflict: true };
  }

  if (report.status === ReportStatus.UNDER_REVIEW) {
    return { success: false, conflict: true };
  }

  if (report.status !== ReportStatus.PENDING && report.status !== ReportStatus.NEEDS_SUPPLEMENT) {
    return { success: false };
  }

  const death = parseNumeric(report.death_missing_count);
  const loss = parseNumeric(report.direct_economic_loss);
  const isMajor = isMajorDisaster(death, loss);
  const deadline = getVerificationDeadline(isMajor);

  const id = generateId();
  const now = formatDate(new Date());

  db.prepare(`
    INSERT INTO verification_records (
      id, report_id, verifier_id, verifier_name, result,
      comments, verified_at, deadline
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `).run(id, input.reportId, input.verifierId, input.verifierName, '', '', now, deadline);

  db.prepare(`
    UPDATE disaster_reports SET status = ?, updated_at = ? WHERE id = ?
  `).run(ReportStatus.UNDER_REVIEW, now, input.reportId);

  const record = db.prepare('SELECT * FROM verification_records WHERE id = ?').get(id) as any;
  
  return {
    success: true,
    record: mapToVerificationRecord(record)
  };
}

export function completeVerification(input: CompleteVerificationInput): {
  success: boolean;
  notFound?: boolean;
  conflict?: boolean;
} {
  const report = db.prepare('SELECT * FROM disaster_reports WHERE id = ?').get(input.reportId) as any;
  
  if (!report) {
    return { success: false, notFound: true };
  }

  const pendingVerification = db.prepare(`
    SELECT * FROM verification_records 
    WHERE report_id = ? AND (result IS NULL OR result = '')
    ORDER BY verified_at DESC LIMIT 1
  `).get(input.reportId) as any;

  if (!pendingVerification) {
    return { success: false, conflict: true };
  }

  const now = formatDate(new Date());

  db.prepare(`
    UPDATE verification_records 
    SET result = ?, comments = ?, verified_at = ?
    WHERE id = ?
  `).run(input.result, input.comments || '', now, pendingVerification.id);

  const newStatus = input.result === VerificationResult.PASSED
    ? ReportStatus.VERIFIED
    : ReportStatus.NEEDS_SUPPLEMENT;

  db.prepare(`
    UPDATE disaster_reports SET status = ?, updated_at = ? WHERE id = ?
  `).run(newStatus, now, input.reportId);

  return { success: true };
}

export function getVerificationRecords(reportId: string): VerificationRecord[] {
  const rows = db.prepare(`
    SELECT * FROM verification_records WHERE report_id = ? ORDER BY verified_at DESC
  `).all(reportId) as any[];
  
  return rows.map(mapToVerificationRecord);
}

function mapToVerificationRecord(row: any): VerificationRecord {
  return {
    id: row.id,
    reportId: row.report_id,
    verifierId: row.verifier_id,
    verifierName: row.verifier_name,
    result: (row.result || '') as VerificationResult,
    comments: row.comments || '',
    verifiedAt: row.verified_at,
    deadline: row.deadline
  };
}
