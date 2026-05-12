import { Router, Request, Response } from 'express';
import { subscriptionService } from '../services/subscriptionService';
import { PlanId } from '../types';

const router = Router();

router.get('/customer/:customerId', (req: Request, res: Response) => {
  const { customerId } = req.params;
  const subs = subscriptionService.getSubscriptionsByCustomer(customerId);
  res.json(subs);
});

router.post('/', (req: Request, res: Response) => {
  const { customerId, planId } = req.body;

  if (!customerId || !planId) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  const result = subscriptionService.createSubscription(customerId, planId as PlanId);

  if ('error' in result) {
    return res.status(400).json({ error: result.error });
  }

  res.status(201).json(result);
});

router.post('/:subscriptionId/upgrade', (req: Request, res: Response) => {
  const { subscriptionId } = req.params;
  const { planId } = req.body;

  if (!planId) {
    return res.status(400).json({ error: '缺少套餐ID' });
  }

  const result = subscriptionService.upgradePlan(subscriptionId, planId as PlanId);

  if ('error' in result) {
    return res.status(result.status).json({ error: result.error });
  }

  res.json(result);
});

router.delete('/:subscriptionId', (req: Request, res: Response) => {
  const { subscriptionId } = req.params;
  const result = subscriptionService.cancelSubscription(subscriptionId);

  if (!result.success) {
    return res.status(404).json({ error: result.message });
  }

  res.json({ success: true });
});

export default router;
