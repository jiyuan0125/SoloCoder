import Database from 'better-sqlite3';
import { v4 as uuidv4 } from 'uuid';
import { Webhook, Delivery, DeliveryStatus } from './types';

const DB_PATH = './webhook.db';

let dbInstance: Database.Database | null = null;

function getDb(): Database.Database {
  if (!dbInstance) {
    dbInstance = new Database(DB_PATH);
    dbInstance.pragma('journal_mode = WAL');
    initDb(dbInstance);
  }
  return dbInstance;
}

function initDb(db: Database.Database): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS webhooks (
      id TEXT PRIMARY KEY,
      event_type TEXT NOT NULL,
      callback_url TEXT NOT NULL,
      secret TEXT NOT NULL,
      is_active INTEGER DEFAULT 1,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS deliveries (
      id TEXT PRIMARY KEY,
      webhook_id TEXT NOT NULL,
      event_type TEXT NOT NULL,
      event_data TEXT NOT NULL,
      request_url TEXT NOT NULL,
      request_body TEXT NOT NULL,
      response_status INTEGER,
      response_body TEXT,
      status TEXT NOT NULL DEFAULT 'pending',
      retry_count INTEGER NOT NULL DEFAULT 0,
      next_retry_at TEXT,
      last_delivered_at TEXT,
      created_at TEXT NOT NULL,
      FOREIGN KEY (webhook_id) REFERENCES webhooks(id)
    );

    CREATE INDEX IF NOT EXISTS idx_deliveries_webhook_id ON deliveries(webhook_id);
    CREATE INDEX IF NOT EXISTS idx_deliveries_status ON deliveries(status);
    CREATE INDEX IF NOT EXISTS idx_deliveries_next_retry ON deliveries(next_retry_at);
    CREATE INDEX IF NOT EXISTS idx_webhooks_event_type ON webhooks(event_type);
  `);
}

export function createWebhook(eventType: string, callbackUrl: string, secret: string): Webhook {
  const db = getDb();
  const now = new Date().toISOString();
  const id = uuidv4();
  
  const stmt = db.prepare(`
    INSERT INTO webhooks (id, event_type, callback_url, secret, is_active, created_at, updated_at)
    VALUES (?, ?, ?, ?, 1, ?, ?)
  `);
  stmt.run(id, eventType, callbackUrl, secret, now, now);
  
  return getWebhookById(id)!;
}

export function getWebhookById(id: string): Webhook | undefined {
  const db = getDb();
  return db.prepare('SELECT * FROM webhooks WHERE id = ?').get(id) as Webhook | undefined;
}

export function getWebhooksByEventType(eventType: string): Webhook[] {
  const db = getDb();
  return db.prepare('SELECT * FROM webhooks WHERE event_type = ? AND is_active = 1').all(eventType) as Webhook[];
}

export function listWebhooks(): Webhook[] {
  const db = getDb();
  return db.prepare('SELECT * FROM webhooks ORDER BY created_at DESC').all() as Webhook[];
}

export function updateWebhook(id: string, callbackUrl?: string, isActive?: boolean): Webhook | undefined {
  const db = getDb();
  const now = new Date().toISOString();
  
  const fields: string[] = [];
  const values: any[] = [];
  
  if (callbackUrl !== undefined) {
    fields.push('callback_url = ?');
    values.push(callbackUrl);
  }
  if (isActive !== undefined) {
    fields.push('is_active = ?');
    values.push(isActive ? 1 : 0);
  }
  
  if (fields.length === 0) return getWebhookById(id);
  
  fields.push('updated_at = ?');
  values.push(now, id);
  
  const stmt = db.prepare(`UPDATE webhooks SET ${fields.join(', ')} WHERE id = ?`);
  stmt.run(...values);
  
  return getWebhookById(id);
}

export function deleteWebhook(id: string): void {
  const db = getDb();
  const tx = db.transaction(() => {
    db.prepare('DELETE FROM deliveries WHERE webhook_id = ?').run(id);
    db.prepare('DELETE FROM webhooks WHERE id = ?').run(id);
  });
  tx();
}

export function hasPendingDeliveries(webhookId: string): boolean {
  const db = getDb();
  const result = db.prepare(`
    SELECT COUNT(*) as count FROM deliveries 
    WHERE webhook_id = ? AND status IN ('pending', 'processing')
  `).get(webhookId) as { count: number };
  return result.count > 0;
}

export function createDelivery(
  webhookId: string,
  eventType: string,
  eventData: string,
  requestUrl: string,
  requestBody: string
): Delivery {
  const db = getDb();
  const now = new Date().toISOString();
  const id = uuidv4();
  
  const stmt = db.prepare(`
    INSERT INTO deliveries (id, webhook_id, event_type, event_data, request_url, request_body, status, created_at)
    VALUES (?, ?, ?, ?, ?, ?, 'pending', ?)
  `);
  stmt.run(id, webhookId, eventType, eventData, requestUrl, requestBody, now);
  
  return getDeliveryById(id)!;
}

export function getDeliveryById(id: string): Delivery | undefined {
  const db = getDb();
  return db.prepare('SELECT * FROM deliveries WHERE id = ?').get(id) as Delivery | undefined;
}

export function getDeliveriesByWebhookId(webhookId: string): Delivery[] {
  const db = getDb();
  return db.prepare('SELECT * FROM deliveries WHERE webhook_id = ? ORDER BY created_at DESC').all(webhookId) as Delivery[];
}

export function getPendingDeliveries(webhookId: string): Delivery[] {
  const db = getDb();
  return db.prepare(`
    SELECT * FROM deliveries 
    WHERE webhook_id = ? AND status IN ('pending', 'processing')
    ORDER BY created_at ASC
  `).all(webhookId) as Delivery[];
}

export function getDeliveriesReadyToSend(): Delivery[] {
  const db = getDb();
  const now = new Date().toISOString();
  return db.prepare(`
    SELECT d.* FROM deliveries d
    INNER JOIN webhooks w ON d.webhook_id = w.id
    WHERE w.is_active = 1 
    AND d.status IN ('pending', 'failed')
    AND (d.next_retry_at IS NULL OR d.next_retry_at <= ?)
    ORDER BY d.created_at ASC
  `).all(now) as Delivery[];
}

export function updateDeliveryStatus(
  id: string,
  status: DeliveryStatus,
  responseStatus?: number,
  responseBody?: string,
  retryCount?: number,
  nextRetryAt?: string | null
): void {
  const db = getDb();
  const now = new Date().toISOString();
  
  const fields: string[] = ['status = ?'];
  const values: any[] = [status];
  
  if (responseStatus !== undefined) {
    fields.push('response_status = ?');
    values.push(responseStatus);
  }
  if (responseBody !== undefined) {
    fields.push('response_body = ?');
    values.push(responseBody.slice(0, 500));
  }
  if (retryCount !== undefined) {
    fields.push('retry_count = ?');
    values.push(retryCount);
  }
  if (nextRetryAt !== undefined) {
    fields.push('next_retry_at = ?');
    values.push(nextRetryAt);
  }
  
  fields.push('last_delivered_at = ?');
  values.push(now, id);
  
  const stmt = db.prepare(`UPDATE deliveries SET ${fields.join(', ')} WHERE id = ?`);
  stmt.run(...values);
}

export function getDbInstance(): Database.Database {
  return getDb();
}
