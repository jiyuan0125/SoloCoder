import { Router, Request, Response } from 'express';
import db from '../db';
import { validateAmount, generateId } from '../utils';
import { Transaction } from '../types';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { seller_id, type, amount, cycle } = req.body;
  if (!seller_id || !type || !amount || !cycle) {
    return res.status(400).json({ error: 'seller_id, type, amount and cycle are required' });
  }
  if (type !== 'receivable' && type !== 'refund') {
    return res.status(400).json({ error: 'type must be receivable or refund' });
  }
  if (!validateAmount(amount)) {
    return res.status(400).json({ error: 'amount must be non-negative integer (cents)' });
  }
  const seller = db.prepare('SELECT * FROM sellers WHERE id = ?').get(seller_id);
  if (!seller) {
    return res.status(404).json({ error: 'Seller not found' });
  }
  const id = generateId();
  const now = Date.now();
  db.prepare(
    'INSERT INTO transactions (id, seller_id, type, amount, cycle, created_at) VALUES (?, ?, ?, ?, ?, ?)'
  ).run(id, seller_id, type, amount, cycle, now);
  const tx = db.prepare('SELECT * FROM transactions WHERE id = ?').get(id) as Transaction;
  res.status(201).json(tx);
});

router.get('/seller/:sellerId', (req: Request, res: Response) => {
  const { sellerId } = req.params;
  const { cycle } = req.query;
  let query = 'SELECT * FROM transactions WHERE seller_id = ?';
  const params: any[] = [sellerId];
  if (cycle) {
    query += ' AND cycle = ?';
    params.push(cycle);
  }
  query += ' ORDER BY created_at DESC';
  const transactions = db.prepare(query).all(...params);
  res.json(transactions);
});

router.get('/:id', (req: Request, res: Response) => {
  const tx = db.prepare('SELECT * FROM transactions WHERE id = ?').get(req.params.id);
  if (!tx) {
    return res.status(404).json({ error: 'Transaction not found' });
  }
  res.json(tx);
});

export default router;
