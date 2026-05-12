import { Router, Request, Response } from 'express';
import db from '../db';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const { seller_id, cycle } = req.query;
  let query = 'SELECT * FROM fee_records WHERE 1=1';
  const params: any[] = [];
  if (seller_id) {
    query += ' AND seller_id = ?';
    params.push(seller_id);
  }
  if (cycle) {
    query += ' AND cycle = ?';
    params.push(cycle);
  }
  query += ' ORDER BY created_at DESC';
  const records = db.prepare(query).all(...params);
  res.json(records);
});

router.get('/settlement/:settlementId', (req: Request, res: Response) => {
  const record = db.prepare('SELECT * FROM fee_records WHERE settlement_id = ?').get(req.params.settlementId);
  if (!record) {
    return res.status(404).json({ error: 'Fee record not found' });
  }
  res.json(record);
});

export default router;
