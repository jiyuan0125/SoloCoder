import { Request, Response, NextFunction } from 'express';

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export function handleError(error: Error, _req: Request, res: Response, _next: NextFunction): void {
  const message = error.message;
  let statusCode = 500;

  if (message.startsWith('BAD_REQUEST:')) {
    statusCode = 400;
  } else if (message.startsWith('NOT_FOUND:')) {
    statusCode = 404;
  } else if (message.startsWith('CONFLICT:')) {
    statusCode = 409;
  } else if (message.startsWith('UNAUTHORIZED:')) {
    statusCode = 401;
  } else if (message.startsWith('FORBIDDEN:')) {
    statusCode = 403;
  }

  const errorMessage = message.includes(':') 
    ? message.split(':').slice(1).join(':').trim()
    : message;

  res.status(statusCode).json({
    success: false,
    error: errorMessage
  });
}

export function asyncHandler(fn: (req: Request, res: Response, next: NextFunction) => Promise<void>) {
  return (req: Request, res: Response, next: NextFunction): void => {
    Promise.resolve(fn(req, res, next)).catch(next);
  };
}
