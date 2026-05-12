import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { AccountType, InterestRecord, InterestSettlement } from '../types';

const router = Router();

const formatInterestRecord = (row: any): InterestRecord => ({
  id: row.id,
  accountId: row.account_id,
  balanceSnapshot: row.balance_snapshot,
  dailyInterest: row.daily_interest,
  date: row.date,
  createdAt: row.created_at
});

const formatInterestSettlement = (row: any): InterestSettlement => ({
  id: row.id,
  accountId: row.account_id,
  year: row.year,
  month: row.month,
  totalInterest: row.total_interest,
  settlementDate: row.settlement_date,
  createdAt: row.created_at
});

const getDailyRate = (): number => {
  const config = db.prepare('SELECT daily_interest_rate FROM system_config WHERE id = ?').get('config') as { daily_interest_rate: number } | undefined;
  return config ? config.daily_interest_rate : 0.0001;
};

const getMainPoolAccount = () => {
  return db.prepare('SELECT * FROM accounts WHERE type = ? LIMIT 1').get(AccountType.MAIN_POOL) as { id: string; balance: number } | undefined;
};

router.get('/records', (req: Request, res: Response) => {
  const { accountId, startDate, endDate } = req.query;
  
  let sql = 'SELECT * FROM interest_records WHERE 1=1';
  const params: any[] = [];

  if (accountId) {
    sql += ' AND account_id = ?';
    params.push(accountId);
  }
  if (startDate) {
    sql += ' AND date >= ?';
    params.push(startDate);
  }
  if (endDate) {
    sql += ' AND date <= ?';
    params.push(endDate);
  }
  sql += ' ORDER BY date DESC';

  const rows = db.prepare(sql).all(...params);
  res.json(rows.map(formatInterestRecord));
});

router.get('/settlements', (req: Request, res: Response) => {
  const { accountId, year, month } = req.query;
  
  let sql = 'SELECT * FROM interest_settlements WHERE 1=1';
  const params: any[] = [];

  if (accountId) {
    sql += ' AND account_id = ?';
    params.push(accountId);
  }
  if (year) {
    sql += ' AND year = ?';
    params.push(Number(year));
  }
  if (month) {
    sql += ' AND month = ?';
    params.push(Number(month));
  }
  sql += ' ORDER BY year DESC, month DESC';

  const rows = db.prepare(sql).all(...params);
  res.json(rows.map(formatInterestSettlement));
});

router.post('/calculate-daily', (req: Request, res: Response) => {
  const { date } = req.body;
  const targetDate = date || new Date().toISOString().split('T')[0];

  const lockResult = db.prepare('UPDATE locks SET locked = 1, updated_at = datetime(\'now\') WHERE id = ? AND locked = 0')
    .run('interest_calc');

  if (lockResult.changes === 0) {
    return res.status(409).json({ error: '利息计算正在进行中，请稍后再试' });
  }

  try {
    const mainAccount = getMainPoolAccount();
    if (!mainAccount) {
      db.prepare('UPDATE locks SET locked = 0, updated_at = datetime(\'now\') WHERE id = ?').run('interest_calc');
      return res.status(404).json({ error: '未找到资金池主账户' });
    }

    const existing = db.prepare('SELECT id FROM interest_records WHERE account_id = ? AND date = ?')
      .get(mainAccount.id, targetDate);
    
    if (existing) {
      db.prepare('UPDATE locks SET locked = 0, updated_at = datetime(\'now\') WHERE id = ?').run('interest_calc');
      return res.status(409).json({ error: '该日期利息已计算' });
    }

    const balanceSnapshot = mainAccount.balance;
    const dailyRate = getDailyRate();
    const dailyInterest = Math.floor(balanceSnapshot * dailyRate);
    const now = new Date().toISOString();
    const id = uuidv4();

    db.prepare(`
      INSERT INTO interest_records (id, account_id, balance_snapshot, daily_interest, date, created_at)
      VALUES (?, ?, ?, ?, ?, ?)
    `).run(id, mainAccount.id, balanceSnapshot, dailyInterest, targetDate, now);

    const record = db.prepare('SELECT * FROM interest_records WHERE id = ?').get(id);
    
    db.prepare('UPDATE locks SET locked = 0, updated_at = datetime(\'now\') WHERE id = ?').run('interest_calc');
    res.status(201).json(formatInterestRecord(record));
  } catch (error) {
    db.prepare('UPDATE locks SET locked = 0, updated_at = datetime(\'now\') WHERE id = ?').run('interest_calc');
    throw error;
  }
});

router.post('/settle-monthly', (req: Request, res: Response) => {
  const { year, month } = req.body;
  const now = new Date();
  const targetYear = year !== undefined ? Number(year) : now.getFullYear();
  const targetMonth = month !== undefined ? Number(month) : now.getMonth() + 1;

  const lockResult = db.prepare('UPDATE locks SET locked = 1, updated_at = datetime(\'now\') WHERE id = ? AND locked = 0')
    .run('interest_calc');

  if (lockResult.changes === 0) {
    return res.status(409).json({ error: '利息计算正在进行中，请稍后再试' });
  }

  try {
    const mainAccount = getMainPoolAccount();
    if (!mainAccount) {
      db.prepare('UPDATE locks SET locked = 0, updated_at = datetime(\'now\') WHERE id = ?').run('interest_calc');
      return res.status(404).json({ error: '未找到资金池主账户' });
    }

    const existingSettlement = db.prepare(
      'SELECT id FROM interest_settlements WHERE account_id = ? AND year = ? AND month = ?'
    ).get(mainAccount.id, targetYear, targetMonth);

    if (existingSettlement) {
      db.prepare('UPDATE locks SET locked = 0, updated_at = datetime(\'now\') WHERE id = ?').run('interest_calc');
      return res.status(409).json({ error: '该月份利息已结息' });
    }

    const startDate = `${targetYear}-${String(targetMonth).padStart(2, '0')}-01`;
    const endDate = `${targetYear}-${String(targetMonth).padStart(2, '0')}-31`;

    const result = db.prepare(`
      SELECT COALESCE(SUM(daily_interest), 0) as total
      FROM interest_records
      WHERE account_id = ? AND date >= ? AND date <= ?
    `).get(mainAccount.id, startDate, endDate) as { total: number };

    const totalInterest = result.total || 0;
    const createdAt = new Date().toISOString();
    const settlementDate = `${targetYear}-${String(targetMonth).padStart(2, '0')}-20`;
    const id = uuidv4();

    const tx = db.transaction(() => {
      db.prepare(`
        INSERT INTO interest_settlements (id, account_id, year, month, total_interest, settlement_date, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
      `).run(id, mainAccount.id, targetYear, targetMonth, totalInterest, settlementDate, createdAt);

      if (totalInterest > 0) {
        db.prepare('UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?')
          .run(totalInterest, createdAt, mainAccount.id);
      }
    });

    tx();

    const settlement = db.prepare('SELECT * FROM interest_settlements WHERE id = ?').get(id);
    
    db.prepare('UPDATE locks SET locked = 0, updated_at = datetime(\'now\') WHERE id = ?').run('interest_calc');
    res.status(201).json(formatInterestSettlement(settlement));
  } catch (error) {
    db.prepare('UPDATE locks SET locked = 0, updated_at = datetime(\'now\') WHERE id = ?').run('interest_calc');
    throw error;
  }
});

export default router;
