import { Router, Request, Response } from 'express';
import { RuleService, RuleValidationError } from '../services/ruleService';

const router = Router();
const ruleService = RuleService.getInstance();

router.get('/', (req: Request, res: Response) => {
  const rules = ruleService.findAll();
  res.json(rules);
});

router.get('/:id', (req: Request, res: Response) => {
  const rule = ruleService.findById(req.params.id);
  if (!rule) {
    return res.status(404).json({ error: 'Rule not found' });
  }
  res.json(rule);
});

router.post('/', (req: Request, res: Response) => {
  try {
    const { name, condition, weight, priority, enabled } = req.body;

    const created = ruleService.create({
      name,
      condition,
      weight: Number(weight),
      priority: Number(priority) || 0,
      enabled: enabled !== undefined ? enabled : true
    });

    res.status(201).json(created);
  } catch (e) {
    if (e instanceof RuleValidationError) {
      return res.status(400).json({
        error: e.message,
        field: e.field,
        code: e.code
      });
    }
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id', (req: Request, res: Response) => {
  try {
    const { name, condition, weight, priority, enabled } = req.body;
    const updates: any = {};

    if (name !== undefined) updates.name = name;
    if (condition !== undefined) updates.condition = condition;
    if (weight !== undefined) updates.weight = Number(weight);
    if (priority !== undefined) updates.priority = Number(priority);
    if (enabled !== undefined) updates.enabled = enabled;

    const updated = ruleService.update(req.params.id, updates);
    if (!updated) {
      return res.status(404).json({ error: 'Rule not found' });
    }

    res.json(updated);
  } catch (e) {
    if (e instanceof RuleValidationError) {
      return res.status(400).json({
        error: e.message,
        field: e.field,
        code: e.code
      });
    }
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.delete('/:id', (req: Request, res: Response) => {
  const deleted = ruleService.delete(req.params.id);
  if (!deleted) {
    return res.status(404).json({ error: 'Rule not found' });
  }
  res.status(204).send();
});

export default router;
