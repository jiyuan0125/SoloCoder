import { db } from '../database';
import { Request } from 'express';

export function getClientIp(req: Request): string {
  const xff = req.headers['x-forwarded-for'];
  if (xff) {
    const ips = Array.isArray(xff) ? xff[0] : xff.split(',')[0];
    return ips.trim();
  }
  return req.ip || req.socket.remoteAddress || 'unknown';
}

export function getUserAgent(req: Request): string {
  return req.headers['user-agent'] || 'unknown';
}

export function logAuth(
  userId: number | null,
  username: string | null,
  action: string,
  success: boolean,
  req: Request
): void {
  const ip = getClientIp(req);
  const userAgent = getUserAgent(req);
  
  db.prepare(`
    INSERT INTO auth_logs (user_id, username, action, success, ip, user_agent)
    VALUES (?, ?, ?, ?, ?, ?)
  `).run(userId, username, action, success ? 1 : 0, ip, userAgent);
}
