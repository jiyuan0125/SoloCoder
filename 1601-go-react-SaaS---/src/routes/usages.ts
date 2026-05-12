import { Router, Request, Response } from 'express';
import { usageService } from '../services/UsageService';

const router = Router();

router.post('/api-calls', (req: Request, res: Response) => {
  const { subscriptionId, count } = req.body;

  if (!subscriptionId) {
    return res.status(400).json({ error: '缺少订阅ID' });
  }

  const usage = usageService.recordApiCall(subscriptionId, count || 1);
  res.json(usage);
});

router.post('/storage', (req: Request, res: Response) => {
  const { subscriptionId, storageGB } = req.body;

  if (!subscriptionId || storageGB === undefined) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  const usage = usageService.recordStorage(subscriptionId, storageGB);
  res.json(usage);
});

router.get('/:subscriptionId/current', (req: Request, res: Response) => {
  const { subscriptionId } = req.params;
  const usage = usageService.getCurrentUsage(subscriptionId);
  res.json(usage || { apiCalls: 0, storageGB: 0 });
});

export default router;
