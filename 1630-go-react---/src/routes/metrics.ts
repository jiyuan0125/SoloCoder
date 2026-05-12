import { Router, Request, Response } from 'express';
import { getDb } from '../db';
import { evaluateThreshold, checkAlertRules } from '../alerts';

const router = Router();
const VALID_METRICS = ['cpu_usage', 'memory_usage', 'qps', 'error_rate', 'avg_response_time'];

router.post('/services/:name/metrics', async (req: Request, res: Response) => {
  const { name } = req.params;
  const { metric_name, value, timestamp } = req.body;

  if (!metric_name || typeof metric_name !== 'string') {
    return res.status(400).json({ error: 'metric_name is required' });
  }
  if (!VALID_METRICS.includes(metric_name)) {
    return res.status(400).json({ error: 'invalid metric name' });
  }
  if (typeof value !== 'number' || isNaN(value)) {
    return res.status(400).json({ error: 'value must be a valid number' });
  }
  if (typeof timestamp !== 'number' || isNaN(timestamp)) {
    return res.status(400).json({ error: 'timestamp must be a valid number' });
  }

  const db = getDb();

  try {
    const service = await db.get('SELECT name FROM services WHERE name = ?', [name]);
    if (!service) {
      return res.status(404).json({ error: 'service not found' });
    }

    await db.run(
      'INSERT INTO metrics (service_name, metric_name, value, timestamp) VALUES (?, ?, ?, ?)',
      [name, metric_name, value, timestamp]
    );

    await checkAlertRules(name, metric_name, value, timestamp);
    res.status(201).json({ service_name: name, metric_name, value, timestamp});
  } catch (e: any) {
    if (e.code === 'SQLITE_CONSTRAINT_UNIQUE' || e.message?.includes('UNIQUE')) {
      res.status(409).json({ error: 'duplicate metric report' });
    } else {
      console.error(e);
      res.status(500).json({ error: 'internal server error' });
    }
  }
});

router.get('/metrics', async (req: Request, res: Response) => {
  const { service_name, start_time, end_time } = req.query;

  if (!service_name || typeof service_name !== 'string') {
    return res.status(400).json({ error: 'service_name is required' });
  }
  if (!start_time || typeof start_time !== 'string' || isNaN(Number(start_time))) {
    return res.status(400).json({ error: 'start_time is required' });
  }
  if (!end_time || typeof end_time !== 'string' || isNaN(Number(end_time))) {
    return res.status(400).json({ error: 'end_time is required' });
  }

  const db = getDb();

  try {
    const metrics = await db.all(
      'SELECT * FROM metrics WHERE service_name = ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp ASC',
      [service_name, Number(start_time), Number(end_time)]
    );
    res.json(metrics);
  } catch (e) {
    console.error(e);
    res.status(500).json({ error: 'internal server error' });
  }
});

export default router;
