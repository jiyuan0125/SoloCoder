import Database from 'better-sqlite3';
import path from 'path';
import { AuditRecord, AuditResult, AuditStatus, Content, ContentType, SensitiveWordMatch } from '../types';

const DB_PATH = path.join(process.cwd(), 'audit.db');

const db = new Database(DB_PATH);
db.pragma('journal_mode = WAL');

db.exec(`
  CREATE TABLE IF NOT EXISTS contents (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS audit_records (
    id TEXT PRIMARY KEY,
    content_id TEXT NOT NULL,
    status TEXT NOT NULL,
    matches TEXT,
    created_at INTEGER NOT NULL,
    decided_at INTEGER,
    decision TEXT,
    FOREIGN KEY (content_id) REFERENCES contents(id)
  );

  CREATE INDEX IF NOT EXISTS idx_audit_records_content_id ON audit_records(content_id);
  CREATE INDEX IF NOT EXISTS idx_audit_records_status ON audit_records(status);
`);

export function createContent(content: {
  id: string;
  type: ContentType;
  content: string;
  status: AuditStatus;
}): void {
  const stmt = db.prepare(`
    INSERT INTO contents (id, type, content, status, created_at)
    VALUES (?, ?, ?, ?, ?)
  `);
  stmt.run(content.id, content.type, content.content, content.status, Date.now());
}

export function getContentById(id: string): Content | null {
  const stmt = db.prepare(`SELECT * FROM contents WHERE id = ?`);
  const row = stmt.get(id) as {
    id: string;
    type: string;
    content: string;
    status: string;
    created_at: number;
  } | null;

  if (!row) return null;

  return {
    id: row.id,
    type: row.type as ContentType,
    content: row.content,
    status: row.status as AuditStatus,
    createdAt: row.created_at
  };
}

export function updateContentStatus(id: string, status: AuditStatus): void {
  const stmt = db.prepare(`UPDATE contents SET status = ? WHERE id = ?`);
  stmt.run(status, id);
}

export function createAuditRecord(record: {
  id: string;
  contentId: string;
  status: AuditStatus;
  matches?: SensitiveWordMatch[];
}): void {
  const stmt = db.prepare(`
    INSERT INTO audit_records (id, content_id, status, matches, created_at, decided_at, decision)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `);
  stmt.run(
    record.id,
    record.contentId,
    record.status,
    record.matches ? JSON.stringify(record.matches) : null,
    Date.now(),
    null,
    null
  );
}

export function getAuditRecordById(id: string): AuditRecord | null {
  const stmt = db.prepare(`SELECT * FROM audit_records WHERE id = ?`);
  const row = stmt.get(id) as {
    id: string;
    content_id: string;
    status: string;
    matches: string | null;
    created_at: number;
    decided_at: number | null;
    decision: string | null;
  } | null;

  if (!row) return null;

  const matches = row.matches ? JSON.parse(row.matches) as SensitiveWordMatch[] : undefined;

  return {
    id: row.id,
    contentId: row.content_id,
    status: row.status as AuditStatus,
    matches,
    createdAt: row.created_at,
    decidedAt: row.decided_at ?? undefined,
    decision: row.decision ? row.decision as AuditResult : undefined
  };
}

export function updateAuditRecordDecision(
  id: string,
  status: AuditStatus,
  decision: AuditResult
): void {
  const stmt = db.prepare(`
    UPDATE audit_records
    SET status = ?, decided_at = ?, decision = ?
    WHERE id = ?
  `);
  stmt.run(status, Date.now(), decision, id);
}

export function getQueueItems(): AuditRecord[] {
  const stmt = db.prepare(`
    SELECT * FROM audit_records
    WHERE status IN ('pending', 'queue')
    ORDER BY created_at ASC
  `);
  const rows = stmt.all() as {
    id: string;
    content_id: string;
    status: string;
    matches: string | null;
    created_at: number;
    decided_at: number | null;
    decision: string | null;
  }[];

  return rows.map(row => ({
    id: row.id,
    contentId: row.content_id,
    status: row.status as AuditStatus,
    matches: row.matches ? JSON.parse(row.matches) as SensitiveWordMatch[] : undefined,
    createdAt: row.created_at,
    decidedAt: row.decided_at ?? undefined,
    decision: row.decision ? row.decision as AuditResult : undefined
  }));
}

export { db };
