import { Router, Request, Response } from 'express';
import db from '../database';
import { todayDate, nowDateTime } from '../utils';
import { updateLedgerForInbound } from '../ledger';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const { warehouse_id } = req.query as Record<string, string>;
  const warehouseId = warehouse_id ? Number(warehouse_id) : null;

  if (warehouseId !== null && (!Number.isFinite(warehouseId) || warehouseId <= 0)) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  let sql = 'SELECT *, voided AS voided FROM issue_records WHERE 1=1';
  const params: unknown[] = [];

  if (warehouseId !== null) {
    sql += ' AND warehouse_id = ?';
    params.push(warehouseId);
  }

  sql += ' ORDER BY issued_at DESC, id DESC';

  const rows = db.prepare(sql).all(...params);
  res.json(rows);
});

router.post('/:id/void', (req: Request, res: Response) => {
  const issueId = Number(req.params.id);
  if (!Number.isFinite(issueId) || issueId <= 0) {
    return res.status(400).json({ message: '无效的发放记录ID' });
  }

  const tx = db.transaction(() => {
    const issue = db
      .prepare('SELECT * FROM issue_records WHERE id = ?')
      .get(issueId) as
      | {
          id: number;
          warehouse_id: number;
          material_name: string;
          specification: string;
          category: string;
          quantity: number;
          issued_at: string;
          voided: number;
          voided_at: string | null;
        }
      | undefined;

    if (!issue) {
      return { status: 404, message: '发放记录不存在' };
    }

    if (issue.voided) {
      return { status: 409, message: '发放记录已作废，不能再次作废' };
    }

    const lines = db
      .prepare('SELECT batch_id, quantity FROM issue_lines WHERE issue_id = ?')
      .all(issueId) as Array<{ batch_id: number; quantity: number }>;

    for (const line of lines) {
      const batch = db
        .prepare('SELECT id FROM material_batches WHERE id = ?')
        .get(line.batch_id);

      if (batch) {
        db.prepare('UPDATE material_batches SET quantity = quantity + ? WHERE id = ?').run(
          line.quantity,
          line.batch_id
        );
      } else {
        const originalBatch = db
          .prepare(
            `SELECT name, specification, category, supplier, inbound_date, expiry_date
             FROM material_batches
             WHERE id = ?`
          )
          .get(line.batch_id) as
          | {
              name: string;
              specification: string;
              category: string;
              supplier: string;
              inbound_date: string;
              expiry_date: string | null;
            }
          | undefined;

        if (originalBatch) {
          db.prepare(
            `INSERT INTO material_batches
             (warehouse_id, name, specification, category, quantity, inbound_date, expiry_date, supplier)
             VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
          ).run(
            issue.warehouse_id,
            originalBatch.name,
            originalBatch.specification,
            originalBatch.category,
            line.quantity,
            originalBatch.inbound_date,
            originalBatch.expiry_date,
            originalBatch.supplier
          );
        }
      }
    }

    db.prepare(
      `UPDATE issue_records
       SET voided = 1, voided_at = ?
       WHERE id = ?`
    ).run(nowDateTime(), issueId);

    updateLedgerForInbound(
      issue.warehouse_id,
      issue.material_name,
      issue.specification,
      issue.category,
      issue.quantity
    );

    return { status: 200, data: db.prepare('SELECT *, voided AS voided FROM issue_records WHERE id = ?').get(issueId) };
  });

  const result = tx();
  if (result.status !== 200) {
    return res.status(result.status).json({ message: (result as { message: string }).message });
  }

  res.json((result as { data: unknown }).data);
});

router.get('/summary/daily', (req: Request, res: Response) => {
  const { date, warehouse_id } = req.query as Record<string, string>;
  const targetDate = date ?? todayDate();
  const warehouseId = warehouse_id ? Number(warehouse_id) : null;

  if (warehouseId !== null && (!Number.isFinite(warehouseId) || warehouseId <= 0)) {
    return res.status(400).json({ message: '无效的仓库ID' });
  }

  let sql = `
    SELECT category, material_name, specification,
           COALESCE(SUM(CASE WHEN voided = 0 THEN quantity ELSE 0 END), 0) AS total_quantity
    FROM issue_records
    WHERE date(issued_at) = date(?)
  `;
  const params: unknown[] = [targetDate];

  if (warehouseId !== null) {
    sql += ' AND warehouse_id = ?';
    params.push(warehouseId);
  }

  sql += ' GROUP BY category, material_name, specification ORDER BY category, material_name';

  const rows = db.prepare(sql).all(...params);
  res.json({ date: targetDate, items: rows });
});

export default router;
