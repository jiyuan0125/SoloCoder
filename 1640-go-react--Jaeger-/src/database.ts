import Database from 'better-sqlite3';
import { Span, Trace, Dependency } from './types';

const db = new Database('./traces.db');

db.exec(`
  CREATE TABLE IF NOT EXISTS spans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    traceId TEXT NOT NULL,
    spanId TEXT NOT NULL,
    parentSpanId TEXT,
    operationName TEXT NOT NULL,
    serviceName TEXT NOT NULL,
    startTime INTEGER NOT NULL,
    duration INTEGER NOT NULL,
    tags TEXT,
    UNIQUE(traceId, spanId)
  );
  
  CREATE INDEX IF NOT EXISTS idx_spans_traceId ON spans(traceId);
  CREATE INDEX IF NOT EXISTS idx_spans_startTime ON spans(startTime);
  CREATE INDEX IF NOT EXISTS idx_spans_serviceName ON spans(serviceName);
`);

function insertSpan(span: Span): void {
  const stmt = db.prepare(`
    INSERT OR REPLACE INTO spans 
    (traceId, spanId, parentSpanId, operationName, serviceName, startTime, duration, tags)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `);
  stmt.run(
    span.traceId,
    span.spanId,
    span.parentSpanId || null,
    span.operationName,
    span.serviceName,
    span.startTime,
    span.duration,
    span.tags ? JSON.stringify(span.tags) : null
  );
}

function insertSpansBatch(spans: Span[]): void {
  const tx = db.transaction((spanList: Span[]) => {
    for (const span of spanList) {
      insertSpan(span);
    }
  });
  tx(spans);
}

function getSpansByTraceId(traceId: string): Span[] {
  const rows = db.prepare('SELECT * FROM spans WHERE traceId = ? ORDER BY startTime ASC').all(traceId);
  return rows.map((row: any) => rowToSpan(row));
}

function getSpanById(traceId: string, spanId: string): Span | null {
  const row = db.prepare('SELECT * FROM spans WHERE traceId = ? AND spanId = ?').get(traceId, spanId) as any;
  return row ? rowToSpan(row) : null;
}

function rowToSpan(row: any): Span {
  return {
    traceId: row.traceId,
    spanId: row.spanId,
    parentSpanId: row.parentSpanId || undefined,
    operationName: row.operationName,
    serviceName: row.serviceName,
    startTime: row.startTime,
    duration: row.duration,
    tags: row.tags ? JSON.parse(row.tags) : undefined,
  };
}

function searchTraces(params: {
  serviceName?: string;
  operationName?: string;
  startTimeMin?: number;
  startTimeMax?: number;
  minDuration?: number;
  page: number;
  pageSize: number;
}): { traces: Trace[]; total: number } {
  const conditions: string[] = [];
  const values: any[] = [];

  if (params.serviceName) {
    conditions.push('traceId IN (SELECT DISTINCT traceId FROM spans WHERE serviceName = ?)');
    values.push(params.serviceName);
  }

  if (params.operationName) {
    conditions.push('traceId IN (SELECT DISTINCT traceId FROM spans WHERE operationName = ?)');
    values.push(params.operationName);
  }

  if (params.startTimeMin) {
    conditions.push('startTime >= ?');
    values.push(params.startTimeMin);
  }

  if (params.startTimeMax) {
    conditions.push('startTime <= ?');
    values.push(params.startTimeMax);
  }

  if (params.minDuration) {
    conditions.push('duration >= ?');
    values.push(params.minDuration);
  }

  const whereClause = conditions.length > 0 ? 'WHERE ' + conditions.join(' AND ') : '';

  const countStmt = db.prepare(`
    SELECT COUNT(DISTINCT traceId) as total FROM spans ${whereClause}
  `);
  const { total } = countStmt.get(...values) as { total: number };

  const offset = (params.page - 1) * params.pageSize;
  const traceIdsStmt = db.prepare(`
    SELECT DISTINCT traceId, MAX(startTime) as maxStartTime 
    FROM spans ${whereClause}
    GROUP BY traceId
    ORDER BY maxStartTime DESC
    LIMIT ? OFFSET ?
  `);
  const traceIdsRows = traceIdsStmt.all(...values, params.pageSize, offset) as any[];

  const traces: Trace[] = [];
  for (const row of traceIdsRows) {
    const spans = getSpansByTraceId(row.traceId);
    if (spans.length > 0) {
      const startTime = Math.min(...spans.map(s => s.startTime));
      const endTime = Math.max(...spans.map(s => s.startTime + s.duration));
      traces.push({
        traceId: row.traceId,
        spans,
        startTime,
        duration: endTime - startTime,
      });
    }
  }

  return { traces, total };
}

function getDependencies(): Dependency[] {
  const rows = db.prepare(`
    SELECT 
      child.serviceName as fromService,
      parent.serviceName as toService,
      COUNT(*) as callCount,
      AVG(child.duration) as avgDuration
    FROM spans child
    JOIN spans parent ON child.parentSpanId = parent.spanId AND child.traceId = parent.traceId
    WHERE child.parentSpanId IS NOT NULL
    GROUP BY child.serviceName, parent.serviceName
  `).all() as any[];

  return rows.map(row => ({
    from: row.fromService,
    to: row.toService,
    callCount: row.callCount,
    avgDuration: Math.round(row.avgDuration),
  }));
}

function cleanupOldData(daysAgo: number = 7): void {
  const cutoffTime = Date.now() - daysAgo * 24 * 60 * 60 * 1000;
  const stmt = db.prepare('DELETE FROM spans WHERE startTime < ?');
  stmt.run(cutoffTime);
}

export {
  insertSpan,
  insertSpansBatch,
  getSpansByTraceId,
  getSpanById,
  searchTraces,
  getDependencies,
  cleanupOldData,
};
