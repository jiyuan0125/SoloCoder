import { Router, Request, Response } from 'express';
import { ruleService } from '../services/ruleService';

const router = Router({ mergeParams: true });

router.get('/', (req: Request, res: Response): void => {
  const { ruleId } = req.params;
  
  const rule = ruleService.getRule(ruleId);
  if (!rule) {
    res.status(404).json({ error: '规则不存在', ruleId });
    return;
  }

  const conditions = ruleService.getConditions(ruleId);
  res.json({ conditions, ruleId });
});

router.post('/', (req: Request, res: Response): void => {
  const { ruleId } = req.params;
  const { key, value, operator } = req.body;

  const rule = ruleService.getRule(ruleId);
  if (!rule) {
    res.status(404).json({ error: '规则不存在', ruleId });
    return;
  }

  if (!key || !value) {
    res.status(400).json({ error: '缺少必填字段: key, value' });
    return;
  }

  const validOperators = ['equals', 'contains', 'regex'];
  if (operator && !validOperators.includes(operator)) {
    res.status(400).json({ 
      error: 'operator 必须是: equals, contains, regex 之一' 
    });
    return;
  }

  const condition = ruleService.addCondition(
    ruleId,
    key,
    value,
    operator || 'equals'
  );

  if (!condition) {
    res.status(500).json({ error: '创建条件失败' });
    return;
  }

  res.status(201).json({ condition });
});

router.get('/:conditionId', (req: Request, res: Response): void => {
  const { ruleId, conditionId } = req.params;

  const rule = ruleService.getRule(ruleId);
  if (!rule) {
    res.status(404).json({ error: '规则不存在', ruleId });
    return;
  }

  const condition = ruleService.getCondition(conditionId);
  if (!condition) {
    res.status(404).json({ error: '条件不存在', conditionId });
    return;
  }

  if (condition.ruleId !== ruleId) {
    res.status(404).json({ 
      error: '条件不属于该规则',
      conditionId,
      ruleId
    });
    return;
  }

  res.json({ condition });
});

router.put('/:conditionId', (req: Request, res: Response): void => {
  const { ruleId, conditionId } = req.params;
  const { key, value, operator } = req.body;

  const rule = ruleService.getRule(ruleId);
  if (!rule) {
    res.status(404).json({ error: '规则不存在', ruleId });
    return;
  }

  const existingCondition = ruleService.getCondition(conditionId);
  if (!existingCondition) {
    res.status(404).json({ error: '条件不存在', conditionId });
    return;
  }

  if (existingCondition.ruleId !== ruleId) {
    res.status(404).json({ 
      error: '条件不属于该规则',
      conditionId,
      ruleId
    });
    return;
  }

  const validOperators = ['equals', 'contains', 'regex'];
  if (operator !== undefined && !validOperators.includes(operator)) {
    res.status(400).json({ 
      error: 'operator 必须是: equals, contains, regex 之一' 
    });
    return;
  }

  const updated = ruleService.updateCondition(conditionId, {
    key,
    value,
    operator
  });

  if (!updated) {
    res.status(500).json({ error: '更新条件失败' });
    return;
  }

  res.json({ condition: updated });
});

router.delete('/:conditionId', (req: Request, res: Response): void => {
  const { ruleId, conditionId } = req.params;

  const rule = ruleService.getRule(ruleId);
  if (!rule) {
    res.status(404).json({ error: '规则不存在', ruleId });
    return;
  }

  const existingCondition = ruleService.getCondition(conditionId);
  if (!existingCondition) {
    res.status(404).json({ error: '条件不存在', conditionId });
    return;
  }

  if (existingCondition.ruleId !== ruleId) {
    res.status(404).json({ 
      error: '条件不属于该规则',
      conditionId,
      ruleId
    });
    return;
  }

  const deleted = ruleService.deleteCondition(conditionId);
  if (deleted) {
    res.json({ 
      success: true, 
      message: '条件已删除',
      conditionId,
      ruleId
    });
  } else {
    res.status(500).json({ error: '删除条件失败' });
  }
});

export default router;
