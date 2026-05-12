import { Router, Request, Response } from 'express';
import db from '../db';
import { validateAmount, generateId, calculateFee } from '../utils';
import { Seller, SellerLevel, SETTLEMENT_STATUS, Settlement } from '../types';

const router = Router();

router.post('/generate', (req: Request, res: Response) => {
  const { seller_id, cycle } = req.body;
  if (!seller_id || !cycle) {
    return res.status(400).json({ error: 'seller_id and cycle are required' });
  }
  const seller = db.prepare('SELECT * FROM sellers WHERE id = ?').get(seller_id) as Seller;
  if (!seller) {
    return res.status(404).json({ error: 'Seller not found' });
  }
  const existing = db.prepare(
    'SELECT * FROM settlements WHERE seller_id = ? AND cycle = ?'
  ).get(seller_id, cycle);
  if (existing) {
    return res.status(409).json({ error: 'Settlement already exists for this seller and cycle' });
  }
  const transactions = db.prepare(
    "SELECT * FROM transactions WHERE seller_id = ? AND cycle = ?"
  ).all(seller_id, cycle);
  let totalReceivable = 0;
  let totalRefund = 0;
  for (const tx of transactions) {
    if ((tx as any).type === 'receivable') {
      totalReceivable += (tx as any).amount;
    } else if ((tx as any).type === 'refund') {
      totalRefund += (tx as any).amount;
    }
  }
  const netAmount = totalReceivable - totalRefund;
  if (netAmount < 0) {
    return res.status(400).json({ error: 'Net amount cannot be negative' });
  }
  const fee = calculateFee(netAmount, seller.level as SellerLevel);
  const payableAmount = netAmount - fee;
  if (payableAmount < 0) {
    return res.status(400).json({ error: 'Payable amount cannot be negative' });
  }
  const id = generateId();
  const now = Date.now();
  db.prepare(
    `INSERT INTO settlements (id, seller_id, cycle, total_receivable, total_refund, net_amount, fee, payable_amount, status, created_at, confirmed_at, paid_at)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
  ).run(
    id, seller_id, cycle, totalReceivable, totalRefund, netAmount, fee, payableAmount,
    SETTLEMENT_STATUS.PENDING, now, null, null
  );
  const settlement = db.prepare('SELECT * FROM settlements WHERE id = ?').get(id);
  res.status(201).json(settlement);
});

router.post('/:id/confirm', (req: Request, res: Response) => {
  const { id } = req.params;
  const settlement = db.prepare('SELECT * FROM settlements WHERE id = ?').get(id) as any;
  if (!settlement) {
    return res.status(404).json({ error: 'Settlement not found' });
  }
  if (settlement.status !== SETTLEMENT_STATUS.PENDING) {
    return res.status(409).json({ error: 'Only pending settlements can be confirmed' });
  }
  const now = Date.now();
  const feeRecordId = generateId();
  const tx = db.transaction(() => {
    db.prepare(
      `INSERT INTO fee_records (id, settlement_id, seller_id, cycle, amount, created_at)
       VALUES (?, ?, ?, ?, ?, ?)`
    ).run(
      feeRecordId, settlement.id, settlement.seller_id, settlement.cycle, settlement.fee, now
    );
    db.prepare(
      'UPDATE settlements SET status = ?, confirmed_at = ? WHERE id = ?'
    ).run(SETTLEMENT_STATUS.CONFIRMED, now, id);
  });
  try {
    tx();
  } catch (error) {
    console.error('Transaction error:', error);
    return res.status(500).json({ error: 'Failed to confirm settlement' });
  }
  const updated = db.prepare('SELECT * FROM settlements WHERE id = ?').get(id);
  res.json(updated);
});

router.post('/:id/pay', (req: Request, res: Response) => {
  const { id } = req.params;
  const settlement = db.prepare('SELECT * FROM settlements WHERE id = ?').get(id) as any;
  if (!settlement) {
    return res.status(404).json({ error: 'Settlement not found' });
  }
  if (settlement.status !== SETTLEMENT_STATUS.CONFIRMED) {
    return res.status(409).json({ error: 'Only confirmed settlements can be paid' });
  }
  const now = Date.now();
  db.prepare(
    'UPDATE settlements SET status = ?, paid_at = ? WHERE id = ?'
  ).run(SETTLEMENT_STATUS.PAID, now, id);
  const updated = db.prepare('SELECT * FROM settlements WHERE id = ?').get(id);
  res.json(updated);
});

router.put('/:id', (req: Request, res: Response) => {
  const { id } = req.params;
  const { total_receivable, total_refund } = req.body;
  const settlement = db.prepare('SELECT * FROM settlements WHERE id = ?').get(id) as any;
  if (!settlement) {
    return res.status(404).json({ error: 'Settlement not found' });
  }
  if (settlement.status !== SETTLEMENT_STATUS.PENDING) {
    return res.status(409).json({ error: 'Only pending settlements can be modified' });
  }
  const newReceivable = total_receivable !== undefined ? total_receivable : settlement.total_receivable;
  const newRefund = total_refund !== undefined ? total_refund : settlement.total_refund;
  if (!validateAmount(newReceivable) || !validateAmount(newRefund)) {
    return res.status(400).json({ error: 'Amounts must be non-negative integers' });
  }
  const netAmount = newReceivable - newRefund;
  if (netAmount < 0) {
    return res.status(400).json({ error: 'Net amount cannot be negative' });
  }
  const seller = db.prepare('SELECT * FROM sellers WHERE id = ?').get(settlement.seller_id) as Seller;
  const fee = calculateFee(netAmount, seller.level as SellerLevel);
  const payableAmount = netAmount - fee;
  if (payableAmount < 0) {
    return res.status(400).json({ error: 'Payable amount cannot be negative' });
  }
  db.prepare(
    `UPDATE settlements SET total_receivable = ?, total_refund = ?, net_amount = ?, fee = ?, payable_amount = ?
     WHERE id = ?`
  ).run(newReceivable, newRefund, netAmount, fee, payableAmount, id);
  const updated = db.prepare('SELECT * FROM settlements WHERE id = ?').get(id);
  res.json(updated);
});

router.get('/:id', (req: Request, res: Response) => {
  const settlement = db.prepare('SELECT * FROM settlements WHERE id = ?').get(req.params.id);
  if (!settlement) {
    return res.status(404).json({ error: 'Settlement not found' });
  }
  res.json(settlement);
});

router.get('/seller/:sellerId', (req: Request, res: Response) => {
  const { sellerId } = req.params;
  const { cycle, status } = req.query;
  let query = 'SELECT * FROM settlements WHERE seller_id = ?';
  const params: any[] = [sellerId];
  if (cycle) {
    query += ' AND cycle = ?';
    params.push(cycle);
  }
  if (status) {
    query += ' AND status = ?';
    params.push(status);
  }
  query += ' ORDER BY created_at DESC';
  const settlements = db.prepare(query).all(...params);
  res.json(settlements);
});

router.get('/', (req: Request, res: Response) => {
  const { status, cycle } = req.query;
  let query = 'SELECT * FROM settlements WHERE 1=1';
  const params: any[] = [];
  if (status) {
    query += ' AND status = ?';
    params.push(status);
  }
  if (cycle) {
    query += ' AND cycle = ?';
    params.push(cycle);
  }
  query += ' ORDER BY created_at DESC';
  const settlements = db.prepare(query).all(...params);
  res.json(settlements);
});

export default router;
