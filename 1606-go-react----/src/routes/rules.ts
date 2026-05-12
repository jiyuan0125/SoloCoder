import { Router, Request, Response } from 'express';
import { ruleEngine } from '../services/ruleEngine';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { name, priority, conditionLogic, conditions, action, sensitiveWordIds } = req.body;
  
  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: '规则名称不能为空' });
  }
  
  if (priority === undefined || typeof priority !== 'number') {
    return res.status(400).json({ error: '优先级必须是数字' });
  }
  
  if (!['AND', 'OR'].includes(conditionLogic)) {
    return res.status(400).json({ error: '条件逻辑必须是 AND 或 OR' });
  }
  
  if (!Array.isArray(conditions)) {
    return res.status(400).json({ error: '条件必须是数组' });
  }
  
  if (!['REJECT', 'PENDING', 'PASS', 'LOG'].includes(action)) {
    return res.status(400).json({ error: '操作必须是 REJECT、PENDING、PASS 或 LOG' });
  }
  
  const rule = ruleEngine.create(
    name,
    priority,
    conditionLogic as 'AND' | 'OR',
    conditions,
    action as 'REJECT' | 'PENDING' | 'PASS' | 'LOG',
    sensitiveWordIds || []
  );
  
  return res.status(201).json(rule);
});

router.get('/', (_req: Request, res: Response) => {
  const rules = ruleEngine.findAll();
  return res.json(rules);
});

router.get('/:id', (req: Request, res: Response) => {
  const rule = ruleEngine.findById(req.params.id);
  if (!rule) {
    return res.status(404).json({ error: '规则不存在' });
  }
  return res.json(rule);
});

router.put('/:id', (req: Request, res: Response) => {
  const { name, priority, conditionLogic, conditions, action, sensitiveWordIds } = req.body;
  
  if (conditionLogic !== undefined && !['AND', 'OR'].includes(conditionLogic)) {
    return res.status(400).json({ error: '条件逻辑必须是 AND 或 OR' });
  }
  
  if (action !== undefined && !['REJECT', 'PENDING', 'PASS', 'LOG'].includes(action)) {
    return res.status(400).json({ error: '操作必须是 REJECT、PENDING、PASS 或 LOG' });
  }
  
  const updated = ruleEngine.update(
    req.params.id,
    name,
    priority,
    conditionLogic as 'AND' | 'OR' | undefined,
    conditions,
    action as 'REJECT' | 'PENDING' | 'PASS' | 'LOG' | undefined,
    sensitiveWordIds
  );
  
  if (!updated) {
    return res.status(404).json({ error: '规则不存在' });
  }
  
  return res.json(updated);
});

router.delete('/:id', (req: Request, res: Response) => {
  const deleted = ruleEngine.delete(req.params.id);
  if (!deleted) {
    return res.status(404).json({ error: '规则不存在' });
  }
  return res.status(204).send();
});

export default router;
