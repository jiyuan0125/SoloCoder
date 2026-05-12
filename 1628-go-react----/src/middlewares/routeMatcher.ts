import type { Request, Response, NextFunction } from 'express';
import { getAllRoutePrefixes } from '../services/serviceStore';
import type { MatchedService } from '../types';

export function normalizePrefix(prefix: string): string {
  let normalized = prefix.trim();
  if (!normalized.startsWith('/')) normalized = '/' + normalized;
  if (normalized.endsWith('/')) normalized = normalized.slice(0, -1);
  return normalized;
}

export function matchRoute(requestPath: string): MatchedService | undefined {
  const services = getAllRoutePrefixes();
  const normalizedPath = normalizePrefix(requestPath);
  
  let bestMatch: MatchedService | undefined;
  let exactMatch: MatchedService | undefined;

  for (const service of services) {
    const prefix = normalizePrefix(service.route_prefix);
    
    if (normalizedPath === prefix) {
      exactMatch = {
        service,
        matchedPrefix: prefix,
        remainingPath: '/'
      };
      break;
    }

    if (normalizedPath.startsWith(prefix + '/')) {
      const remainingPath = normalizedPath.slice(prefix.length) || '/';
      const currentMatch: MatchedService = {
        service,
        matchedPrefix: prefix,
        remainingPath
      };

      if (!bestMatch || prefix.length > bestMatch.matchedPrefix.length) {
        bestMatch = currentMatch;
      }
    }
  }

  return exactMatch || bestMatch;
}

export function routeMatcherMiddleware(req: Request, res: Response, next: NextFunction): void {
  if (req.path.startsWith('/gateway/')) {
    return next();
  }

  const matched = matchRoute(req.path);
  
  if (!matched) {
    res.status(404).json({ error: 'Not Found', message: 'No upstream service matched the requested path' });
    return;
  }

  req.matchedService = matched;
  next();
}
