import { Request, Response, NextFunction } from 'express';

interface AppError extends Error {
  statusCode?: number;
}

class ApiError extends Error {
  statusCode: number;
  constructor(statusCode: number, message: string) {
    super(message);
    this.statusCode = statusCode;
    this.name = 'ApiError';
  }
}

const errorHandler = (
  err: AppError,
  req: Request,
  res: Response,
  next: NextFunction
): void => {
  const statusCode = err.statusCode || 500;
  const message = err.message || 'Internal Server Error';

  res.status(statusCode).json({
    error: {
      status: statusCode,
      message: message,
      path: req.path,
      method: req.method,
    },
  });
};

const notFoundHandler = (req: Request, res: Response): void => {
  res.status(404).json({
    error: {
      status: 404,
      message: 'Not Found',
      path: req.path,
    },
  });
};

export { errorHandler, notFoundHandler, ApiError };
