import { Router, Request, Response } from 'express';
import { z } from 'zod';
import { 
  createPlan, 
  getPlan, 
  getPlanWithStats, 
  updatePlanStatus, 
  createPlanSchema,
  selectBestPlan,
  recordImpression,
  recordClick,
  recordConversion,
  checkBudgetAndPause,
} from './services';
import { AdStatus } from './types';

const router = Router();

router.post('/plans', (req: Request, res: Response) => {
  try {
    const data = createPlanSchema.parse(req.body);
    const plan = createPlan(data);
    res.status(201).json(plan);
  } catch (e: any) {
    if (e.status === 400) {
      res.status(400).json({ error: e.message });
    } else if (e instanceof z.ZodError) {
      res.status(400).json({ error: e.errors });
    } else {
      res.status(500).json({ error: 'Internal server error' });
    }
  }
});

router.get('/plans/:id', (req: Request, res: Response) => {
  const plan = getPlanWithStats(req.params.id);
  if (!plan) {
    res.status(404).json({ error: 'Plan not found' });
    return;
  }
  res.json(plan);
});

router.patch('/plans/:id/status', (req: Request, res: Response) => {
  try {
    const schema = z.object({ status: z.enum(['running', 'paused', 'ended']) });
    const { status } = schema.parse(req.body);
    const plan = updatePlanStatus(req.params.id, status as AdStatus);
    if (!plan) {
      res.status(404).json({ error: 'Plan not found' });
      return;
    }
    res.json(plan);
  } catch (e: any) {
    if (e.status === 400) {
      res.status(400).json({ error: e.message });
    } else if (e instanceof z.ZodError) {
      res.status(400).json({ error: e.errors });
    } else {
      res.status(500).json({ error: 'Internal server error' });
    }
  }
});

router.post('/auction', (req: Request, res: Response) => {
  try {
    const schema = z.object({
      age: z.string().optional(),
      gender: z.enum(['male', 'female', 'all']).optional(),
      region: z.string().optional(),
    });
    const opp = schema.parse(req.body);
    const plan = selectBestPlan(opp);
    if (!plan) {
      res.status(404).json({ error: 'No eligible plan' });
      return;
    }
    res.json(plan);
  } catch (e: any) {
    if (e instanceof z.ZodError) {
      res.status(400).json({ error: e.errors });
    } else {
      res.status(500).json({ error: 'Internal server error' });
    }
  }
});

router.post('/plans/:id/impression', (req: Request, res: Response) => {
  const plan = getPlan(req.params.id);
  if (!plan) {
    res.status(404).json({ error: 'Plan not found' });
    return;
  }
  const result = recordImpression(req.params.id);
  if (!result.success) {
    res.status(400).json({ error: 'Cannot record impression' });
    return;
  }
  if (result.notifyFailed) {
    res.status(200).json({ 
      success: true, 
      cost: result.cost, 
      budgetExhausted: result.budgetExhausted,
      message: '状态已更新通知发送失败'
    });
    return;
  }
  res.json({ success: true, cost: result.cost, budgetExhausted: result.budgetExhausted });
});

router.post('/plans/:id/click', (req: Request, res: Response) => {
  const plan = getPlan(req.params.id);
  if (!plan) {
    res.status(404).json({ error: 'Plan not found' });
    return;
  }
  recordClick(req.params.id);
  res.json({ success: true });
});

router.post('/plans/:id/conversion', (req: Request, res: Response) => {
  const plan = getPlan(req.params.id);
  if (!plan) {
    res.status(404).json({ error: 'Plan not found' });
    return;
  }
  recordConversion(req.params.id);
  res.json({ success: true });
});

export default router;
