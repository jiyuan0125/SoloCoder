import Database, { type Database as DatabaseType } from 'better-sqlite3';
import path from 'path';
import { Channel, FeeTier, PaymentRecord, HealthMetric, PaymentStatus } from './types';

const dbPath = path.join(__dirname, '..', 'payment-gateway.db');
const db = new Database(dbPath);

db.exec(`
  PRAGMA journal_mode = WAL;
  
  CREATE TABLE IF NOT EXISTS channels (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_degraded INTEGER DEFAULT 0,
    degraded_at INTEGER,
    created_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS fee_tiers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id TEXT NOT NULL,
    max_amount INTEGER NOT NULL,
    rate REAL NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES channels(id)
  );

  CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    merchant_order_no TEXT NOT NULL UNIQUE,
    amount INTEGER NOT NULL,
    channel_id TEXT NOT NULL,
    channel_transaction_no TEXT,
    status TEXT NOT NULL,
    refunded_amount INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    paid_at INTEGER,
    closed_at INTEGER,
    refunded_at INTEGER
  );

  CREATE TABLE IF NOT EXISTS health_metrics (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    hour INTEGER NOT NULL,
    total_requests INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    total_response_time INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    UNIQUE(channel_id, hour)
  );
`);

export const getDb = (): DatabaseType => db;

export function getChannelById(id: string): Channel | null {
  const row = db.prepare('SELECT * FROM channels WHERE id = ?').get(id) as any;
  if (!row) return null;
  return rowToChannel(row);
}

export function getAllChannels(): Channel[] {
  const rows = db.prepare('SELECT * FROM channels').all() as any[];
  return rows.map(rowToChannel);
}

export function getActiveChannels(): Channel[] {
  const rows = db.prepare('SELECT * FROM channels WHERE is_degraded = 0').all() as any[];
  return rows.map(rowToChannel);
}

export function updateChannelDegraded(channelId: string, isDegraded: boolean): void {
  const stmt = db.prepare(`
    UPDATE channels 
    SET is_degraded = ?, degraded_at = ?
    WHERE id = ?
  `);
  stmt.run(isDegraded ? 1 : 0, isDegraded ? Date.now() : null, channelId);
}

function rowToChannel(row: any): Channel {
  const feeTiers = db
    .prepare('SELECT max_amount, rate FROM fee_tiers WHERE channel_id = ? ORDER BY max_amount ASC')
    .all(row.id) as any[];
  
  return {
    id: row.id,
    name: row.name,
    feeTiers: feeTiers.map(ft => ({ maxAmount: ft.max_amount, rate: ft.rate })),
    isDegraded: !!row.is_degraded,
    degradedAt: row.degraded_at,
    createdAt: row.created_at,
  };
}

export function getPaymentById(id: string): PaymentRecord | null {
  const row = db.prepare('SELECT * FROM payments WHERE id = ?').get(id) as any;
  if (!row) return null;
  return rowToPayment(row);
}

export function getPaymentByMerchantOrderNo(merchantOrderNo: string): PaymentRecord | null {
  const row = db.prepare('SELECT * FROM payments WHERE merchant_order_no = ?').get(merchantOrderNo) as any;
  if (!row) return null;
  return rowToPayment(row);
}

function rowToPayment(row: any): PaymentRecord {
  return {
    id: row.id,
    merchantOrderNo: row.merchant_order_no,
    amount: row.amount,
    channelId: row.channel_id,
    channelTransactionNo: row.channel_transaction_no,
    status: row.status as PaymentStatus,
    refundedAmount: row.refunded_amount,
    createdAt: row.created_at,
    paidAt: row.paid_at,
    closedAt: row.closed_at,
    refundedAt: row.refunded_at,
  };
}

export function insertPayment(payment: PaymentRecord): void {
  const stmt = db.prepare(`
    INSERT INTO payments (
      id, merchant_order_no, amount, channel_id, 
      channel_transaction_no, status, refunded_amount,
      created_at, paid_at, closed_at, refunded_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  stmt.run(
    payment.id,
    payment.merchantOrderNo,
    payment.amount,
    payment.channelId,
    payment.channelTransactionNo,
    payment.status,
    payment.refundedAmount,
    payment.createdAt,
    payment.paidAt,
    payment.closedAt,
    payment.refundedAt,
  );
}

export function updatePaymentStatus(
  id: string,
  status: PaymentStatus,
  extra: { channelTransactionNo?: string; refundedAmount?: number } = {}
): void {
  const now = Date.now();
  const updates: string[] = ['status = ?'];
  const params: any[] = [status, id];

  if (extra.channelTransactionNo !== undefined) {
    updates.push('channel_transaction_no = ?');
    params.splice(params.length - 1, 0, extra.channelTransactionNo);
  }

  if (status === 'PAID') {
    updates.push('paid_at = ?');
    params.splice(params.length - 1, 0, now);
  } else if (status === 'CLOSED') {
    updates.push('closed_at = ?');
    params.splice(params.length - 1, 0, now);
  } else if (status === 'REFUNDED') {
    updates.push('refunded_at = ?');
    params.splice(params.length - 1, 0, now);
    if (extra.refundedAmount !== undefined) {
      updates.push('refunded_amount = ?');
      params.splice(params.length - 1, 0, extra.refundedAmount);
    }
  }

  const sql = `UPDATE payments SET ${updates.join(', ')} WHERE id = ?`;
  db.prepare(sql).run(...params);
}

export function incrementRefundedAmount(id: string, refundAmount: number): void {
  db.prepare('UPDATE payments SET refunded_amount = refunded_amount + ? WHERE id = ?').run(refundAmount, id);
}

export function getOrCreateHealthMetric(channelId: string, hour: number): HealthMetric {
  const existing = db
    .prepare('SELECT * FROM health_metrics WHERE channel_id = ? AND hour = ?')
    .get(channelId, hour) as any;

  if (existing) {
    return rowToMetric(existing);
  }

  const id = `metric_${channelId}_${hour}`;
  const now = Date.now();
  db.prepare(`
    INSERT INTO health_metrics (id, channel_id, hour, total_requests, success_count, total_response_time, created_at)
    VALUES (?, ?, ?, 0, 0, 0, ?)
  `).run(id, channelId, hour, now);

  return {
    id,
    channelId,
    hour,
    totalRequests: 0,
    successCount: 0,
    totalResponseTime: 0,
    createdAt: now,
  };
}

export function recordHealthMetric(
  channelId: string,
  hour: number,
  success: boolean,
  responseTimeMs: number
): void {
  db.prepare(`
    UPDATE health_metrics 
    SET total_requests = total_requests + 1,
        success_count = success_count + ?,
        total_response_time = total_response_time + ?
    WHERE channel_id = ? AND hour = ?
  `).run(success ? 1 : 0, responseTimeMs, channelId, hour);
}

export function getHealthMetricsSince(channelId: string, sinceHour: number): HealthMetric[] {
  const rows = db
    .prepare('SELECT * FROM health_metrics WHERE channel_id = ? AND hour >= ?')
    .all(channelId, sinceHour) as any[];
  return rows.map(rowToMetric);
}

function rowToMetric(row: any): HealthMetric {
  return {
    id: row.id,
    channelId: row.channel_id,
    hour: row.hour,
    totalRequests: row.total_requests,
    successCount: row.success_count,
    totalResponseTime: row.total_response_time,
    createdAt: row.created_at,
  };
}

export function initializeChannels(): void {
  const channels = [
    {
      id: 'wechat',
      name: '微信支付',
      tiers: [
        { maxAmount: 100000, rate: 0.0038 },
        { maxAmount: Infinity, rate: 0.0032 },
      ],
    },
    {
      id: 'alipay',
      name: '支付宝',
      tiers: [
        { maxAmount: 50000, rate: 0.0042 },
        { maxAmount: 200000, rate: 0.0035 },
        { maxAmount: Infinity, rate: 0.0030 },
      ],
    },
    {
      id: 'unionpay',
      name: '银联',
      tiers: [
        { maxAmount: 200000, rate: 0.0033 },
        { maxAmount: Infinity, rate: 0.0028 },
      ],
    },
  ];

  const insertChannel = db.prepare(`
    INSERT OR IGNORE INTO channels (id, name, is_degraded, created_at)
    VALUES (?, ?, 0, ?)
  `);

  const insertTier = db.prepare(`
    INSERT OR IGNORE INTO fee_tiers (channel_id, max_amount, rate)
    VALUES (?, ?, ?)
  `);

  const now = Date.now();
  for (const ch of channels) {
    insertChannel.run(ch.id, ch.name, now);
    for (const tier of ch.tiers) {
      const maxAmt = tier.maxAmount === Infinity ? 999999999999 : tier.maxAmount;
      insertTier.run(ch.id, maxAmt, tier.rate);
    }
  }
}
