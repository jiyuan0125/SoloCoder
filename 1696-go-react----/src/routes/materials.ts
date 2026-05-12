import { Router, Request, Response } from 'express';
import db from '../database';
import { HttpError, validateMaterialInput, getStock, getAvailableStock } from '../utils';
import { updateLedgerForInbound, updateLedgerForOutbound } from '../ledger';
import { nowDateTime } from '../utils';
import { todayDate } from '../utils';

const router = Router();

router.post('/inbound', (req: Request, res: Response) => {
  const {
    warehouse_id,
    name,
    specification,
    category,
    quantity,
    inbound_date,
    expiry_date,
    supplier,
  } = req.body as {
    warehouse_id?: number;
    name?: string;
    specification?: string;
    category?: string;
    quantity?: number;
    inbound_date?: string;
    expiry_date?: string | null;
    supplier?: string;
  };

  const warehouseId = Number(warehouse_id);
  if (!Number.isFinite(warehouseId) || warehouseId <= 0) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  const warehouse = db
    .prepare('SELECT id FROM warehouses WHERE id = ?')
    .get(warehouseId);

  if (!warehouse) {
    return res.status(404).json({ message: '仓库不存在' });
  }

  try {
    validateMaterialInput({ name: String(name ?? ''), quantity: Number(quantity ?? 0) });
  } catch (err) {
    if (err instanceof HttpError) {
      return res.status(err.status).json({ message: err.message });
    }
    throw err;
  }

  const actualInbound = inbound_date ?? todayDate();

  const info = db
    .prepare(
      `INSERT INTO material_batches
       (warehouse_id, name, specification, category, quantity, inbound_date, expiry_date, supplier)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
    )
    .run(
      warehouseId,
      String(name).trim(),
      String(specification ?? '').trim(),
      String(category ?? '').trim(),
      Number(quantity),
      actualInbound,
      expiry_date && String(expiry_date).trim() !== '' ? String(expiry_date).trim() : null,
      String(supplier ?? '').trim()
    );

  updateLedgerForInbound(
    warehouseId,
    String(name).trim(),
    String(specification ?? '').trim(),
    String(category ?? '').trim(),
    Number(quantity)
  );

  const batch = db
    .prepare('SELECT * FROM material_batches WHERE id = ?')
    .get(info.lastInsertRowid);

  res.status(201).json(batch);
});

router.post('/issue', (req: Request, res: Response) => {
  const {
    warehouse_id,
    name,
    specification,
    category,
    quantity,
  } = req.body as {
    warehouse_id?: number;
    name?: string;
    specification?: string;
    category?: string;
    quantity?: number;
  };

  const warehouseId = Number(warehouse_id);
  if (!Number.isFinite(warehouseId) || warehouseId <= 0) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  const warehouse = db
    .prepare('SELECT id FROM warehouses WHERE id = ?')
    .get(warehouseId);

  if (!warehouse) {
    return res.status(404).json({ message: '仓库不存在' });
  }

  const materialName = String(name ?? '').trim();
  const materialSpec = String(specification ?? '').trim();
  const materialCategory = String(category ?? '').trim();
  const issueQuantity = Number(quantity ?? 0);

  try {
    validateMaterialInput({ name: materialName, quantity: issueQuantity });
  } catch (err) {
    if (err instanceof HttpError) {
      return res.status(err.status).json({ message: err.message });
    }
    throw err;
  }

  const available = getAvailableStock(
    warehouseId,
    materialName,
    materialSpec,
    materialCategory
  );

  if (available < issueQuantity) {
    return res.status(400).json({ message: '可用库存不足' });
  }

  const tx = db.transaction(() => {
    const batches = db
      .prepare(
        `SELECT id, quantity FROM material_batches
         WHERE warehouse_id = ? AND name = ? AND specification = ? AND category = ?
           AND quantity > 0
         ORDER BY inbound_date ASC, id ASC`
      )
      .all(warehouseId, materialName, materialSpec, materialCategory) as Array<{
      id: number;
      quantity: number;
    }>;

    let remaining = issueQuantity;
    const consumed: Array<{ batchId: number; quantity: number }> = [];

    for (const batch of batches) {
      if (remaining <= 0) break;
      const take = Math.min(batch.quantity, remaining);
      consumed.push({ batchId: batch.id, quantity: take });
      remaining -= take;
    }

    if (remaining > 0) {
      throw new Error('库存异常，实际库存不足');
    }

    const issuedAt = nowDateTime();
    const issueInfo = db
      .prepare(
        `INSERT INTO issue_records
         (warehouse_id, material_name, specification, category, quantity, issued_at)
         VALUES (?, ?, ?, ?, ?, ?)`
      )
      .run(
        warehouseId,
        materialName,
        materialSpec,
        materialCategory,
        issueQuantity,
        issuedAt
      );

    const issueId = Number(issueInfo.lastInsertRowid);

    for (const item of consumed) {
      db.prepare('INSERT INTO issue_lines (issue_id, batch_id, quantity) VALUES (?, ?, ?)').run(
        issueId,
        item.batchId,
        item.quantity
      );
      db.prepare('UPDATE material_batches SET quantity = quantity - ? WHERE id = ?').run(
        item.quantity,
        item.batchId
      );
    }

    updateLedgerForOutbound(
      warehouseId,
      materialName,
      materialSpec,
      materialCategory,
      issueQuantity
    );

    return issueId;
  });

  let issueId: number;
  try {
    issueId = tx();
  } catch (err) {
    if (err instanceof Error && err.message.includes('库存异常')) {
      return res.status(400).json({ message: err.message });
    }
    throw err;
  }

  const issue = db
    .prepare(
      `SELECT *, voided AS voided FROM issue_records WHERE id = ?`
    )
    .get(issueId);

  res.status(201).json(issue);
});

router.get('/stock', (req: Request, res: Response) => {
  const { warehouse_id, name, specification, category } = req.query as Record<string, string>;

  const warehouseId = warehouse_id ? Number(warehouse_id) : null;
  if (warehouseId !== null && (!Number.isFinite(warehouseId) || warehouseId <= 0)) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  let sql = `SELECT warehouse_id, name, specification, category, COALESCE(SUM(quantity), 0) AS total
             FROM material_batches`;
  const conditions: string[] = [];
  const params: unknown[] = [];

  if (warehouseId !== null) {
    conditions.push('warehouse_id = ?');
    params.push(warehouseId);
  }
  if (name) {
    conditions.push('name = ?');
    params.push(String(name));
  }
  if (specification) {
    conditions.push('specification = ?');
    params.push(String(specification));
  }
  if (category) {
    conditions.push('category = ?');
    params.push(String(category));
  }

  if (conditions.length > 0) {
    sql += ' WHERE ' + conditions.join(' AND ');
  }

  sql += ' GROUP BY warehouse_id, name, specification, category ORDER BY warehouse_id, name';

  const rows = db.prepare(sql).all(...params);
  res.json(rows);
});

router.get('/batches', (req: Request, res: Response) => {
  const { warehouse_id } = req.query as Record<string, string>;
  const warehouseId = warehouse_id ? Number(warehouse_id) : null;

  if (warehouseId !== null && (!Number.isFinite(warehouseId) || warehouseId <= 0)) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  let sql = 'SELECT * FROM material_batches WHERE 1=1';
  const params: unknown[] = [];

  if (warehouseId !== null) {
    sql += ' AND warehouse_id = ?';
    params.push(warehouseId);
  }

  sql += ' ORDER BY inbound_date ASC, id ASC';

  const rows = db.prepare(sql).all(...params);
  res.json(rows);
});

export default router;
