import type { Request, Response, NextFunction } from 'express';
import { db } from '../database';
import type { RateLimitConfig } from '../types';

interface WindowRequest {
  id: string;
  key: string;
  timestamp: number;
}

const defaultConfig: RateLimitConfig = {
  perSecond: 10,
  perMinute: 100,
  perHour: 1000
};

function ensureSlidingWindowsTable(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS sliding_windows (
      id TEXT PRIMARY KEY,
      window_key TEXT NOT NULL,
      timestamp INTEGER NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_sliding_windows_key ON sliding_windows(window_key);
    CREATE INDEX IF NOT EXISTS idx_sliding_windows_timestamp ON sliding_windows(timestamp);
  `);
}

function countRequestsInWindow(windowKey: string, windowMs: number): { count: number; earliestTime: number } {
  const now = Date.now();
  const startTime = now - windowMs;
  
  const result = db
    .prepare(
      'SELECT COUNT(*) as count, MIN(timestamp) as earliest FROM sliding_windows WHERE window_key = ? AND timestamp >= ?'
    )
    .get(windowKey, startTime) as { count: number; earliest: number | null };

  return {
    count: result.count || 0,
    earliestTime: result.earliest || now
  };
}

function addRequest(windowKey: string): void {
  const { v4: uuidv4 } = require('uuid');
  const id = uuidv4();
  db.prepare('INSERT INTO sliding_windows (id, window_key, timestamp) VALUES (?, ?, ?)')
    .run(id, windowKey, Date.now());
}

function cleanupOldWindows(): void {
  const cutoff = Date.now() - (60 * 60 * 1000);
  db.prepare('DELETE FROM sliding_windows WHERE timestamp < ?').run(cutoff);
}

function checkRateLimit(
  windowKey: string,
  config: RateLimitConfig
): { allowed: boolean; retryAfter: number } {
  const now = Date.now();
  let retryAfter = 0;

  if (config.perSecond !== undefined) {
    const { count, earliestTime } = countRequestsInWindow(windowKey + ':s', 1000);
    if (count >= config.perSecond) {
      retryAfter = Math.max(retryAfter, Math.ceil((earliestTime + 1000 - now) / 1000));
    }
  }

  if (config.perMinute !== undefined) {
    const { count, earliestTime } = countRequestsInWindow(windowKey + ':m', 60 * 1000);
    if (count >= config.perMinute) {
      retryAfter = Math.max(retryAfter, Math.ceil((earliestTime + 60 * 1000 - now) / 1000));
    }
  }

  if (config.perHour !== undefined) {
    const { count, earliestTime } = countRequestsInWindow(windowKey + ':h', 60 * 60 * 1000);
    if (count >= config.perHour) {
      retryAfter = Math.max(retryAfter, Math.ceil((earliestTime + 60 * 60 * 1000 - now) / 1000));
    }
  }

  return {
    allowed: retryAfter === 0,
    retryAfter
  };
}

function recordRequests(windowKey: string): void {
  addRequest(windowKey + ':s');
  addRequest(windowKey + ':m');
  addRequest(windowKey + ':h');
}

function getClientIp(req: Request): string {
  const xff = req.headers['x-forwarded-for'];
  if (typeof xff === 'string') {
    return xff.split(',')[0].trim();
  }
  return req.ip || req.socket.remoteAddress || 'unknown';
}

function getRateLimitConfig(type: string, identifier: string): RateLimitConfig {
  const result = db
    .prepare('SELECT per_second, per_minute, per_hour FROM rate_limit_configs WHERE type = ? AND identifier = ?')
    .get(type, identifier) as { per_second?: number; per_minute?: number; per_hour?: number } | undefined;

  if (result) {
    return {
      perSecond: result.per_second,
      perMinute: result.per_minute,
      perHour: result.per_hour
    };
  }

  return defaultConfig;
}

export function rateLimiterMiddleware(req: Request, res: Response, next: NextFunction): void {
  if (req.path.startsWith('/gateway/')) {
    return next();
  }

  ensureSlidingWindowsTable();
  
  const clientIp = getClientIp(req);
  req.clientIp = clientIp;
  const userId = req.apiKey?.client_id || 'anonymous';
  const requestPath = req.matchedService?.matchedPrefix || req.path;

  const checks = [
    { key: `ip:${clientIp}`, config: getRateLimitConfig('ip', clientIp) },
    { key: `user:${userId}`, config: getRateLimitConfig('user', userId) },
    { key: `path:${requestPath}`, config: getRateLimitConfig('path', requestPath) }
  ];

  for (const { key, config } of checks) {
    const { allowed, retryAfter } = checkRateLimit(key, config);
    
    if (!allowed) {
      res.status(429).json({
        error: 'Too Many Requests',
        message: 'Rate limit exceeded',
        retryAfter
      });
      return;
    }
  }

  for (const { key } of checks) {
    recordRequests(key);
  }

  if (Math.random() < 0.01) {
    cleanupOldWindows();
  }

  next();
}
