import { Router, Request, Response } from 'express';
import { ruleService, RuleValidationError } from '../services/ruleService';
import { validateCIDR } from '../utils/ipUtils';

const router = Router();

router.get('/', (_req: Request, res: Response): void => {
  const rules = ruleService.getAllRules();
  res.json({ rules });
});

router.post('/', (req: Request, res: Response): void => {
  const { type, ipPattern, expiresAt, durationMinutes } = req.body;

  if (!type || !ipPattern) {
    res.status(400).json({ error: '缺少必填字段: type, ipPattern' });
    return;
  }

  if (type !== 'whitelist' && type !== 'blacklist') {
    res.status(400).json({ error: 'type 必须是 whitelist 或 blacklist' });
    return;
  }

  const cidrInfo = validateCIDR(ipPattern);
  if (!cidrInfo.isValid) {
    res.status(400).json({ 
      error: 'CIDR格式不合法',
      ipPattern,
      isCIDR: cidrInfo.isCIDR
    });
    return;
  }

  let calculatedExpiresAt = expiresAt;
  if (durationMinutes !== undefined) {
    calculatedExpiresAt = Date.now() + durationMinutes * 60 * 1000;
  }

  const result = ruleService.createRule({
    type,
    ipPattern,
    expiresAt: calculatedExpiresAt
  });

  if (!result.success) {
    if (result.error === RuleValidationError.RULE_CONFLICT) {
      res.status(409).json({ 
        error: '规则冲突：同一IP不能同时在白名单和黑名单中',
        conflictingRule: result.conflictRule
      });
      return;
    }
    res.status(400).json({ error: '创建规则失败' });
    return;
  }

  res.status(201).json({ rule: result.rule });
});

router.get('/:ruleId', (req: Request, res: Response): void => {
  const { ruleId } = req.params;
  const rule = ruleService.getRule(ruleId);

  if (!rule) {
    res.status(404).json({ error: '规则不存在', ruleId });
    return;
  }

  res.json({ rule });
});

router.put('/:ruleId', (req: Request, res: Response): void => {
  const { ruleId } = req.params;
  const { type, ipPattern, enabled, expiresAt, durationMinutes } = req.body;

  const existingRule = ruleService.getRule(ruleId);
  if (!existingRule) {
    res.status(404).json({ error: '规则不存在', ruleId });
    return;
  }

  if (type !== undefined && type !== 'whitelist' && type !== 'blacklist') {
    res.status(400).json({ error: 'type 必须是 whitelist 或 blacklist' });
    return;
  }

  if (ipPattern !== undefined) {
    const cidrInfo = validateCIDR(ipPattern);
    if (!cidrInfo.isValid) {
      res.status(400).json({ 
        error: 'CIDR格式不合法',
        ipPattern,
        isCIDR: cidrInfo.isCIDR
      });
      return;
    }
  }

  let calculatedExpiresAt = expiresAt;
  if (durationMinutes !== undefined) {
    calculatedExpiresAt = Date.now() + durationMinutes * 60 * 1000;
  }

  const result = ruleService.updateRule(ruleId, {
    type,
    ipPattern,
    enabled,
    expiresAt: calculatedExpiresAt
  });

  if (!result.success) {
    if (result.notFound) {
      res.status(404).json({ error: '规则不存在', ruleId });
      return;
    }
    if (result.error === RuleValidationError.RULE_CONFLICT) {
      res.status(409).json({ 
        error: '规则冲突：同一IP不能同时在白名单和黑名单中',
        conflictingRule: result.conflictRule
      });
      return;
    }
    res.status(400).json({ error: '更新规则失败' });
    return;
  }

  res.json({ rule: result.rule });
});

router.delete('/:ruleId', (req: Request, res: Response): void => {
  const { ruleId } = req.params;
  
  const existingRule = ruleService.getRule(ruleId);
  if (!existingRule) {
    res.status(404).json({ error: '规则不存在', ruleId });
    return;
  }

  const deleted = ruleService.deleteRule(ruleId);
  if (deleted) {
    res.json({ success: true, message: '规则已删除', ruleId });
  } else {
    res.status(500).json({ error: '删除规则失败' });
  }
});

export default router;
