import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Account, AccountType } from '../types';

const router = Router();

const formatAccount = (row: any): Account => ({
  id: row.id,
  name: row.name,
  bank: row.bank,
  balance: row.balance,
  type: row.type as AccountType,
  currency: row.currency,
  allowOverdraft: Boolean(row.allow_overdraft),
  createdAt: row.created_at,
  updatedAt: row.updated_at
});

router.get('/', (_req: Request, res: Response) => {
  const rows = db.prepare('SELECT * FROM accounts ORDER BY created_at DESC').all();
  res.json(rows.map(formatAccount));
});

router.get('/:id', (req: Request, res: Response) => {
  const row = db.prepare('SELECT * FROM accounts WHERE id = ?').get(req.params.id);
  if (!row) {
    return res.status(404).json({ error: '账户不存在' });
  }
  res.json(formatAccount(row));
});

router.post('/', (req: Request, res: Response) => {
  const { name, bank, balance = 0, type, allowOverdraft = false } = req.body;

  if (!name || !bank || !type) {
    return res.status(400).json({ error: '名称、银行和账户类型不能为空' });
  }

  if (!Object.values(AccountType).includes(type)) {
    return res.status(400).json({ error: '无效的账户类型' });
  }

  if (typeof balance !== 'number' || balance < 0 || !Number.isInteger(balance)) {
    return res.status(400).json({ error: '余额必须是非负整数' });
  }

  const existing = db.prepare('SELECT id FROM accounts WHERE name = ?').get(name);
  if (existing) {
    return res.status(409).json({ error: '账户名称已存在' });
  }

  const now = new Date().toISOString();
  const id = uuidv4();

  db.prepare(`
    INSERT INTO accounts (id, name, bank, balance, type, currency, allow_overdraft, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, 'CNY', ?, ?, ?)
  `).run(id, name, bank, balance, type, allowOverdraft ? 1 : 0, now, now);

  const account = db.prepare('SELECT * FROM accounts WHERE id = ?').get(id);
  res.status(201).json(formatAccount(account));
});

router.put('/:id', (req: Request, res: Response) => {
  const { name, bank } = req.body;

  const existing = db.prepare('SELECT * FROM accounts WHERE id = ?').get(req.params.id);
  if (!existing) {
    return res.status(404).json({ error: '账户不存在' });
  }

  if (name) {
    const duplicate = db.prepare('SELECT id FROM accounts WHERE name = ? AND id != ?').get(name, req.params.id);
    if (duplicate) {
      return res.status(409).json({ error: '账户名称已存在' });
    }
  }

  const now = new Date().toISOString();
  const updates: string[] = [];
  const params: any[] = [];

  if (name !== undefined) {
    updates.push('name = ?');
    params.push(name);
  }
  if (bank !== undefined) {
    updates.push('bank = ?');
    params.push(bank);
  }

  if (updates.length === 0) {
    return res.json(formatAccount(existing));
  }

  updates.push('updated_at = ?');
  params.push(now, req.params.id);

  db.prepare(`UPDATE accounts SET ${updates.join(', ')} WHERE id = ?`).run(...params);
  const updated = db.prepare('SELECT * FROM accounts WHERE id = ?').get(req.params.id);
  res.json(formatAccount(updated));
});

router.delete('/:id', (req: Request, res: Response) => {
  const existing = db.prepare('SELECT * FROM accounts WHERE id = ?').get(req.params.id);
  if (!existing) {
    return res.status(404).json({ error: '账户不存在' });
  }

  db.prepare('DELETE FROM accounts WHERE id = ?').run(req.params.id);
  res.status(204).send();
});

export default router;
