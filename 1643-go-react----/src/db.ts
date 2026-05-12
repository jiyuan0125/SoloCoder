import Database from 'better-sqlite3';
import { ProfilingRecord, ProfilingSample, StackFrame, ProfilingType } from './types';

const db = new Database('profiling.db');

db.exec(`
  CREATE TABLE IF NOT EXISTS profiling_records (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL,
    type TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    duration INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    created_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS profiling_samples (
    id TEXT PRIMARY KEY,
    record_id TEXT NOT NULL,
    stack_id TEXT NOT NULL,
    sample_count INTEGER NOT NULL,
    self_time INTEGER NOT NULL,
    FOREIGN KEY (record_id) REFERENCES profiling_records(id) ON DELETE CASCADE
  );

  CREATE TABLE IF NOT EXISTS stack_frames (
    id TEXT PRIMARY KEY,
    stack_id TEXT NOT NULL,
    depth INTEGER NOT NULL,
    function_name TEXT NOT NULL,
    file_name TEXT NOT NULL,
    line_number INTEGER NOT NULL
  );

  CREATE INDEX IF NOT EXISTS idx_records_app_id ON profiling_records(app_id);
  CREATE INDEX IF NOT EXISTS idx_records_status ON profiling_records(status);
  CREATE INDEX IF NOT EXISTS idx_samples_record_id ON profiling_samples(record_id);
  CREATE INDEX IF NOT EXISTS idx_frames_stack_id ON stack_frames(stack_id);
`);

export function getRunningRecord(appId: string): ProfilingRecord | undefined {
  const row = db.prepare(`
    SELECT id, app_id as appId, type, start_time as startTime, duration, status
    FROM profiling_records
    WHERE app_id = ? AND status = 'running'
    LIMIT 1
  `).get(appId) as any;
  return row;
}

export function insertRecord(record: ProfilingRecord): void {
  db.prepare(`
    INSERT INTO profiling_records (id, app_id, type, start_time, duration, status, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `).run(record.id, record.appId, record.type, record.startTime, record.duration, record.status, Date.now());
}

export function getRecordById(id: string): ProfilingRecord | undefined {
  const row = db.prepare(`
    SELECT id, app_id as appId, type, start_time as startTime, duration, status
    FROM profiling_records
    WHERE id = ?
  `).get(id) as any;
  return row;
}

export function updateRecordStatus(id: string, duration: number): void {
  db.prepare(`
    UPDATE profiling_records
    SET status = 'completed', duration = ?
    WHERE id = ?
  `).run(duration, id);
}

export function getRecordsByApp(appId: string): ProfilingRecord[] {
  const rows = db.prepare(`
    SELECT id, app_id as appId, type, start_time as startTime, duration, status
    FROM profiling_records
    WHERE app_id = ?
    ORDER BY created_at DESC
  `).all(appId) as any[];
  return rows;
}

export function getRecordCount(appId: string): number {
  const row = db.prepare(`
    SELECT COUNT(*) as count FROM profiling_records WHERE app_id = ?
  `).get(appId) as any;
  return row.count;
}

export function deleteOldestRecords(appId: string, count: number): void {
  const rows = db.prepare(`
    SELECT id FROM profiling_records
    WHERE app_id = ?
    ORDER BY created_at ASC
    LIMIT ?
  `).all(appId, count) as any[];
  
  for (const row of rows) {
    deleteRecord(row.id);
  }
}

export function deleteRecord(id: string): void {
  db.prepare('DELETE FROM profiling_samples WHERE record_id = ?').run(id);
  db.prepare('DELETE FROM profiling_records WHERE id = ?').run(id);
}

export function insertSample(sample: ProfilingSample): void {
  db.prepare(`
    INSERT INTO profiling_samples (id, record_id, stack_id, sample_count, self_time)
    VALUES (?, ?, ?, ?, ?)
  `).run(sample.id, sample.recordId, sample.stackId, sample.sampleCount, sample.selfTime);
}

export function insertStackFrame(frame: StackFrame): void {
  db.prepare(`
    INSERT INTO stack_frames (id, stack_id, depth, function_name, file_name, line_number)
    VALUES (?, ?, ?, ?, ?, ?)
  `).run(frame.id, frame.stackId, frame.depth, frame.functionName, frame.fileName, frame.lineNumber);
}

interface SampleWithFrame {
  sample_count: number;
  self_time: number;
  depth: number;
  function_name: string;
  file_name: string;
  line_number: number;
}

export function getSamplesWithFrames(recordId: string): SampleWithFrame[] {
  return db.prepare(`
    SELECT s.sample_count, s.self_time, f.depth, f.function_name, f.file_name, f.line_number
    FROM profiling_samples s
    JOIN stack_frames f ON s.stack_id = f.stack_id
    WHERE s.record_id = ?
    ORDER BY s.id, f.depth
  `).all(recordId) as SampleWithFrame[];
}
