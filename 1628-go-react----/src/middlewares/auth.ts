import type { Request, Response, NextFunction } from 'express';
import { createHash, randomBytes } from 'crypto';
import { v4 as uuidv4 } from 'uuid';
import { db } from '../database';
import { getAllRoutePrefixes } from '../services/serviceStore';
import { normalizePrefix } from './routeMatcher';
import type { ApiKey, Client, KeyStats } from '../types';

export function hashApiKey(key: string): string {
  return createHash('sha256').update(key).digest('hex');
}

export function generateApiKey(): string {
  return 'sk_' + randomBytes(32).toString('hex');
}

export function parsePermissions(permissionsStr: string | null): string[] {
  if (!permissionsStr) return [];
  try {
    const parsed = JSON.parse(permissionsStr);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function findApiKey(keyHash: string): ApiKey | undefined {
  const row = db
    .prepare('SELECT * FROM api_keys WHERE key_hash = ? AND is_active = 1')
    .get(keyHash) as 
    | (ApiKey & { permissions: string | null })
    | undefined;

  if (!row) return undefined;

  return {
    ...row,
    permissions: parsePermissions(row.permissions)
  };
}

function checkKeyPermissions(apiKey: ApiKey, matchedPrefix: string | undefined): boolean {
  if (apiKey.permissions.length === 0) {
    return true;
  }

  if (!matchedPrefix) {
    return false;
  }

  const registeredPrefixes = getAllRoutePrefixes().map(s => normalizePrefix(s.route_prefix));
  const normalizedTarget = normalizePrefix(matchedPrefix);

  for (const allowedPrefix of apiKey.permissions) {
    const normalizedAllowed = normalizePrefix(allowedPrefix);

    if (!registeredPrefixes.includes(normalizedAllowed)) {
      continue;
    }

    if (normalizedAllowed === normalizedTarget) {
      return true;
    }

    if (normalizedTarget.startsWith(normalizedAllowed + '/')) {
      return true;
    }
  }

  return false;
}

export function updateKeyStats(keyId: string, success: boolean): void {
  const now = Date.now();
  
  const existing = db
    .prepare('SELECT * FROM key_stats WHERE key_id = ?')
    .get(keyId) as KeyStats | undefined;

  if (!existing) {
    const id = uuidv4();
    db.prepare(`
      INSERT INTO key_stats (id, key_id, total_requests, successful_requests, failed_requests, last_request_at)
      VALUES (?, ?, 1, ?, ?, ?)
    `).run(id, keyId, success ? 1 : 0, success ? 0 : 1, now);
    return;
  }

  db.prepare(`
    UPDATE key_stats 
    SET total_requests = total_requests + 1,
        successful_requests = successful_requests + ?,
        failed_requests = failed_requests + ?,
        last_request_at = ?
    WHERE key_id = ?
  `).run(success ? 1 : 0, success ? 0 : 1, now, keyId);
}

export function authMiddleware(req: Request, res: Response, next: NextFunction): void {
  if (req.path.startsWith('/gateway/')) {
    return next();
  }

  const apiKeyHeader = req.headers['x-api-key'] || req.headers['authorization'];
  
  if (!apiKeyHeader || typeof apiKeyHeader !== 'string') {
    res.status(401).json({ error: 'Unauthorized', message: 'API key required' });
    return;
  }

  let plainKey: string;
  if (apiKeyHeader.startsWith('Bearer ')) {
    plainKey = apiKeyHeader.slice(7).trim();
  } else {
    plainKey = apiKeyHeader.trim();
  }

  const keyHash = hashApiKey(plainKey);
  const apiKey = findApiKey(keyHash);

  if (!apiKey) {
    res.status(401).json({ error: 'Unauthorized', message: 'Invalid API key' });
    return;
  }

  if (apiKey.expires_at && Date.now() > apiKey.expires_at) {
    res.status(401).json({ error: 'Unauthorized', message: 'API key has expired' });
    return;
  }

  const matchedPrefix = req.matchedService?.matchedPrefix;
  if (!checkKeyPermissions(apiKey, matchedPrefix)) {
    res.status(403).json({
      error: 'Forbidden',
      message: 'API key does not have permission to access this route'
    });
    return;
  }

  req.apiKey = apiKey;
  next();
}

export function createClient(name: string, description?: string): Client {
  const now = Date.now();
  const id = uuidv4();
  db.prepare(`
    INSERT INTO clients (id, name, description, created_at)
    VALUES (?, ?, ?, ?)
  `).run(id, name, description || null, now);
  
  return db.prepare('SELECT * FROM clients WHERE id = ?').get(id) as Client;
}

export function getAllClients(): Client[] {
  return db.prepare('SELECT * FROM clients ORDER BY created_at DESC').all() as Client[];
}

export function getClientById(id: string): Client | undefined {
  return db.prepare('SELECT * FROM clients WHERE id = ?').get(id) as Client | undefined;
}

export function deleteClient(id: string): boolean {
  const result = db.prepare('DELETE FROM clients WHERE id = ?').run(id);
  return result.changes > 0;
}

export function createApiKeyForClient(
  clientId: string,
  expiresAt?: number,
  permissions: string[] = []
): ApiKey & { plain_key: string } {
  const plainKey = generateApiKey();
  const keyHash = hashApiKey(plainKey);
  const keyPrefix = plainKey.slice(0, 8);
  const now = Date.now();
  const id = uuidv4();

  db.prepare(`
    INSERT INTO api_keys (id, client_id, key_hash, key_prefix, expires_at, permissions, is_active, created_at)
    VALUES (?, ?, ?, ?, ?, ?, 1, ?)
  `).run(
    id,
    clientId,
    keyHash,
    keyPrefix,
    expiresAt || null,
    JSON.stringify(permissions),
    now
  );

  return {
    id,
    client_id: clientId,
    key_hash: keyHash,
    key_prefix: keyPrefix,
    expires_at: expiresAt,
    permissions,
    is_active: 1,
    created_at: now,
    plain_key: plainKey
  };
}

export function getKeysByClientId(clientId: string): Omit<ApiKey, 'key_hash'>[] {
  const rows = db
    .prepare('SELECT id, client_id, key_prefix, expires_at, permissions, is_active, created_at FROM api_keys WHERE client_id = ?')
    .all(clientId) as 
    (Omit<ApiKey, 'key_hash' | 'permissions'> & { permissions: string | null })[];

  return rows.map(row => ({
    ...row,
    permissions: parsePermissions(row.permissions)
  }));
}

export function getKeyStats(keyId: string): KeyStats | undefined {
  return db.prepare('SELECT * FROM key_stats WHERE key_id = ?').get(keyId) as KeyStats | undefined;
}

export function revokeApiKey(keyId: string): boolean {
  const result = db.prepare('UPDATE api_keys SET is_active = 0 WHERE id = ?').run(keyId);
  return result.changes > 0;
}
