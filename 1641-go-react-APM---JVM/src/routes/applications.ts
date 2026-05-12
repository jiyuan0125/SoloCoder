import { Router, Request, Response } from 'express';
import {
  createApplication,
  applicationExistsByName,
  getApplicationById,
  addMetrics,
  updateApplicationLastReport,
  addTraces,
  getAlertsByApplication,
} from '../storage';
import { MetricsPayload } from '../types';
import { runAnomalyDetection, checkAlertRules } from '../services/anomalyDetection';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { name, description } = req.body;
  
  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: 'Name is required and must be a string' });
  }
  
  if (applicationExistsByName(name)) {
    return res.status(409).json({ error: 'Application with this name already exists' });
  }
  
  const app = createApplication(name, description || '');
  res.status(201).json(app);
});

router.post('/:appId/metrics', (req: Request, res: Response) => {
  const { appId } = req.params;
  const payload = req.body as MetricsPayload;
  
  const app = getApplicationById(appId);
  if (!app) {
    return res.status(404).json({ error: 'Application not found' });
  }
  
  if (!payload.http || !payload.database || !payload.jvm) {
    return res.status(400).json({ error: 'Missing required metrics sections: http, database, jvm' });
  }
  
  const now = Date.now();
  const aggregated = addMetrics(
    appId,
    now,
    payload.http,
    payload.database,
    payload.jvm,
    payload.custom || {}
  );
  
  updateApplicationLastReport(appId);
  
  if (payload.traces && payload.traces.length > 0) {
    addTraces(payload.traces);
  }
  
  setImmediate(() => {
    try {
      runAnomalyDetection(appId);
      checkAlertRules(appId, aggregated);
    } catch (err) {
      console.error('Error in anomaly detection:', err);
    }
  });
  
  res.status(200).json({ status: 'ok', timestamp: now });
});

router.get('/:appId/alerts', (req: Request, res: Response) => {
  const { appId } = req.params;
  
  const app = getApplicationById(appId);
  if (!app) {
    return res.status(404).json({ error: 'Application not found' });
  }
  
  const alerts = getAlertsByApplication(appId);
  res.status(200).json(alerts);
});

export default router;
