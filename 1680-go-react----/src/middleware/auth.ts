import { Request, Response, NextFunction } from 'express';
import { getUserById } from '../services/userService';
import { sendError } from '../utils';

declare global {
  namespace Express {
    interface Request {
      user?: any;
    }
  }
}

export const authenticate = (req: Request, res: Response, next: NextFunction) => {
  const userIdHeader = req.headers['x-user-id'];
  
  if (!userIdHeader) {
    return sendError(res, { code: 401, message: '未登录' });
  }

  const userId = parseInt(userIdHeader as string, 10);
  if (isNaN(userId)) {
    return sendError(res, { code: 401, message: '无效的用户ID' });
  }

  const user = getUserById(userId);
  if (!user) {
    return sendError(res, { code: 401, message: '用户不存在' });
  }

  req.user = user;
  next();
};

export const requireRegistered = (req: Request, res: Response, next: NextFunction) => {
  if (!req.user) {
    return sendError(res, { code: 401, message: '未登录' });
  }

  if (req.user.isRegistered !== 1) {
    return sendError(res, { code: 403, message: '未注册业主' });
  }

  next();
};

export const requireVerified = (req: Request, res: Response, next: NextFunction) => {
  if (!req.user) {
    return sendError(res, { code: 401, message: '未登录' });
  }

  if (req.user.isRegistered !== 1) {
    return sendError(res, { code: 403, message: '未注册业主' });
  }

  if (req.user.isVerified !== 1) {
    return sendError(res, { code: 403, message: '未实名认证' });
  }

  next();
};

export const requireCommittee = (req: Request, res: Response, next: NextFunction) => {
  if (!req.user) {
    return sendError(res, { code: 401, message: '未登录' });
  }

  if (req.user.role !== 'committee') {
    return sendError(res, { code: 403, message: '需要业委会权限' });
  }

  next();
};

export const requireExecutor = (req: Request, res: Response, next: NextFunction) => {
  if (!req.user) {
    return sendError(res, { code: 401, message: '未登录' });
  }

  if (req.user.role !== 'executor') {
    return sendError(res, { code: 403, message: '需要执行人权限' });
  }

  next();
};
