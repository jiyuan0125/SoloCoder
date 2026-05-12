import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Transfer, TransferStatus } from '../types';

const router = Router();

const formatTransfer = (row: any): Transfer => ({
  id: row.id,
  fromAccountId: row.from_account_id,
  toAccountId: row.to_account_id,
  amount: row.amount,
  reason: row.reason,
  status: row.status as TransferStatus,
  operator: row.operator,
  createdAt: row.created_at,
  completedAt: row.completed_at,
  failureReason: row.failure_reason
});

const waitForLock = async (): Promise<void> => {
  return new Promise((resolve) => {
    const checkLock = () => {
      const lock = db.prepare('SELECT locked FROM locks WHERE id = ?').get('interest_calc') as { locked: number } | undefined;
      if (!lock || lock.locked === 0) {
        resolve();
      } else {
        setTimeout(checkLock, 100);
      }
    };
    checkLock();
  });
};

router.get('/', (_req: Request, res: Response) => {
  const rows = db.prepare('SELECT * FROM transfers ORDER BY created_at DESC').all();
  res.json(rows.map(formatTransfer));
});

router.get('/:id', (req: Request, res: Response) => {
  const row = db.prepare('SELECT * FROM transfers WHERE id = ?').get(req.params.id);
  if (!row) {
    return res.status(404).json({ error: '调拨记录不存在' });
  }
  res.json(formatTransfer(row));
});

router.post('/', async (req: Request, res: Response) => {
  const { id, fromAccountId, toAccountId, amount, reason, operator } = req.body;

  if (!fromAccountId || !toAccountId || !operator) {
    return res.status(400).json({ error: '调出账户、调入账户和操作人不能为空' });
  }

  if (typeof amount !== 'number' || amount <= 0 || !Number.isInteger(amount)) {
    return res.status(400).json({ error: '金额必须是正整数' });
  }

  if (fromAccountId === toAccountId) {
    return res.status(400).json({ error: '调出账户和调入账户不能相同' });
  }

  if (id) {
    const existingTransfer = db.prepare('SELECT id FROM transfers WHERE id = ?').get(id);
    if (existingTransfer) {
      return res.status(409).json({ error: '该调拨已提交' });
    }
  }

  const transferId = id || uuidv4();
  const now = new Date().toISOString();

  await waitForLock();

  const tx = db.transaction((): { success: boolean; data?: any; error?: string; code?: number } => {
    const fromAccount = db.prepare('SELECT * FROM accounts WHERE id = ?').get(fromAccountId) as { allow_overdraft: number; balance: number } | undefined;
    if (!fromAccount) {
      return { success: false, error: '调出账户不存在', code: 404 };
    }

    const toAccount = db.prepare('SELECT * FROM accounts WHERE id = ?').get(toAccountId);
    if (!toAccount) {
      return { success: false, error: '调入账户不存在', code: 404 };
    }

    if (!fromAccount.allow_overdraft && fromAccount.balance < amount) {
      db.prepare(`
        INSERT INTO transfers (id, from_account_id, to_account_id, amount, reason, status, operator, created_at, failure_reason)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
      `).run(transferId, fromAccountId, toAccountId, amount, reason, TransferStatus.FAILED, operator, now, '余额不足');
      
      return { success: false, error: '余额不足', code: 400 };
    }

    db.prepare('UPDATE accounts SET balance = balance - ?, updated_at = ? WHERE id = ?').run(amount, now, fromAccountId);

    db.prepare('UPDATE accounts SET balance = balance + ?, updated_at = ? WHERE id = ?').run(amount, now, toAccountId);

    db.prepare(`
      INSERT INTO transfers (id, from_account_id, to_account_id, amount, reason, status, operator, created_at, completed_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `).run(transferId, fromAccountId, toAccountId, amount, reason, TransferStatus.SUCCESS, operator, now, now);

    const transfer = db.prepare('SELECT * FROM transfers WHERE id = ?').get(transferId);
    return { success: true, data: transfer };
  });

  const result = tx();

  if (!result.success) {
    return res.status(result.code || 400).json({ error: result.error });
  }

  res.status(201).json(formatTransfer(result.data));
});

export default router;
