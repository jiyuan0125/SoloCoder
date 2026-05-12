import { Router, Request, Response } from 'express';
import db from '../db';
import { validateAmount, validateSellerLevel, generateId } from '../utils';
import { Seller, SellerLevel } from '../types';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { id, name, level } = req.body;
  const sellerId = id || generateId();
  if (!name || !level) {
    return res.status(400).json({ error: 'name and level are required' });
  }
  if (!validateSellerLevel(level)) {
    return res.status(404).json({ error: 'Invalid seller level' });
  }
  const existing = db.prepare('SELECT * FROM sellers WHERE id = ?').get(sellerId);
  if (existing) {
    return res.status(409).json({ error: 'Seller already exists' });
  }
  const now = Date.now();
  db.prepare('INSERT INTO sellers (id, name, level, created_at) VALUES (?, ?, ?, ?)').run(
    sellerId, name, level, now
  );
  const seller = db.prepare('SELECT * FROM sellers WHERE id = ?').get(sellerId) as Seller;
  res.status(201).json(seller);
});

router.get('/:id', (req: Request, res: Response) => {
  const seller = db.prepare('SELECT * FROM sellers WHERE id = ?').get(req.params.id);
  if (!seller) {
    return res.status(404).json({ error: 'Seller not found' });
  }
  res.json(seller);
});

router.get('/', (req: Request, res: Response) => {
  const sellers = db.prepare('SELECT * FROM sellers').all();
  res.json(sellers);
});

router.put('/:id/level', (req: Request, res: Response) => {
  const { level } = req.body;
  if (!level || !validateSellerLevel(level)) {
    return res.status(404).json({ error: 'Invalid seller level' });
  }
  const seller = db.prepare('SELECT * FROM sellers WHERE id = ?').get(req.params.id);
  if (!seller) {
    return res.status(404).json({ error: 'Seller not found' });
  }
  db.prepare('UPDATE sellers SET level = ? WHERE id = ?').run(level, req.params.id);
  const updated = db.prepare('SELECT * FROM sellers WHERE id = ?').get(req.params.id);
  res.json(updated);
});

export default router;
