import { Router, Request, Response } from 'express';
import db from '../database';
import { todayDate, nowDateTime } from '../utils';
import { updateLedgerForInbound } from '../ledger';

const router = Router();

router.get('/daily', (req: Request, res: Response) => {
  const { date, warehouse_id } = req.query as Record<string, string>;
  const targetDate = date ?? todayDate();
  const warehouseId = warehouse_id ? Number(warehouse_id) : null;

  if (warehouseId !== null && (!Number.isFinite(warehouseId) || warehouseId <= 0)) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  let sql = 'SELECT * FROM daily_ledger WHERE ledger_date = date(?)';
  const params: unknown[] = [targetDate];

  if (warehouseId !== null) {
    sql += ' AND warehouse_id = ?';
    params.push(warehouseId);
  }

  sql += ' ORDER BY warehouse_id, category, material_name';

  const rows = db.prepare(sql).all(...params);
  res.json({ date: targetDate, items: rows });
});

router.post('/inventory-check', (req: Request, res: Response) => {
  const {
    warehouse_id,
    items,
  } = req.body as {
    warehouse_id?: number;
    items?: Array<{
      name: string;
      specification?: string;
      category?: string;
      actual_quantity: number;
    }>;
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

  if (!Array.isArray(items)) {
    return res.status(400).json({ message: '盘点明细必须为数组' });
  }

  const differences: Array<{
    id: number;
    warehouse_id: number;
    material_name: string;
    specification: string;
    category: string;
    expected_quantity: number;
    actual_quantity: number;
    difference_quantity: number;
    recorded_at: string;
  }> = [];
  const tasks: Array<{
    id: number;
    warehouse_id: number;
    material_name: string;
    specification: string;
    category: string;
    expected_quantity: number;
    actual_quantity: number;
    difference_quantity: number;
    created_at: string;
    resolved: boolean;
  }> = [];

  const tx = db.transaction(() => {
    for (const item of items) {
      if (!item || !item.name || String(item.name).trim() === '') {
        continue;
      }
      const name = String(item.name).trim();
      const specification = String(item.specification ?? '').trim();
      const category = String(item.category ?? '').trim();
      const actual = Number(item.actual_quantity ?? 0);

      const row = db
        .prepare(
          `SELECT COALESCE(SUM(quantity), 0) AS total FROM material_batches
           WHERE warehouse_id = ? AND name = ? AND specification = ? AND category = ?`
        )
        .get(warehouseId, name, specification, category) as { total: number };

      const expected = row.total;
      const difference = actual - expected;

      if (difference !== 0) {
        const info = db
          .prepare(
            `INSERT INTO inventory_differences
             (warehouse_id, material_name, specification, category, expected_quantity, actual_quantity, difference_quantity)
             VALUES (?, ?, ?, ?, ?, ?, ?)`
          )
          .run(warehouseId, name, specification, category, expected, actual, difference);

        const diff = db
          .prepare('SELECT * FROM inventory_differences WHERE id = ?')
          .get(info.lastInsertRowid) as typeof differences[0];

        differences.push(diff);

        const taskInfo = db
          .prepare(
            `INSERT INTO check_tasks
             (warehouse_id, material_name, specification, category, expected_quantity, actual_quantity, difference_quantity)
             VALUES (?, ?, ?, ?, ?, ?, ?)`
          )
          .run(warehouseId, name, specification, category, expected, actual, difference);

        const task = db
          .prepare('SELECT * FROM check_tasks WHERE id = ?')
          .get(taskInfo.lastInsertRowid) as typeof tasks[0];

        tasks.push(task);
      }
    }
  });

  tx();

  res.json({
    warehouse_id: warehouseId,
    recorded_at: nowDateTime(),
    differences,
    check_tasks: tasks,
  });
});

router.get('/differences', (req: Request, res: Response) => {
  const { warehouse_id } = req.query as Record<string, string>;
  const warehouseId = warehouse_id ? Number(warehouse_id) : null;

  if (warehouseId !== null && (!Number.isFinite(warehouseId) || warehouseId <= 0)) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  let sql = 'SELECT * FROM inventory_differences WHERE 1=1';
  const params: unknown[] = [];

  if (warehouseId !== null) {
    sql += ' AND warehouse_id = ?';
    params.push(warehouseId);
  }

  sql += ' ORDER BY recorded_at DESC, id DESC';

  const rows = db.prepare(sql).all(...params);
  res.json(rows);
});

router.get('/check-tasks', (req: Request, res: Response) => {
  const { warehouse_id, resolved } = req.query as Record<string, string>;
  const warehouseId = warehouse_id ? Number(warehouse_id) : null;
  const resolvedFilter =
    resolved !== undefined ? (String(resolved) === 'true' ? 1 : 0) : null;

  if (warehouseId !== null && (!Number.isFinite(warehouseId) || warehouseId <= 0)) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  let sql = 'SELECT *, resolved AS resolved FROM check_tasks WHERE 1=1';
  const params: unknown[] = [];

  if (warehouseId !== null) {
    sql += ' AND warehouse_id = ?';
    params.push(warehouseId);
  }
  if (resolvedFilter !== null) {
    sql += ' AND resolved = ?';
    params.push(resolvedFilter);
  }

  sql += ' ORDER BY created_at DESC, id DESC';

  const rows = db.prepare(sql).all(...params);
  res.json(rows);
});

router.post('/check-tasks/:id/resolve', (req: Request, res: Response) => {
  const taskId = Number(req.params.id);
  if (!Number.isFinite(taskId) || taskId <= 0) {
    return res.status(400).json({ message: '无效的核查任务ID' });
  }

  const tx = db.transaction(() => {
    const task = db
      .prepare('SELECT * FROM check_tasks WHERE id = ?')
      .get(taskId) as
      | {
          id: number;
          warehouse_id: number;
          material_name: string;
          specification: string;
          category: string;
          expected_quantity: number;
          actual_quantity: number;
          difference_quantity: number;
          created_at: string;
          resolved: number;
        }
      | undefined;

    if (!task) {
      return { status: 404, message: '核查任务不存在' };
    }

    if (task.resolved) {
      return { status: 400, message: '核查任务已处理' };
    }

    db.prepare('UPDATE check_tasks SET resolved = 1 WHERE id = ?').run(taskId);
    return {
      status: 200,
      data: db
        .prepare('SELECT *, resolved AS resolved FROM check_tasks WHERE id = ?')
        .get(taskId),
    };
  });

  const result = tx();
  if (result.status !== 200) {
    return res.status(result.status).json({ message: (result as { message: string }).message });
  }

  res.json((result as { data: unknown }).data);
});

export default router;
