import { Request, Response, NextFunction } from 'express';

export function errorHandler(
  err: any,
  req: Request,
  res: Response,
  next: NextFunction
) {
  if (err.status) {
    const response: any = { message: err.message };
    if (err.allowedActions) {
      response.allowedActions = err.allowedActions;
    }
    res.status(err.status).json(response);
  } else {
    res.status(500).json({ message: err.message || '内部服务器错误' });
  }
}
