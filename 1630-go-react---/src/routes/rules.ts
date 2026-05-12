import { Router, Request, Response } from 'express';
import { getDb } from '../db';

const router = Router();
const VALID_OPERATORS = ['gt', 'lt', 'eq'];
const VALID_LEVELS = ['warning', 'critical', 'emergency'];
const VALID_METRICS = ['cpu_usage', 'memory_usage', 'qps', 'error_rate', 'avg_response_time'];

router.post('/', async (req: Request, res: Response) => {
  const { metric_name, operator, threshold, duration, level } = req.body;

  if (!metric_name || typeof metric_name !== 'string' || !VALID_METRICS.includes(metric_name)) {
    return res.status(400).json({ error: 'invalid or missing metric_name' });
  }
  if (!operator || typeof operator !== 'string' || !VALID_OPERATORS.includes(operator)) {
    return res.status(400).json({ error: 'invalid or missing operator' });
  }
  if (typeof threshold !== 'number' || isNaN(threshold)) {
    return res.status(400).json({ error: 'threshold must be a number' });
  }
  if (typeof duration !== 'number' || isNaN(duration) || duration < 0) {
    return res.status(400).json({ error: 'duration must be a non-negative number' });
  }
  if (!level || typeof level !== 'string' || !VALID_LEVELS.includes(level)) {
    return res.status(400).json({ error: 'invalid or missing level' });
  }

  const db = getDb();

  try {
    const result = await db.run(
      'INSERT INTO alert_rules (metric_name, operator, threshold, duration, level) VALUES (?, ?, ?, ?, ?)',
      [metric_name, operator, threshold, duration, level]
    );
    res.status(201).json({
      id: result.lastID,
      metric_name,
      operator,
      threshold,
      duration,
      level
    });
  } catch (e) {
    console.error(e);
    res.status(500).json({ error: 'internal server error' });
  }
});

router.get('/', async (_req: Request, res: Response) => {
  const db = getDb();
  try {
    const rules = await db.all('SELECT * FROM alert_rules');
    res.json(rules);
  } catch (e) {
    console.error(e);
    res.status(500).json({ error: 'internal server error' });
  }
});

router.delete('/:id', async (req: Request, res: Response) => {
  const id = Number(req.params.id);

  if (isNaN(id)) {
    return res.status(400).json({ error: 'invalid rule id' });
  }

  const db = getDb();

  try {
    const pendingAlerts = await db.get(
      'SELECT * FROM alerts WHERE rule_id = ? AND status = ?',
      [id, 'pending']
    );

    if (pendingAlerts) {
      return res.status(400).json({ error: 'rule has active alerts' });
    }

    await db.run('DELETE FROM alert_rules WHERE id = ?', [id]);
    res.status(204).send();
  } catch (e) {
    console.error(e);
    res.status(500).json({ error: 'internal server error' });
  }
});

router.post('/alerts/:id/status', async (req: Request, res: Response) => {
  const id = Number(req.params.id);
  const { status } = req.body;

  if (isNaN(id)) {
    return res.status(400).json({ error: 'invalid alert id' });
  }

  const validTransitions: Record<string, string[]> = {
    pending: ['processing'],
    processing: ['recovered'],
    recovered: []
  };

  if (!status || typeof status !== 'string' || !Object.keys(validTransitions).includes(status)) {
    return res.status(400).json({ error: 'invalid status' });
  }

  const db = getDb();

  try {
    const alert = await db.get('SELECT * FROM alerts WHERE id = ?', [id]);
    if (!alert) {
      return res.status(404).json({ error: 'alert not found' });
    }

    const currentStatus = alert.status;
    const allowed = validTransitions[currentStatus] || [];

    if (!allowed.includes(status)) {
      return res.status(400).json({ error: `invalid state transition from ${currentStatus} to ${status}` });
    }

    await db.run('UPDATE alerts SET status = ? WHERE id = ?', [status, id]);
    res.json({ id, status });
  } catch (e) {
    console.error(e);
    res.status(500).json({ error: 'internal server error' });
  }
});

export default router;
