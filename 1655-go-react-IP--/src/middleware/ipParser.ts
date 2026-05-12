import { Request, Response, NextFunction } from 'express';
import { extractRealIP } from '../utils/ipUtils';

declare global {
  namespace Express {
    interface Request {
      realIP?: string;
      isTrustedProxy?: boolean;
      riskScore?: number;
      riskFactors?: Array<{
        type: string;
        description: string;
        score: number;
      }>;
    }
  }
}

export function parseClientIP(req: Request, _res: Response, next: NextFunction): void {
  const parsed = extractRealIP(req);
  req.realIP = parsed.ip;
  req.isTrustedProxy = parsed.isTrustedProxy;
  next();
}
