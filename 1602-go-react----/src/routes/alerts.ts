import { Router, Request, Response } from 'express';
import db from '../db';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const { handled, tenant_id } = req.query;
  
  const params: Array<string | number> = [];
  const conditions: string[] = [];

  if (handled !== undefined) {
    conditions.push('handled = ?');
    params.push(handled === 'true' ? 1 : 0);
  }

  if (tenant_id !== undefined) {
    conditions.push('tenant_id = ?');
    params.push(Number(tenant_id));
  }

  let query = 'SELECT * FROM alerts';
  if (conditions.length > 0) {
    query += ' WHERE ' + conditions.join(' AND ');
  }
  query += ' ORDER BY created_at DESC';

  const stmt = db.prepare(query);
  const alerts = params.length > 0 ? stmt.all(...params) : stmt.all();
  res.json(alerts);
});

router.get('/:id', (req: Request, res: Response) => {
  const alert = db.prepare('SELECT * FROM alerts WHERE id = ?').get(Number(req.params.id));
  if (!alert) {
    return res.status(404).json({ message: '告警不存在' });
  }
  res.json(alert);
});

router.put('/:id/handle', (req: Request, res: Response) => {
  const id = Number(req.params.id);
  
  const existing = db.prepare('SELECT * FROM alerts WHERE id = ?').get(id);
  if (!existing) {
    return res.status(404).json({ message: '告警不存在' });
  }

  db.prepare('UPDATE alerts SET handled = 1 WHERE id = ?').run(id);
  
  const updated = db.prepare('SELECT * FROM alerts WHERE id = ?').get(id);
  res.json(updated);
});

router.delete('/:id', (req: Request, res: Response) => {
  const id = Number(req.params.id);
  
  const existing = db.prepare('SELECT * FROM alerts WHERE id = ?').get(id);
  if (!existing) {
    return res.status(404).json({ message: '告警不存在' });
  }

  db.prepare('DELETE FROM alerts WHERE id = ?').run(id);
  res.json({ message: '告警已删除' });
});

export default router;
