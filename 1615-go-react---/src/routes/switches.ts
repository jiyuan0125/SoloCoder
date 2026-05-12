import { Router, Request, Response, NextFunction } from 'express';
import { SwitchService } from '../service/SwitchService';
import { AppError, BadRequestError } from '../utils/errors';

const router = Router();
const service = new SwitchService();

function handleError(err: unknown, res: Response): void {
  if (err instanceof AppError) {
    res.status(err.statusCode).json({
      error: err.message,
      code: err.code
    });
    return;
  }
  console.error('Unexpected error:', err);
  res.status(500).json({ error: 'Internal server error' });
}

function parseEvaluationContext(req: Request): { userId?: string; attributes?: Record<string, any> } {
  const context: { userId?: string; attributes?: Record<string, any> } = {};
  if (typeof req.query.userId === 'string' && req.query.userId) {
    context.userId = req.query.userId;
  }
  if (typeof req.query.attributes === 'string') {
    try {
      context.attributes = JSON.parse(req.query.attributes);
    } catch {
      throw new BadRequestError('Invalid attributes query parameter: must be valid JSON');
    }
  }
  return context;
}

function parseServiceName(req: Request): string | undefined {
  const svc = req.query.serviceName ?? req.header('X-Service-Name');
  return typeof svc === 'string' && svc ? svc : undefined;
}

router.post('/', (req: Request, res: Response) => {
  try {
    const result = service.create(req.body);
    res.status(201).json(result);
  } catch (err) {
    handleError(err, res);
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const all = service.getAll();
    res.json(all);
  } catch (err) {
    handleError(err, res);
  }
});

router.get('/:key', (req: Request, res: Response) => {
  try {
    const context = parseEvaluationContext(req);
    const serviceName = parseServiceName(req);
    const result = service.getByKey(req.params.key, context, serviceName);
    res.json(result);
  } catch (err) {
    handleError(err, res);
  }
});

router.put('/:key', (req: Request, res: Response) => {
  try {
    const result = service.update(req.params.key, req.body);
    res.json(result);
  } catch (err) {
    handleError(err, res);
  }
});

router.delete('/:key', (req: Request, res: Response) => {
  try {
    const operator = req.query.operator ?? req.header('X-Operator');
    if (typeof operator !== 'string' || !operator) {
      throw new BadRequestError('operator is required (query param or X-Operator header)');
    }
    service.delete(req.params.key, operator);
    res.status(204).send();
  } catch (err) {
    handleError(err, res);
  }
});

router.post('/batch', (req: Request, res: Response) => {
  try {
    const keys = req.body.keys;
    if (!Array.isArray(keys)) {
      throw new BadRequestError('Request body must contain "keys" array');
    }
    const context: { userId?: string; attributes?: Record<string, any> } = {};
    if (req.body.userId) context.userId = req.body.userId;
    if (req.body.attributes) context.attributes = req.body.attributes;
    const serviceName = req.body.serviceName ?? req.header('X-Service-Name');
    const svcName = typeof serviceName === 'string' ? serviceName : undefined;
    const result = service.batchGet(keys, context, svcName);
    if (result.stale) {
      res.setHeader('X-Config-Stale', 'true');
    }
    res.json(result);
  } catch (err) {
    handleError(err, res);
  }
});

router.get('/:key/history', (req: Request, res: Response) => {
  try {
    const limit = typeof req.query.limit === 'string' ? parseInt(req.query.limit, 10) : 50;
    const history = service.getHistory(req.params.key, isNaN(limit) ? 50 : limit);
    res.json(history);
  } catch (err) {
    handleError(err, res);
  }
});

router.get('/sync/version', (req: Request, res: Response) => {
  try {
    const version = service.getCurrentVersion();
    res.json({ version });
  } catch (err) {
    handleError(err, res);
  }
});

router.get('/sync/incremental', (req: Request, res: Response) => {
  try {
    const since = typeof req.query.sinceVersion === 'string' ? parseInt(req.query.sinceVersion, 10) : 0;
    if (isNaN(since)) {
      throw new BadRequestError('sinceVersion must be a number');
    }
    const update = service.getIncremental(since);
    res.json(update);
  } catch (err) {
    handleError(err, res);
  }
});

export default router;
