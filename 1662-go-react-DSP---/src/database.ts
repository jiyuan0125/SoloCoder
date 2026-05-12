import sqlite3 from 'sqlite3';
import { v4 as uuidv4 } from 'uuid';
import {
  Campaign,
  CampaignStatus,
  CampaignCreateRequest,
  CampaignUpdateRequest,
  BidLog,
  BidLogStatus,
  BidParticipant,
  CampaignStatistics,
  BiddingStrategy,
  Targeting
} from './types';

const db = new sqlite3.Database(':memory:');

export function initDatabase(): Promise<void> {
  return new Promise((resolve, reject) => {
    db.serialize(() => {
      db.run(`
        CREATE TABLE IF NOT EXISTS campaigns (
          id TEXT PRIMARY KEY,
          name TEXT NOT NULL UNIQUE,
          daily_budget REAL NOT NULL,
          total_budget REAL NOT NULL,
          targeting TEXT NOT NULL,
          bidding_strategy TEXT NOT NULL,
          bid_amount REAL NOT NULL,
          status TEXT NOT NULL,
          daily_spent REAL NOT NULL DEFAULT 0,
          total_spent REAL NOT NULL DEFAULT 0,
          created_at TEXT NOT NULL,
          updated_at TEXT NOT NULL
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS bid_logs (
          id TEXT PRIMARY KEY,
          request_id TEXT NOT NULL,
          participants TEXT NOT NULL,
          winner_campaign_id TEXT,
          winner_campaign_name TEXT,
          final_price REAL NOT NULL DEFAULT 0,
          status TEXT NOT NULL,
          created_at TEXT NOT NULL
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS campaign_bid_stats (
          id TEXT PRIMARY KEY,
          campaign_id TEXT NOT NULL,
          bid_count INTEGER NOT NULL DEFAULT 0,
          win_count INTEGER NOT NULL DEFAULT 0,
          FOREIGN KEY (campaign_id) REFERENCES campaigns(id)
        )
      `);

      resolve();
    });
  });
}

function rowToCampaign(row: any): Campaign {
  return {
    id: row.id,
    name: row.name,
    dailyBudget: row.daily_budget,
    totalBudget: row.total_budget,
    targeting: JSON.parse(row.targeting),
    biddingStrategy: row.bidding_strategy as BiddingStrategy,
    bidAmount: row.bid_amount,
    status: row.status as CampaignStatus,
    dailySpent: row.daily_spent,
    totalSpent: row.total_spent,
    createdAt: row.created_at,
    updatedAt: row.updated_at
  };
}

export function getCampaignById(id: string): Promise<Campaign | null> {
  return new Promise((resolve, reject) => {
    db.get('SELECT * FROM campaigns WHERE id = ?', [id], (err, row: any) => {
      if (err) return reject(err);
      resolve(row ? rowToCampaign(row) : null);
    });
  });
}

export function getCampaignByName(name: string): Promise<Campaign | null> {
  return new Promise((resolve, reject) => {
    db.get('SELECT * FROM campaigns WHERE name = ?', [name], (err, row: any) => {
      if (err) return reject(err);
      resolve(row ? rowToCampaign(row) : null);
    });
  });
}

export function getAllCampaigns(): Promise<Campaign[]> {
  return new Promise((resolve, reject) => {
    db.all('SELECT * FROM campaigns', (err, rows: any[]) => {
      if (err) return reject(err);
      resolve(rows.map(rowToCampaign));
    });
  });
}

export function getActiveCampaigns(): Promise<Campaign[]> {
  return new Promise((resolve, reject) => {
    db.all(
      'SELECT * FROM campaigns WHERE status = ?',
      [CampaignStatus.ACTIVE],
      (err, rows: any[]) => {
        if (err) return reject(err);
        resolve(rows.map(rowToCampaign));
      }
    );
  });
}

export function createCampaign(req: CampaignCreateRequest): Promise<Campaign> {
  const id = uuidv4();
  const now = new Date().toISOString();
  const campaign: Campaign = {
    id,
    name: req.name,
    dailyBudget: req.dailyBudget,
    totalBudget: req.totalBudget,
    targeting: req.targeting,
    biddingStrategy: req.biddingStrategy,
    bidAmount: req.bidAmount,
    status: CampaignStatus.ACTIVE,
    dailySpent: 0,
    totalSpent: 0,
    createdAt: now,
    updatedAt: now
  };

  return new Promise((resolve, reject) => {
    db.run(
      `INSERT INTO campaigns (id, name, daily_budget, total_budget, targeting, bidding_strategy, bid_amount, status, daily_spent, total_spent, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        id,
        req.name,
        req.dailyBudget,
        req.totalBudget,
        JSON.stringify(req.targeting),
        req.biddingStrategy,
        req.bidAmount,
        CampaignStatus.ACTIVE,
        0,
        0,
        now,
        now
      ],
      function (err) {
        if (err) return reject(err);
        db.run(
          'INSERT INTO campaign_bid_stats (id, campaign_id, bid_count, win_count) VALUES (?, ?, 0, 0)',
          [uuidv4(), id],
          (err2) => {
            if (err2) return reject(err2);
            resolve(campaign);
          }
        );
      }
    );
  });
}

export function updateCampaign(id: string, req: CampaignUpdateRequest): Promise<Campaign | null> {
  return new Promise(async (resolve, reject) => {
    const existing = await getCampaignById(id);
    if (!existing) return resolve(null);

    const now = new Date().toISOString();
    const updated: Campaign = {
      ...existing,
      name: req.name ?? existing.name,
      dailyBudget: req.dailyBudget ?? existing.dailyBudget,
      totalBudget: req.totalBudget ?? existing.totalBudget,
      targeting: req.targeting ?? existing.targeting,
      biddingStrategy: req.biddingStrategy ?? existing.biddingStrategy,
      bidAmount: req.bidAmount ?? existing.bidAmount,
      status: req.status ?? existing.status,
      updatedAt: now
    };

    db.run(
      `UPDATE campaigns 
       SET name = ?, daily_budget = ?, total_budget = ?, targeting = ?, bidding_strategy = ?, bid_amount = ?, status = ?, updated_at = ?
       WHERE id = ?`,
      [
        updated.name,
        updated.dailyBudget,
        updated.totalBudget,
        JSON.stringify(updated.targeting),
        updated.biddingStrategy,
        updated.bidAmount,
        updated.status,
        now,
        id
      ],
      function (err) {
        if (err) return reject(err);
        resolve(updated);
      }
    );
  });
}

export function updateCampaignStatus(id: string, status: CampaignStatus): Promise<boolean> {
  return new Promise((resolve, reject) => {
    const now = new Date().toISOString();
    db.run(
      'UPDATE campaigns SET status = ?, updated_at = ? WHERE id = ?',
      [status, now, id],
      function (err) {
        if (err) return reject(err);
        resolve(this.changes > 0);
      }
    );
  });
}

export function checkDailyAndTotalBudget(): Promise<void> {
  return new Promise((resolve, reject) => {
    db.run(
      `UPDATE campaigns 
       SET status = ?, updated_at = ?
       WHERE status = ? AND (daily_spent >= daily_budget OR total_spent >= total_budget)`,
      [CampaignStatus.ENDED, new Date().toISOString(), CampaignStatus.ACTIVE],
      function (err) {
        if (err) return reject(err);
        resolve();
      }
    );
  });
}

export function incrementBidCount(campaignIds: string[]): Promise<void> {
  if (campaignIds.length === 0) return Promise.resolve();
  
  return new Promise((resolve, reject) => {
    const placeholders = campaignIds.map(() => '?').join(',');
    db.run(
      `UPDATE campaign_bid_stats SET bid_count = bid_count + 1 WHERE campaign_id IN (${placeholders})`,
      campaignIds,
      function (err) {
        if (err) return reject(err);
        resolve();
      }
    );
  });
}

export function incrementWinCount(campaignId: string): Promise<void> {
  return new Promise((resolve, reject) => {
    db.run(
      'UPDATE campaign_bid_stats SET win_count = win_count + 1 WHERE campaign_id = ?',
      [campaignId],
      function (err) {
        if (err) return reject(err);
        resolve();
      }
    );
  });
}

export function deductBudget(
  campaignId: string,
  amount: number
): Promise<{ success: boolean; dailyExceeded: boolean; totalExceeded: boolean }> {
  return new Promise((resolve, reject) => {
    db.serialize(() => {
      db.get(
        'SELECT daily_budget, total_budget, daily_spent, total_spent FROM campaigns WHERE id = ? AND status = ?',
        [campaignId, CampaignStatus.ACTIVE],
        (err, row: any) => {
          if (err) return reject(err);
          if (!row) return resolve({ success: false, dailyExceeded: false, totalExceeded: false });

          const newDailySpent = row.daily_spent + amount;
          const newTotalSpent = row.total_spent + amount;

          if (newDailySpent > row.daily_budget) {
            return resolve({ success: false, dailyExceeded: true, totalExceeded: false });
          }
          if (newTotalSpent > row.total_budget) {
            return resolve({ success: false, dailyExceeded: false, totalExceeded: true });
          }

          db.run(
            'UPDATE campaigns SET daily_spent = ?, total_spent = ?, updated_at = ? WHERE id = ?',
            [newDailySpent, newTotalSpent, new Date().toISOString(), campaignId],
            function (err2) {
              if (err2) return reject(err2);
              resolve({ success: true, dailyExceeded: false, totalExceeded: false });
            }
          );
        }
      );
    });
  });
}

export function createBidLog(
  requestId: string,
  participants: BidParticipant[],
  winner: { id: string; name: string } | null,
  finalPrice: number,
  status: BidLogStatus
): Promise<BidLog> {
  const id = uuidv4();
  const now = new Date().toISOString();

  return new Promise((resolve, reject) => {
    db.run(
      `INSERT INTO bid_logs (id, request_id, participants, winner_campaign_id, winner_campaign_name, final_price, status, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        id,
        requestId,
        JSON.stringify(participants),
        winner?.id ?? null,
        winner?.name ?? null,
        finalPrice,
        status,
        now
      ],
      function (err) {
        if (err) return reject(err);
        resolve({
          id,
          requestId,
          participants: JSON.stringify(participants),
          winnerCampaignId: winner?.id ?? null,
          winnerCampaignName: winner?.name ?? null,
          finalPrice,
          status,
          createdAt: now
        });
      }
    );
  });
}

export function getCampaignStatistics(): Promise<CampaignStatistics[]> {
  return new Promise((resolve, reject) => {
    db.all(
      `SELECT 
         c.id as campaign_id,
         c.name as campaign_name,
         c.total_spent as total_spent,
         COALESCE(s.bid_count, 0) as bid_count,
         COALESCE(s.win_count, 0) as win_count
       FROM campaigns c
       LEFT JOIN campaign_bid_stats s ON c.id = s.campaign_id`,
      (err, rows: any[]) => {
        if (err) return reject(err);
        resolve(
          rows.map((row) => ({
            campaignId: row.campaign_id,
            campaignName: row.campaign_name,
            bidCount: row.bid_count,
            winCount: row.win_count,
            winRate: row.bid_count > 0 ? row.win_count / row.bid_count : 0,
            totalSpent: row.total_spent
          }))
        );
      }
    );
  });
}

export function closeDatabase(): Promise<void> {
  return new Promise((resolve, reject) => {
    db.close((err) => {
      if (err) return reject(err);
      resolve();
    });
  });
}
