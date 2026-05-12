import { Router, Request, Response } from 'express';
import db from '../database';
import {
  HttpError,
  validateMaterialInput,
  getAvailableStock,
  getStock,
  nextAllocationStatus,
  nowDateTime,
  todayDate,
} from '../utils';
import { updateLedgerForInbound, updateLedgerForOutbound } from '../ledger';

const router = Router();

router.post('/request', (req: Request, res: Response) => {
  const {
    from_warehouse_id,
    to_warehouse_id,
    material_name,
    specification,
    category,
    requested_quantity,
    supplier,
  } = req.body as {
    from_warehouse_id?: number;
    to_warehouse_id?: number;
    material_name?: string;
    specification?: string;
    category?: string;
    requested_quantity?: number;
    supplier?: string;
  };

  const fromId = Number(from_warehouse_id);
  const toId = Number(to_warehouse_id);

  if (!Number.isFinite(fromId) || fromId <= 0 || !Number.isFinite(toId) || toId <= 0) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  const fromWarehouse = db
    .prepare("SELECT id, type FROM warehouses WHERE id = ? AND type = 'central'")
    .get(fromId);
  if (!fromWarehouse) {
    return res.status(404).json({ message: '来源中心库不存在' });
  }

  const toWarehouse = db
    .prepare("SELECT id, type FROM warehouses WHERE id = ? AND type = 'area'")
    .get(toId);
  if (!toWarehouse) {
    return res.status(404).json({ message: '目标片区库不存在' });
  }

  const name = String(material_name ?? '').trim();
  const spec = String(specification ?? '').trim();
  const cat = String(category ?? '').trim();
  const qty = Number(requested_quantity ?? 0);

  try {
    validateMaterialInput({ name, quantity: qty });
  } catch (err) {
    if (err instanceof HttpError) {
      return res.status(err.status).json({ message: err.message });
    }
    throw err;
  }

  const available = getAvailableStock(fromId, name, spec, cat);
  if (available - qty < 0) {
    return res.status(400).json({ message: '中心库可用库存不足以满足新的调拨申请（含在途扣减后为零或负数）' });
  }

  const info = db
    .prepare(
      `INSERT INTO allocations
       (from_warehouse_id, to_warehouse_id, material_name, specification, category, requested_quantity, status, supplier)
       VALUES (?, ?, ?, ?, ?, ?, 'pending_approval', ?)`
    )
    .run(fromId, toId, name, spec, cat, qty, String(supplier ?? '').trim());

  const allocation = db
    .prepare('SELECT * FROM allocations WHERE id = ?')
    .get(info.lastInsertRowid);

  res.status(201).json(allocation);
});

router.post('/:id/approve', (req: Request, res: Response) => {
  const allocationId = Number(req.params.id);
  if (!Number.isFinite(allocationId) || allocationId <= 0) {
    return res.status(400).json({ message: '无效的调拨单ID' });
  }

  const tx = db.transaction(() => {
    const allocation = db
      .prepare('SELECT * FROM allocations WHERE id = ?')
      .get(allocationId) as
      | {
          id: number;
          from_warehouse_id: number;
          to_warehouse_id: number;
          material_name: string;
          specification: string;
          category: string;
          requested_quantity: number;
          approved_quantity: number | null;
          status: string;
          supplier: string;
          created_at: string;
        }
      | undefined;

    if (!allocation) {
      return { status: 404, message: '调拨申请不存在' };
    }

    if (allocation.status !== 'pending_approval') {
      return { status: 400, message: '当前状态不允许审批' };
    }

    const available = getAvailableStock(
      allocation.from_warehouse_id,
      allocation.material_name,
      allocation.specification,
      allocation.category
    );

    const approved = Math.min(available, allocation.requested_quantity);
    if (approved < allocation.requested_quantity) {
      db.prepare(
        `UPDATE allocations
         SET status = 'approved', approved_quantity = ?
         WHERE id = ?`
      ).run(approved, allocationId);
    } else {
      db.prepare(
        `UPDATE allocations
         SET status = 'approved', approved_quantity = ?
         WHERE id = ?`
      ).run(approved, allocationId);
    }

    return {
      status: 200,
      data: db.prepare('SELECT * FROM allocations WHERE id = ?').get(allocationId),
      partial: approved < allocation.requested_quantity,
      available_quantity: available,
    };
  });

  const result = tx();
  if (result.status !== 200) {
    return res.status(result.status).json({ message: (result as { message: string }).message });
  }

  res.json({
    allocation: (result as { data: unknown }).data,
    partial: (result as { partial: boolean }).partial,
    available_quantity: (result as { available_quantity: number }).available_quantity,
  });
});

router.post('/:id/ship', (req: Request, res: Response) => {
  const allocationId = Number(req.params.id);
  if (!Number.isFinite(allocationId) || allocationId <= 0) {
    return res.status(400).json({ message: '无效的调拨单ID' });
  }

  const tx = db.transaction(() => {
    const allocation = db
      .prepare('SELECT * FROM allocations WHERE id = ?')
      .get(allocationId) as
      | {
          id: number;
          from_warehouse_id: number;
          to_warehouse_id: number;
          material_name: string;
          specification: string;
          category: string;
          requested_quantity: number;
          approved_quantity: number | null;
          status: string;
          supplier: string;
          created_at: string;
        }
      | undefined;

    if (!allocation) {
      return { status: 404, message: '调拨申请不存在' };
    }

    if (allocation.status !== 'approved') {
      return { status: 400, message: '当前状态不允许发货' };
    }

    if (allocation.approved_quantity === null || allocation.approved_quantity <= 0) {
      return { status: 400, message: '没有批准可发运的数量' };
    }

    const actualStock = getStock(
      allocation.from_warehouse_id,
      allocation.material_name,
      allocation.specification,
      allocation.category
    );

    if (actualStock < allocation.approved_quantity) {
      return { status: 409, message: '库存不足' };
    }

    const batches = db
      .prepare(
        `SELECT id, quantity, inbound_date, expiry_date
         FROM material_batches
         WHERE warehouse_id = ? AND name = ? AND specification = ? AND category = ?
           AND quantity > 0
         ORDER BY inbound_date ASC, id ASC`
      )
      .all(
        allocation.from_warehouse_id,
        allocation.material_name,
        allocation.specification,
        allocation.category
      ) as Array<{
      id: number;
      quantity: number;
      inbound_date: string;
      expiry_date: string | null;
    }>;

    let remaining = allocation.approved_quantity;
    const picked: Array<{
      batchId: number;
      quantity: number;
      inbound_date: string;
      expiry_date: string | null;
    }> = [];

    for (const batch of batches) {
      if (remaining <= 0) break;
      const take = Math.min(batch.quantity, remaining);
      picked.push({
        batchId: batch.id,
        quantity: take,
        inbound_date: batch.inbound_date,
        expiry_date: batch.expiry_date,
      });
      remaining -= take;
    }

    if (remaining > 0) {
      return { status: 409, message: '库存不足' };
    }

    let totalTransit = 0;
    for (const item of picked) {
      db.prepare('UPDATE material_batches SET quantity = quantity - ? WHERE id = ?').run(
        item.quantity,
        item.batchId
      );
      totalTransit += item.quantity;
    }

    for (const item of picked) {
      db.prepare(
        `INSERT INTO allocation_in_transit
         (allocation_id, warehouse_id, material_name, specification, category, quantity, supplier, inbound_date, expiry_date)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
      ).run(
        allocationId,
        allocation.from_warehouse_id,
        allocation.material_name,
        allocation.specification,
        allocation.category,
        item.quantity,
        allocation.supplier,
        item.inbound_date,
        item.expiry_date
      );
    }

    db.prepare("UPDATE allocations SET status = 'in_transit' WHERE id = ?").run(allocationId);

    updateLedgerForOutbound(
      allocation.from_warehouse_id,
      allocation.material_name,
      allocation.specification,
      allocation.category,
      totalTransit
    );

    return { status: 200, data: db.prepare('SELECT * FROM allocations WHERE id = ?').get(allocationId) };
  });

  const result = tx();
  if (result.status !== 200) {
    return res.status(result.status).json({ message: (result as { message: string }).message });
  }

  res.json((result as { data: unknown }).data);
});

router.post('/:id/arrive', (req: Request, res: Response) => {
  const allocationId = Number(req.params.id);
  if (!Number.isFinite(allocationId) || allocationId <= 0) {
    return res.status(400).json({ message: '无效的调拨单ID' });
  }

  const tx = db.transaction(() => {
    const allocation = db
      .prepare('SELECT * FROM allocations WHERE id = ?')
      .get(allocationId) as
      | {
          id: number;
          status: string;
        }
      | undefined;

    if (!allocation) {
      return { status: 404, message: '调拨申请不存在' };
    }

    if (allocation.status !== 'in_transit') {
      return { status: 400, message: '当前状态不允许标记到达' };
    }

    db.prepare("UPDATE allocations SET status = 'arrived' WHERE id = ?").run(allocationId);
    return { status: 200, data: db.prepare('SELECT * FROM allocations WHERE id = ?').get(allocationId) };
  });

  const result = tx();
  if (result.status !== 200) {
    return res.status(result.status).json({ message: (result as { message: string }).message });
  }

  res.json((result as { data: unknown }).data);
});

router.post('/:id/receive', (req: Request, res: Response) => {
  const allocationId = Number(req.params.id);
  if (!Number.isFinite(allocationId) || allocationId <= 0) {
    return res.status(400).json({ message: '无效的调拨单ID' });
  }

  const tx = db.transaction(() => {
    const allocation = db
      .prepare('SELECT * FROM allocations WHERE id = ?')
      .get(allocationId) as
      | {
          id: number;
          from_warehouse_id: number;
          to_warehouse_id: number;
          material_name: string;
          specification: string;
          category: string;
          approved_quantity: number | null;
          status: string;
          supplier: string;
        }
      | undefined;

    if (!allocation) {
      return { status: 404, message: '调拨申请不存在' };
    }

    if (allocation.status !== 'arrived') {
      return { status: 400, message: '当前状态不允许入库' };
    }

    const inTransitRows = db
      .prepare(
        `SELECT quantity, inbound_date, expiry_date, supplier
         FROM allocation_in_transit
         WHERE allocation_id = ?`
      )
      .all(allocationId) as Array<{
      quantity: number;
      inbound_date: string;
      expiry_date: string | null;
      supplier: string;
    }>;

    if (inTransitRows.length === 0) {
      return { status: 400, message: '没有在途物资可入库' };
    }

    let totalReceived = 0;
    for (const row of inTransitRows) {
      db.prepare(
        `INSERT INTO material_batches
         (warehouse_id, name, specification, category, quantity, inbound_date, expiry_date, supplier)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
      ).run(
        allocation.to_warehouse_id,
        allocation.material_name,
        allocation.specification,
        allocation.category,
        row.quantity,
        row.inbound_date,
        row.expiry_date,
        row.supplier
      );
      totalReceived += row.quantity;
    }

    db.prepare('DELETE FROM allocation_in_transit WHERE allocation_id = ?').run(allocationId);
    db.prepare("UPDATE allocations SET status = 'received' WHERE id = ?").run(allocationId);

    updateLedgerForInbound(
      allocation.to_warehouse_id,
      allocation.material_name,
      allocation.specification,
      allocation.category,
      totalReceived
    );

    return { status: 200, data: db.prepare('SELECT * FROM allocations WHERE id = ?').get(allocationId) };
  });

  const result = tx();
  if (result.status !== 200) {
    return res.status(result.status).json({ message: (result as { message: string }).message });
  }

  res.json((result as { data: unknown }).data);
});

router.get('/', (req: Request, res: Response) => {
  const { from_warehouse_id, to_warehouse_id, status } = req.query as Record<string, string>;

  const fromId = from_warehouse_id ? Number(from_warehouse_id) : null;
  const toId = to_warehouse_id ? Number(to_warehouse_id) : null;

  if (fromId !== null && (!Number.isFinite(fromId) || fromId <= 0)) {
    return res.status(400).json({ message: '无效的来源仓库ID' });
  }
  if (toId !== null && (!Number.isFinite(toId) || toId <= 0)) {
    return res.status(400).json({ message: '无效的目标仓库ID' });
  }

  let sql = 'SELECT * FROM allocations WHERE 1=1';
  const params: unknown[] = [];

  if (fromId !== null) {
    sql += ' AND from_warehouse_id = ?';
    params.push(fromId);
  }
  if (toId !== null) {
    sql += ' AND to_warehouse_id = ?';
    params.push(toId);
  }
  if (status) {
    sql += ' AND status = ?';
    params.push(String(status));
  }

  sql += ' ORDER BY created_at DESC, id DESC';

  const rows = db.prepare(sql).all(...params);
  res.json(rows);
});

router.get('/:id', (req: Request, res: Response) => {
  const allocationId = Number(req.params.id);
  if (!Number.isFinite(allocationId) || allocationId <= 0) {
    return res.status(400).json({ message: '无效的调拨单ID' });
  }

  const row = db.prepare('SELECT * FROM allocations WHERE id = ?').get(allocationId);
  if (!row) {
    return res.status(404).json({ message: '调拨申请不存在' });
  }

  res.json(row);
});

export default router;
