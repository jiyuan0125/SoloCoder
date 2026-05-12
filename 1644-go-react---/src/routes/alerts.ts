import { Router, Request, Response } from 'express';
import { AlertLevel, AlertStatus } from '../types';
import {
  createAlert,
  acknowledgeAlert,
  resolveAlert,
  getAlerts,
  getAlertById,
  findOrCreateAggregation,
  removeAggregation
} from '../services/alertService';
import { sendAggregatedAlert } from '../services/notificationService';

const router = Router();

const validLevels: AlertLevel[] = ['info', 'warning', 'critical', 'emergency'];
const validStatuses: AlertStatus[] = ['open', 'acknowledged', 'resolved'];

interface CreateAlertRequest {
  name: string;
  level: AlertLevel;
  sourceSystem: string;
  description?: string;
  metrics?: Record<string, any>;
}

router.post('/', async (req: Request, res: Response) => {
  try {
    const { name, level, sourceSystem, description = '', metrics = {} } = req.body as CreateAlertRequest;

    if (!name || !level || !sourceSystem) {
      return res.status(400).json({
        error: 'Missing required fields',
        message: 'name, level, and sourceSystem are required'
      });
    }

    if (!validLevels.includes(level)) {
      return res.status(400).json({
        error: 'Invalid level',
        message: 'Level must be one of: info, warning, critical, emergency'
      });
    }

    const alert = await createAlert(name, level, sourceSystem, description, metrics);
    await findOrCreateAggregation(name, level, sourceSystem, description, metrics);

    res.status(201).json(alert);
  } catch (err) {
    console.error('Failed to create alert:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/', async (req: Request, res: Response) => {
  try {
    const { level, status } = req.query;

    const levelParam = level ? (level as AlertLevel) : undefined;
    const statusParam = status ? (status as AlertStatus) : undefined;

    if (levelParam && !validLevels.includes(levelParam)) {
      return res.status(400).json({
        error: 'Invalid level',
        message: 'Level must be one of: info, warning, critical, emergency'
      });
    }

    if (statusParam && !validStatuses.includes(statusParam)) {
      return res.status(400).json({
        error: 'Invalid status',
        message: 'Status must be one of: open, acknowledged, resolved'
      });
    }

    const alerts = await getAlerts(levelParam, statusParam);
    res.json(alerts);
  } catch (err) {
    console.error('Failed to get alerts:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const alert = await getAlertById(id);

    if (!alert) {
      return res.status(404).json({ error: 'Alert not found' });
    }

    res.json(alert);
  } catch (err) {
    console.error('Failed to get alert:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id/acknowledge', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const alert = await acknowledgeAlert(id);

    if (!alert) {
      return res.status(404).json({ error: 'Alert not found' });
    }

    res.json(alert);
  } catch (err) {
    console.error('Failed to acknowledge alert:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id/resolve', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const alert = await resolveAlert(id);

    if (!alert) {
      return res.status(404).json({ error: 'Alert not found' });
    }

    res.json(alert);
  } catch (err) {
    console.error('Failed to resolve alert:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
