import { Router, Request, Response } from 'express';
import { Span, TraceQueryParams } from '../types';
import { store } from '../storage/InMemoryStore';

const router = Router();

const parseNumberQuery = (value: string | undefined): number | undefined => {
  if (value === undefined) return undefined;
  const num = Number(value);
  return isNaN(num) ? undefined : num;
};

const parsePageParams = (req: Request) => {
  const page = parseNumberQuery(req.query.page as string) ?? 1;
  const pageSize = parseNumberQuery(req.query.pageSize as string) ?? 50;
  return {
    page: Math.max(1, page),
    pageSize: Math.max(1, Math.min(pageSize, 1000))
  };
};

router.post('/spans', (req: Request, res: Response) => {
  const span: Span = req.body;

  if (!span.traceId || !span.spanId) {
    return res.status(400).json({
      error: 'traceId and spanId are required'
    });
  }

  store.addSpan(span);
  return res.status(201).json({ received: true });
});

router.get('/traces/slow', (req: Request, res: Response) => {
  const threshold = parseNumberQuery(req.query.threshold as string) ?? 1000;
  const slowTraces = store.getSlowTraces(threshold);
  return res.json(slowTraces);
});

router.get('/traces', (req: Request, res: Response) => {
  const { page, pageSize } = parsePageParams(req);

  const params: TraceQueryParams = {
    serviceName: req.query.serviceName as string | undefined,
    operationName: req.query.operationName as string | undefined,
    startTime: parseNumberQuery(req.query.startTime as string),

    endTime: parseNumberQuery(req.query.endTime as string),
    page,
    pageSize
  };

  const result = store.queryTraces(params);
  const trees = result.data.map(traceId => store.getTraceTree(traceId)).filter(Boolean);

  return res.json({
    data: trees,
    page: result.page,
    pageSize: result.pageSize,
    total: result.total,
    totalPages: result.totalPages
  });
});

router.get('/stats/services', (_req: Request, res: Response) => {
  const stats = store.getServiceStats();
  return res.json(stats);
});

router.get('/traces/:traceId', (req: Request, res: Response) => {
  const { traceId } = req.params;

  if (!store.hasTrace(traceId)) {
    return res.status(404).json({
      error: 'trace not found'
    });
  }

  const tree = store.getTraceTree(traceId);
  return res.json(tree);
});

router.get('/traces/:traceId/spans', (req: Request, res: Response) => {
  const { traceId } = req.params;
  const { page, pageSize } = parsePageParams(req);

  const result = store.getSpansPaginated(traceId, page, pageSize);
  return res.json(result);
});

export default router;
