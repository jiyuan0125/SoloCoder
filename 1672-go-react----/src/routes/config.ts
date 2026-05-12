import { Router, Request, Response } from 'express';
import db from '../database';
import { SystemConfig } from '../types';

const router = Router();

const formatConfig = (row: any): SystemConfig => ({
  dailyInterestRate: row.daily_interest_rate,
  updatedAt: row.updated_at
});

router.get('/', (_req: Request, res: Response) => {
  const config = db.prepare('SELECT * FROM system_config WHERE id = ?').get('config');
  res.json(formatConfig(config));
});

router.put('/', (req: Request, res: Response) => {
  const { dailyInterestRate } = req.body;

  if (typeof dailyInterestRate !== 'number' || dailyInterestRate < 0) {
    return res.status(400).json({ error: '日利率必须是非负数' });
  }

  const now = new Date().toISOString();
  db.prepare('UPDATE system_config SET daily_interest_rate = ?, updated_at = ? WHERE id = ?')
    .run(dailyInterestRate, now, 'config');

  const config = db.prepare('SELECT * FROM system_config WHERE id = ?').get('config');
  res.json(formatConfig(config));
});

export default router;
