import Database from 'better-sqlite3';
import { join } from 'path';
import { AdPlan, HourlyStats, LedgerRecord } from './types';

const dbPath = join(process.cwd(), 'ads.db');
const db = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('synchronous = NORMAL');

db.exec(`
CREATE TABLE IF NOT EXISTS ad_plans (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  budget INTEGER NOT NULL,
  daily_budget INTEGER NOT NULL,
  targeting TEXT NOT NULL,
  bid_type TEXT NOT NULL CHECK (bid_type IN ('CPC', 'CPM')),
  bid_amount INTEGER NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'ended')),
  expected_ctr REAL NOT NULL DEFAULT 0.05,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS hourly_stats (
  id TEXT PRIMARY KEY,
  plan_id TEXT NOT NULL,
  hour INTEGER NOT NULL,
  impressions INTEGER NOT NULL DEFAULT 0,
  clicks INTEGER NOT NULL DEFAULT 0,
  conversions INTEGER NOT NULL DEFAULT 0,
  spend INTEGER NOT NULL DEFAULT 0,
  created_at INTEGER NOT NULL,
  UNIQUE(plan_id, hour)
);

CREATE TABLE IF NOT EXISTS ledger (
  id TEXT PRIMARY KEY,
  plan_id TEXT NOT NULL,
  amount INTEGER NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('deduction', 'refund', 'overcharge')),
  timestamp INTEGER NOT NULL,
  related_id TEXT
);

CREATE TABLE IF NOT EXISTS overcharges (
  id TEXT PRIMARY KEY,
  plan_id TEXT NOT NULL UNIQUE,
  amount INTEGER NOT NULL DEFAULT 0,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_hourly_stats_plan ON hourly_stats(plan_id);
CREATE INDEX IF NOT EXISTS idx_ledger_plan ON ledger(plan_id);
`);

function parseTargeting(t: string) {
  return JSON.parse(t);
}

export const dbMethods = {
  insertPlan(p: Omit<AdPlan, 'createdAt' | 'updatedAt'>) {
    const now = Date.now();
    const stmt = db.prepare(`
      INSERT INTO ad_plans 
      (id, name, budget, daily_budget, targeting, bid_type, bid_amount, status, expected_ctr, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(p.id, p.name, p.budget, p.dailyBudget, JSON.stringify(p.targeting), p.bidType, p.bidAmount, p.status, p.expectedCtr, now, now);
    return this.getPlan(p.id)!;
  },

  getPlan(id: string): AdPlan | null {
    const row = db.prepare('SELECT * FROM ad_plans WHERE id = ?').get(id) as any;
    if (!row) return null;
    return {
      id: row.id,
      name: row.name,
      budget: row.budget,
      dailyBudget: row.daily_budget,
      targeting: parseTargeting(row.targeting),
      bidType: row.bid_type,
      bidAmount: row.bid_amount,
      status: row.status,
      expectedCtr: row.expected_ctr,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    };
  },

  updatePlanStatus(id: string, status: AdPlan['status']) {
    const now = Date.now();
    db.prepare('UPDATE ad_plans SET status = ?, updated_at = ? WHERE id = ?').run(status, now, id);
    return this.getPlan(id)!;
  },

  updatePlanExpectedCtr(id: string, ctr: number) {
    const now = Date.now();
    db.prepare('UPDATE ad_plans SET expected_ctr = ?, updated_at = ? WHERE id = ?').run(ctr, now, id);
  },

  getRunningPlans() {
    const rows = db.prepare("SELECT * FROM ad_plans WHERE status = 'running'").all() as any[];
    return rows.map((row: any) => ({
      id: row.id,
      name: row.name,
      budget: row.budget,
      dailyBudget: row.daily_budget,
      targeting: parseTargeting(row.targeting),
      bidType: row.bid_type,
      bidAmount: row.bid_amount,
      status: row.status,
      expectedCtr: row.expected_ctr,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    })) as AdPlan[];
  },

  getTotalSpend(planId: string): number {
    const row = db.prepare('SELECT COALESCE(SUM(amount), 0) as total FROM ledger WHERE plan_id = ? AND type = ?').get(planId, 'deduction') as any;
    const refundRow = db.prepare('SELECT COALESCE(SUM(amount), 0) as total FROM ledger WHERE plan_id = ? AND type IN (?, ?)').get(planId, 'refund', 'overcharge') as any;
    return row.total - refundRow.total;
  },

  getTodaySpend(planId: string): number {
    const now = new Date();
    const startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
    const row = db.prepare(`
      SELECT COALESCE(SUM(spend), 0) as total FROM hourly_stats 
      WHERE plan_id = ? AND hour >= ?
    `).get(planId, startOfDay) as any;
    return row.total;
  },

  getPlanStats(planId: string) {
    const row = db.prepare(`
      SELECT 
        COALESCE(SUM(impressions), 0) as impressions,
        COALESCE(SUM(clicks), 0) as clicks,
        COALESCE(SUM(conversions), 0) as conversions,
        COALESCE(SUM(spend), 0) as spend
      FROM hourly_stats WHERE plan_id = ?
    `).get(planId) as any;
    return {
      impressions: row.impressions,
      clicks: row.clicks,
      conversions: row.conversions,
      spend: row.spend,
    };
  },

  updateHourlyStats(planId: string, hour: number, data: Partial<Pick<HourlyStats, 'impressions' | 'clicks' | 'conversions' | 'spend'>>) {
    const existing = db.prepare('SELECT * FROM hourly_stats WHERE plan_id = ? AND hour = ?').get(planId, hour) as any;
    if (existing) {
      const impressions = existing.impressions + (data.impressions || 0);
      const clicks = existing.clicks + (data.clicks || 0);
      const conversions = existing.conversions + (data.conversions || 0);
      const spend = existing.spend + (data.spend || 0);
      db.prepare(`
        UPDATE hourly_stats SET impressions = ?, clicks = ?, conversions = ?, spend = ? 
        WHERE plan_id = ? AND hour = ?
      `).run(impressions, clicks, conversions, spend, planId, hour);
    } else {
      db.prepare(`
        INSERT INTO hourly_stats (id, plan_id, hour, impressions, clicks, conversions, spend, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `).run(
        Math.random().toString(36).slice(2),
        planId,
        hour,
        data.impressions || 0,
        data.clicks || 0,
        data.conversions || 0,
        data.spend || 0,
        Date.now()
      );
    }
  },

  insertLedger(record: Omit<LedgerRecord, 'id' | 'timestamp'>) {
    const id = Math.random().toString(36).slice(2);
    db.prepare(`
      INSERT INTO ledger (id, plan_id, amount, type, timestamp, related_id)
      VALUES (?, ?, ?, ?, ?, ?)
    `).run(id, record.planId, record.amount, record.type, Date.now(), record.relatedId || null);
  },

  getOvercharge(planId: string): number {
    const row = db.prepare('SELECT amount FROM overcharges WHERE plan_id = ?').get(planId) as any;
    return row ? row.amount : 0;
  },

  setOvercharge(planId: string, amount: number) {
    const now = Date.now();
    const existing = db.prepare('SELECT id FROM overcharges WHERE plan_id = ?').get(planId);
    if (existing) {
      db.prepare('UPDATE overcharges SET amount = ?, updated_at = ? WHERE plan_id = ?').run(amount, now, planId);
    } else {
      db.prepare('INSERT INTO overcharges (id, plan_id, amount, created_at, updated_at) VALUES (?, ?, ?, ?, ?)').run(
        Math.random().toString(36).slice(2), planId, amount, now, now
      );
    }
  },

  transaction<T>(fn: () => T): T {
    return db.transaction(fn)();
  },
};

export default db;
