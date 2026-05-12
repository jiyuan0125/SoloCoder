import { Request, Response, NextFunction } from 'express';
import { verifyToken, JwtPayload } from '../utils/auth';

declare global {
  namespace Express {
    interface Request {
      user?: JwtPayload;
    }
  }
}

export function requireAuth(req: Request, res: Response, next: NextFunction): void {
  const authHeader = req.headers.authorization;
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    res.status(401).json({ error: '未授权' });
    return;
  }
  
  const token = authHeader.slice(7);
  const result = verifyToken(token);
  
  if (!result.valid) {
    if (result.error === 'malformed') {
      res.status(400).json({ error: 'Token格式错误' });
    } else {
      res.status(401).json({ error: 'Token无效或已过期' });
    }
    return;
  }
  
  req.user = result.payload;
  next();
}

export function optionalAuth(req: Request, res: Response, next: NextFunction): void {
  const authHeader = req.headers.authorization;
  
  if (authHeader && authHeader.startsWith('Bearer ')) {
    const token = authHeader.slice(7);
    const result = verifyToken(token);
    
    if (result.valid) {
      req.user = result.payload;
    }
  }
  
  next();
}
