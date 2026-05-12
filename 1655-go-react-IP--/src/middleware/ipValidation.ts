import { Request, Response, NextFunction } from 'express';
import { ruleService } from '../services/ruleService';

const ADMIN_PATHS = /^\/gateway/;

export function ipValidation(_req: Request, _res: Response, next: NextFunction): void {
  next();
}

export function validateRequestIP(req: Request, res: Response, next: NextFunction): void {
  if (!req.realIP) {
    res.status(400).json({ error: '无法解析客户端IP' });
    return;
  }

  const checkResult = ruleService.checkIP(req.realIP);

  if (checkResult.action === 'allow') {
    res.status(200).json({ 
      message: '白名单命中，放行',
      ip: req.realIP,
      rule: checkResult.matchedRule
    });
    return;
  }

  if (checkResult.action === 'deny') {
    res.status(403).json({ 
      error: '黑名单命中，拒绝访问',
      ip: req.realIP,
      rule: checkResult.matchedRule
    });
    return;
  }

  next();
}

export function protectAdmin(req: Request, res: Response, next: NextFunction): void {
  if (!ADMIN_PATHS.test(req.path)) {
    next();
    return;
  }

  if (!req.realIP) {
    res.status(401).json({ error: '未授权' });
    return;
  }

  const checkResult = ruleService.checkIP(req.realIP);
  
  if (checkResult.action === 'allow') {
    next();
    return;
  }

  res.status(403).json({ 
    error: '无权限访问管理接口',
    ip: req.realIP
  });
}
