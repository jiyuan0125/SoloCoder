import express, { Request, Response } from 'express';
import * as opportunityService from '../services/opportunityService';

const router = express.Router();

router.get('/', (_req: Request, res: Response) => {
  const opportunities = opportunityService.getOpportunities();
  res.json(opportunities);
});

router.get('/:id', (req: Request, res: Response) => {
  const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
  const opportunity = opportunityService.getOpportunityById(id);
  if (opportunity) {
    res.json(opportunity);
  } else {
    res.status(404).json({ error: '商机不存在' });
  }
});

router.get('/:id/history', (req: Request, res: Response) => {
  const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
  const opportunity = opportunityService.getOpportunityById(id);
  if (opportunity) {
    const history = opportunityService.getOpportunityHistory(id);
    res.json(history);
  } else {
    res.status(404).json({ error: '商机不存在' });
  }
});

router.post('/', (req: Request, res: Response) => {
  const { name, customerId, customerName, amount, salespersonId } = req.body;
  if (!name || !customerId || !customerName || amount === undefined || !salespersonId) {
    return res.status(400).json({ error: 'name, customerId, customerName, amount, salespersonId 是必填项' });
  }
  const result = opportunityService.createOpportunity({
    name,
    customerId,
    customerName,
    amount,
    salespersonId,
  });

  if (result.success && result.opportunity) {
    res.status(201).json(result.opportunity);
  } else if (result.error) {
    res.status(result.error.status).json({ error: result.error.message });
  } else {
    res.status(500).json({ error: '创建失败' });
  }
});

router.patch('/:id/stage', (req: Request, res: Response) => {
  const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
  const { newStage, reason } = req.body;
  if (!newStage) {
    return res.status(400).json({ error: 'newStage 是必填项' });
  }
  const result = opportunityService.updateStage(id, { newStage, reason });

  if (result.success && result.opportunity) {
    res.json(result.opportunity);
  } else if (result.error) {
    res.status(result.error.status).json({ error: result.error.message });
  } else {
    res.status(500).json({ error: '更新失败' });
  }
});

router.patch('/:id', (req: Request, res: Response) => {
  const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
  const result = opportunityService.updateOpportunityDetails(id, req.body);

  if (result.success && result.opportunity) {
    res.json(result.opportunity);
  } else if (result.error) {
    res.status(result.error.status).json({ error: result.error.message });
  } else {
    res.status(500).json({ error: '更新失败' });
  }
});

router.delete('/:id', (req: Request, res: Response) => {
  const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
  const deleted = opportunityService.deleteOpportunity(id);
  if (deleted) {
    res.status(204).send();
  } else {
    res.status(404).json({ error: '商机不存在' });
  }
});

export default router;
