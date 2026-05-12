import { db } from '../database';
import { config, AppRow, SessionRow } from '../config';
import { generateRandomId, hmacSign, nowSeconds, generateToken } from '../utils/auth';

export function getAppById(appId: string): AppRow | undefined {
  return db.prepare('SELECT * FROM apps WHERE id = ?').get(appId) as AppRow | undefined;
}

export function isCallbackUrlRegistered(appId: string, callbackUrl: string): boolean {
  const app = getAppById(appId);
  if (!app) return false;
  return app.callback_url === callbackUrl;
}

export interface SsoAuthPayload {
  sessionId: string;
  userId: number;
  username: string;
  appId: string;
  createdAt: number;
  expiresAt: number;
}

export function createSsoSession(
  userId: number,
  username: string,
  appId: string,
  callbackUrl: string
): { sessionId: string; authToken: string; payload: SsoAuthPayload } {
  const sessionId = generateRandomId(16);
  const now = nowSeconds();
  const expiresAt = now + config.ssoTokenExpiresInSeconds;
  
  const app = getAppById(appId);
  if (!app) throw new Error('App not found');
  
  const jwtToken = generateToken(userId, username);
  
  db.prepare(`
    INSERT INTO sessions (id, user_id, app_id, callback_url, token, expires_at)
    VALUES (?, ?, ?, ?, ?, ?)
  `).run(sessionId, userId, appId, callbackUrl, jwtToken, expiresAt);
  
  const payload: SsoAuthPayload = {
    sessionId,
    userId,
    username,
    appId,
    createdAt: now,
    expiresAt
  };
  
  const data = `${sessionId}|${userId}|${username}|${appId}|${now}|${expiresAt}`;
  const signature = hmacSign(data, app.secret);
  const authToken = `${data}|${signature}`;
  
  return { sessionId, authToken, payload };
}

export function buildCallbackUrl(
  baseUrl: string,
  authToken: string
): string {
  const separator = baseUrl.includes('?') ? '&' : '?';
  return `${baseUrl}${separator}token=${encodeURIComponent(authToken)}`;
}

export async function notifyAppLogout(session: SessionRow): Promise<boolean> {
  if (!session.callback_url) return true;
  
  try {
    const app = getAppById(session.app_id);
    if (!app) return false;
    
    const logoutUrl = new URL(session.callback_url);
    logoutUrl.pathname = logoutUrl.pathname.replace('/callback', '/logout');
    
    const now = nowSeconds();
    const data = `logout|${session.user_id}|${session.id}|${now}`;
    const signature = hmacSign(data, app.secret);
    
    const url = `${logoutUrl.toString()}?sessionId=${encodeURIComponent(session.id)}&userId=${session.user_id}&timestamp=${now}&signature=${encodeURIComponent(signature)}`;
    
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);
    
    try {
      await fetch(url, { method: 'GET', signal: controller.signal });
      clearTimeout(timeoutId);
      return true;
    } catch {
      clearTimeout(timeoutId);
      return false;
    }
  } catch {
    return false;
  }
}

export function getUserSessions(userId: number): SessionRow[] {
  const now = nowSeconds();
  return db.prepare(`
    SELECT * FROM sessions 
    WHERE user_id = ? AND expires_at > ?
    ORDER BY created_at DESC
  `).all(userId, now) as SessionRow[];
}

export function deleteSession(sessionId: string): void {
  db.prepare('DELETE FROM sessions WHERE id = ?').run(sessionId);
}

export function deleteAllUserSessions(userId: number): void {
  db.prepare('DELETE FROM sessions WHERE user_id = ?').run(userId);
}

export async function logoutUserFromAllApps(userId: number): Promise<{ total: number; succeeded: number; failed: number }> {
  const sessions = getUserSessions(userId);
  let succeeded = 0;
  let failed = 0;
  
  for (const session of sessions) {
    const success = await notifyAppLogout(session);
    if (success) {
      succeeded++;
    } else {
      failed++;
    }
    deleteSession(session.id);
  }
  
  deleteAllUserSessions(userId);
  
  return { total: sessions.length, succeeded, failed };
}
