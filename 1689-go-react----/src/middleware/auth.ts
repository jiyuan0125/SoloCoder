import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';

const JWT_SECRET = process.env.JWT_SECRET || 'social-work-secret-key-2024';

export interface AuthRequest extends Request {
  user?: {
    id: number;
    username: string;
    role: string;
    name: string;
  };
}

export function authenticateToken(req: AuthRequest, res: Response, next: NextFunction): void {
  const authHeader = req.headers['authorization'];
  const token = authHeader && authHeader.split(' ')[1];

  if (!token) {
    res.status(401).json({ error: '未授权访问' });
    return;
  }

  jwt.verify(token, JWT_SECRET, (err: any, user: any) => {
    if (err) {
      res.status(403).json({ error: 'Token 无效' });
      return;
    }
    req.user = user;
    next();
  });
}

export function requireSupervisor(req: AuthRequest, res: Response, next: NextFunction): void {
  if (!req.user || req.user.role !== 'supervisor') {
    res.status(403).json({ error: '需要主管权限' });
    return;
  }
  next();
}

export function requireSocialWorker(req: AuthRequest, res: Response, next: NextFunction): void {
  if (!req.user || (req.user.role !== 'social_worker' && req.user.role !== 'supervisor')) {
    res.status(403).json({ error: '需要社工权限' });
    return;
  }
  next();
}

export function generateToken(user: { id: number; username: string; role: string; name: string }): string {
  return jwt.sign(user, JWT_SECRET, { expiresIn: '24h' });
}
