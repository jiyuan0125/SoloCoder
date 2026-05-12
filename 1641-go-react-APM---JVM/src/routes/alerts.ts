import { Router, Request, Response } from 'express';
import { createAlertRule, getAllAlertRules } from '../storage';
import { AlertRule } from '../types';

const router = Router();

const VALID_OPERATORS = ['>', '<', '>=', '<=', '==', '!='];
const VALID_NOTIFICATION_TYPES = ['email', 'webhook', 'sms'];

router.post('/rules', (req: Request, res: Response) => {
  const {
    name,
    description,
    metric,
    operator,
    threshold,
    notificationType,
    notificationTarget,
  } = req.body;
  
  if (!name || !metric || !operator || threshold === undefined || !notificationType || !notificationTarget) {
    return res.status(400).json({
      error: 'Missing required fields: name, metric, operator, threshold, notificationType, notificationTarget'
    });
  }
  
  if (!VALID_OPERATORS.includes(operator)) {
    return res.status(400).json({
      error: `Invalid operator. Must be one of: ${VALID_OPERATORS.join(', ')}`
    });
  }
  
  if (!VALID_NOTIFICATION_TYPES.includes(notificationType)) {
    return res.status(400).json({
      error: `Invalid notificationType. Must be one of: ${VALID_NOTIFICATION_TYPES.join(', ')}`
    });
  }
  
  if (typeof threshold !== 'number') {
    return res.status(400).json({ error: 'threshold must be a number'
    });
  }
  
  const rule = createAlertRule(
    name,
    description || '',
    metric,
    operator as AlertRule['operator'],
    threshold,
    notificationType as AlertRule['notificationType'],
    notificationTarget
  );
  
  res.status(201).json(rule);
});

router.get('/rules', (req: Request, res: Response) => {
  const rules = getAllAlertRules();
  res.status(200).json(rules);
});

export default router;
