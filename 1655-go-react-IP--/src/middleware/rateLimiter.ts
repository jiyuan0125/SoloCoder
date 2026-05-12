import { Request, Response, NextFunction } from 'express';
import { rateLimitService } from '../services/rateLimitService';

export function rateLimiter(req: Request, res: Response, next: NextFunction): void {
  if (!req.realIP) {
    res.status(400).json({ error: '无法解析客户端IP' });
    return;
  }

  const result = rateLimitService.checkRateLimit(req.realIP);

  res.setHeader('X-RateLimit-Limit', rateLimitService.getMaxRequests());
  res.setHeader('X-RateLimit-Remaining', result.remaining);
  res.setHeader('X-RateLimit-Reset', result.reset);

  if (!result.allowed) {
    res.status(429).json({ 
      error: '请求频率超限',
      retryAfter: Math.ceil((result.reset - Date.now()) / 1000)
    });
    return;
  }

  next();
}
