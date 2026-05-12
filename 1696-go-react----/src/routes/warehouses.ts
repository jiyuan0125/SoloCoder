import { Router, Request, Response } from 'express';
import db from '../database';
import { HttpError } from '../utils';
import type { WarehouseType } from '../types';

const router = Router();

const isValidWarehouseType = (value: unknown): value is WarehouseType => {
  return value === 'central' || value === 'area';
};

router.post('/', (req: Request, res: Response) => {
  const { name, type } = req.body as { name?: string; type?: string };

  if (!name || String(name).trim() === '') {
    return res.status(400).json({ message: '仓库名称不能为空' });
  }

  if (!isValidWarehouseType(type)) {
    return res.status(400).json({ message: '仓库类型必须为 central 或 area' });
  }

  const existing = db
    .prepare('SELECT id FROM warehouses WHERE name = ?')
    .get(String(name).trim());

  if (existing) {
    return res.status(409).json({ message: '仓库名称重复' });
  }

  const info = db
    .prepare('INSERT INTO warehouses (name, type) VALUES (?, ?)')
    .run(String(name).trim(), type);

  const warehouse = db
    .prepare('SELECT * FROM warehouses WHERE id = ?')
    .get(info.lastInsertRowid);

  res.status(201).json(warehouse);
});

router.get('/', (req: Request, res: Response) => {
  const rows = db.prepare('SELECT * FROM warehouses ORDER BY created_at DESC').all();
  res.json(rows);
});

router.get('/:id', (req: Request, res: Response) => {
  const id = Number(req.params.id);
  if (!Number.isFinite(id) || id <= 0) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  const row = db.prepare('SELECT * FROM warehouses WHERE id = ?').get(id);
  if (!row) {
    return res.status(404).json({ message: '仓库不存在' });
  }

  res.json(row);
});

export default router;
