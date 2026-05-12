import { Router, Request, Response } from 'express';
import db from '../db';

const router = Router();

router.get('/', (_req: Request, res: Response) => {
  const packages = db.prepare('SELECT * FROM packages ORDER BY created_at DESC').all();
  res.json(packages);
});

router.get('/:id', (req: Request, res: Response) => {
  const pkg = db.prepare('SELECT * FROM packages WHERE id = ?').get(Number(req.params.id));
  if (!pkg) {
    return res.status(404).json({ message: '套餐不存在' });
  }
  res.json(pkg);
});

router.post('/', (req: Request, res: Response) => {
  const { name, max_users, max_storage, max_api_calls } = req.body;

  if (!name || max_users === undefined || max_storage === undefined || max_api_calls === undefined) {
    return res.status(400).json({ message: '缺少必填字段' });
  }

  const existing = db.prepare('SELECT * FROM packages WHERE name = ?').get(name);
  if (existing) {
    return res.status(409).json({ message: '套餐名称已存在' });
  }

  const result = db.prepare(
    'INSERT INTO packages (name, max_users, max_storage, max_api_calls) VALUES (?, ?, ?, ?)'
  ).run(name, max_users, max_storage, max_api_calls);

  const pkg = db.prepare('SELECT * FROM packages WHERE id = ?').get(result.lastInsertRowid as number);
  res.status(201).json(pkg);
});

router.put('/:id', (req: Request, res: Response) => {
  const id = Number(req.params.id);
  const { name, max_users, max_storage, max_api_calls } = req.body;

  const existing = db.prepare('SELECT * FROM packages WHERE id = ?').get(id);
  if (!existing) {
    return res.status(404).json({ message: '套餐不存在' });
  }

  if (name) {
    const nameConflict = db.prepare('SELECT * FROM packages WHERE name = ? AND id != ?').get(name, id);
    if (nameConflict) {
      return res.status(409).json({ message: '套餐名称已存在' });
    }
  }

  db.prepare(
    'UPDATE packages SET name = COALESCE(?, name), max_users = COALESCE(?, max_users), max_storage = COALESCE(?, max_storage), max_api_calls = COALESCE(?, max_api_calls), updated_at = CURRENT_TIMESTAMP WHERE id = ?'
  ).run(name, max_users, max_storage, max_api_calls, id);

  const updated = db.prepare('SELECT * FROM packages WHERE id = ?').get(id);
  res.json(updated);
});

router.delete('/:id', (req: Request, res: Response) => {
  const id = Number(req.params.id);

  const existing = db.prepare('SELECT * FROM packages WHERE id = ?').get(id);
  if (!existing) {
    return res.status(404).json({ message: '套餐不存在' });
  }

  const tenantsUsing = db.prepare('SELECT COUNT(*) as count FROM tenants WHERE package_id = ?').get(id) as any;
  if (tenantsUsing.count > 0) {
    return res.status(400).json({ message: `该套餐仍有${tenantsUsing.count}个租户在使用，无法删除` });
  }

  db.prepare('DELETE FROM packages WHERE id = ?').run(id);
  res.json({ message: '套餐已删除' });
});

export default router;
