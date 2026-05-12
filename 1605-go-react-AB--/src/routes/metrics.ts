import { Router, Request, Response } from 'express';
import { recordMetricData, getMetricComparisons } from '../services/metricsService';

const router = Router();

router.post('/record', (req: Request, res: Response): void => {
  try {
    const { experimentId, metricId, variantId, userKey, value } = req.body;

    if (!experimentId || !metricId || !variantId || !userKey || value === undefined) {
      res.status(400).json({ error: 'Missing required fields' });
      return;
    }

    recordMetricData(experimentId, metricId, variantId, userKey, value);
    res.status(201).json({ success: true });
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.get('/comparisons/:experimentId', (req: Request, res: Response): void => {
  try {
    const comparisons = getMetricComparisons(req.params.experimentId);
    res.json(comparisons);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

export default router;
