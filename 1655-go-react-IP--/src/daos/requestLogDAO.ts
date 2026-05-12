import db from '../database';
import { RequestLog } from '../types';
import { v4 as uuidv4 } from 'uuid';

interface LogRow {
  id: string;
  ip: string;
  path: string;
  method: string;
  timestamp: number;
}

function mapLog(row: LogRow): RequestLog {
  return {
    id: row.id,
    ip: row.ip,
    path: row.path,
    method: row.method,
    timestamp: row.timestamp
  };
}

export const requestLogDAO = {
  logRequest(ip: string, path: string, method: string): RequestLog {
    const id = uuidv4();
    const timestamp = Date.now();
    
    db.prepare(`
      INSERT INTO request_logs (id, ip, path, method, timestamp)
      VALUES (?, ?, ?, ?, ?)
    `).run(id, ip, path, method, timestamp);

    return { id, ip, path, method, timestamp };
  },

  getRequestsInWindow(ip: string, startTime: number): number {
    const result = db.prepare(`
      SELECT COUNT(*) as count 
      FROM request_logs 
      WHERE ip = ? AND timestamp >= ?
    `).get(ip, startTime) as { count: number };
    
    return result.count;
  },

  cleanupOldLogs(cutoffTime: number): number {
    const result = db.prepare('DELETE FROM request_logs WHERE timestamp < ?').run(cutoffTime);
    return result.changes;
  },

  getLatestRequest(ip: string): RequestLog | null {
    const row = db.prepare(`
      SELECT * FROM request_logs 
      WHERE ip = ? 
      ORDER BY timestamp DESC 
      LIMIT 1
    `).get(ip) as LogRow | undefined;
    
    return row ? mapLog(row) : null;
  }
};
