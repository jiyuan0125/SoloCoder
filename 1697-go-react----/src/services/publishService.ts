import db from '../database';
import { ReportStatus, PublishedRecord } from '../types';
import { generateId, formatDate } from '../utils';

interface PublishInput {
  reportId: string;
  publisherId: string;
  publisherName: string;
  content: string;
}

interface UpdatePublishedInput {
  reportId: string;
  publisherId: string;
  publisherName: string;
  content: string;
}

export function publishReport(input: PublishInput): {
  success: boolean;
  notFound?: boolean;
  notVerified?: boolean;
  record?: PublishedRecord;
} {
  const report = db.prepare('SELECT * FROM disaster_reports WHERE id = ?').get(input.reportId) as any;
  
  if (!report) {
    return { success: false, notFound: true };
  }

  if (report.status !== ReportStatus.VERIFIED) {
    return { success: false, notVerified: true };
  }

  const id = generateId();
  const now = formatDate(new Date());

  db.prepare(`
    INSERT INTO published_records (
      id, report_id, publisher_id, publisher_name, content,
      version, previous_content, published_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `).run(id, input.reportId, input.publisherId, input.publisherName, input.content, 1, null, now);

  db.prepare(`
    UPDATE disaster_reports SET status = ?, updated_at = ? WHERE id = ?
  `).run(ReportStatus.PUBLISHED, now, input.reportId);

  const record = db.prepare('SELECT * FROM published_records WHERE id = ?').get(id) as any;
  
  return {
    success: true,
    record: mapToPublishedRecord(record)
  };
}

export function updatePublished(input: UpdatePublishedInput): {
  success: boolean;
  notFound?: boolean;
  limitExceeded?: boolean;
  record?: PublishedRecord;
} {
  const existingRecords = db.prepare(`
    SELECT * FROM published_records WHERE report_id = ? ORDER BY version DESC
  `).all(input.reportId) as any[];

  if (existingRecords.length === 0) {
    return { success: false, notFound: true };
  }

  if (existingRecords.length >= 3) {
    return { success: false, limitExceeded: true };
  }

  const latest = existingRecords[0];
  const newVersion = latest.version + 1;
  const id = generateId();
  const now = formatDate(new Date());

  db.prepare(`
    INSERT INTO published_records (
      id, report_id, publisher_id, publisher_name, content,
      version, previous_content, published_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `).run(id, input.reportId, input.publisherId, input.publisherName, input.content, newVersion, latest.content, now);

  const record = db.prepare('SELECT * FROM published_records WHERE id = ?').get(id) as any;
  
  return {
    success: true,
    record: mapToPublishedRecord(record)
  };
}

export function getPublishedRecords(reportId: string): PublishedRecord[] {
  const rows = db.prepare(`
    SELECT * FROM published_records WHERE report_id = ? ORDER BY version DESC
  `).all(reportId) as any[];
  
  return rows.map(mapToPublishedRecord);
}

export function getLatestPublished(reportId: string): PublishedRecord | null {
  const row = db.prepare(`
    SELECT * FROM published_records WHERE report_id = ? ORDER BY version DESC LIMIT 1
  `).get(reportId) as any;
  
  return row ? mapToPublishedRecord(row) : null;
}

function mapToPublishedRecord(row: any): PublishedRecord {
  return {
    id: row.id,
    reportId: row.report_id,
    publisherId: row.publisher_id,
    publisherName: row.publisher_name,
    content: row.content,
    version: row.version,
    previousContent: row.previous_content,
    publishedAt: row.published_at
  };
}
