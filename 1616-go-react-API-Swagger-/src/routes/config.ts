import { Router, Request, Response, NextFunction } from 'express';
import { getConfig, updateConfig } from '../config/config';
import { ApiError } from '../middleware/errorHandler';

const router = Router();

router.get('/', (req: Request, res: Response): void => {
  const config = getConfig();
  res.json(config);
});

router.put('/', (req: Request, res: Response, next: NextFunction): void => {
  try {
    const { proxy_timeout, max_request_body } = req.body;
    const updates: any = {};

    if (proxy_timeout !== undefined) {
      if (typeof proxy_timeout !== 'number' || proxy_timeout < 0) {
        throw new ApiError(400, 'proxy_timeout must be a non-negative number');
      }
      updates.proxy_timeout = proxy_timeout;
    }

    if (max_request_body !== undefined) {
      if (typeof max_request_body !== 'number' || max_request_body < 0) {
        throw new ApiError(400, 'max_request_body must be a non-negative number');
      }
      updates.max_request_body = max_request_body;
    }

    if (Object.keys(updates).length === 0) {
      throw new ApiError(400, 'No valid config fields provided');
    }

    const updatedConfig = updateConfig(updates);
    res.json(updatedConfig);
  } catch (err) {
    next(err);
  }
});

export { router as configRouter };
