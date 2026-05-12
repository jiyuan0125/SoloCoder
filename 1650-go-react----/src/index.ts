import express, { Request, Response } from 'express';
import {
  validateCreateRequest,
  createAuditLog,
  getLogById,
  listAuditLogs,
  compareLogs,
  getRetentionConfig,
  updateRetentionConfig,
  runCleanup,
} from './auditService';
import { CreateAuditLogRequest, QueryAuditLogsRequest, OperationType } from './types';

const app = express();
app.use(express.json());

const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

const METHOD_NOT_ALLOWED_MESSAGE = '日志不可修改';
const RESOURCE_MISMATCH_MESSAGE = '资源不匹配';

function getClientIp(req: Request): string {
  const xForwardedFor = req.headers['x-forwarded-for'];
  if (typeof xForwardedFor === 'string') {
    return xForwardedFor.split(',')[0].trim();
  }
  return req.ip || req.socket.remoteAddress || '';
}

app.post('/audit-logs', (req: Request, res: Response) => {
  const validation = validateCreateRequest(req.body);
  if (!validation.valid) {
    return res.status(400).json({ error: validation.errors.join('; ') });
  }

  const body = req.body as CreateAuditLogRequest;
  const sourceIp = getClientIp(req);

  const result = createAuditLog(body, sourceIp);

  if (result.wasDuplicate) {
    return res.status(200).json({ duplicated: true });
  }

  const log = getLogById(result.logId!);
  return res.status(201).json(log);
});

app.get('/audit-logs', (req: Request, res: Response) => {
  const query: QueryAuditLogsRequest = {};

  if (req.query.operatorId) {
    query.operatorId = String(req.query.operatorId);
  }

  if (req.query.operationType) {
    const op = String(req.query.operationType).toUpperCase() as OperationType;
    if (['CREATE', 'READ', 'UPDATE', 'DELETE'].includes(op)) {
      query.operationType = op;
    }
  }

  if (req.query.resourceType) {
    query.resourceType = String(req.query.resourceType);
  }

  if (req.query.startTime) {
    const t = parseInt(String(req.query.startTime), 10);
    if (!isNaN(t)) query.startTime = t;
  }

  if (req.query.endTime) {
    const t = parseInt(String(req.query.endTime), 10);
    if (!isNaN(t)) query.endTime = t;
  }

  if (req.query.page) {
    const p = parseInt(String(req.query.page), 10);
    if (!isNaN(p) && p > 0) query.page = p;
  }

  if (req.query.pageSize) {
    const ps = parseInt(String(req.query.pageSize), 10);
    if (isNaN(ps) || ps <= 0 || ps > 200) {
      return res.status(400).json({ error: 'pageSize 必须是 1-200 之间的整数' });
    }
    query.pageSize = ps;
  }

  const result = listAuditLogs(query);
  return res.status(200).json(result);
});

app.get('/audit-logs/compare', (req: Request, res: Response) => {
  const log1IdRaw = req.query.log1Id;
  const log2IdRaw = req.query.log2Id;

  if (!log1IdRaw || !log2IdRaw) {
    return res.status(400).json({ error: '缺少参数: log1Id 和 log2Id' });
  }

  const log1Id = parseInt(String(log1IdRaw), 10);
  const log2Id = parseInt(String(log2IdRaw), 10);

  if (isNaN(log1Id) || isNaN(log2Id)) {
    return res.status(400).json({ error: 'log1Id 和 log2Id 必须是整数' });
  }

  const result = compareLogs(log1Id, log2Id);

  if (!result.success) {
    if (result.notFound) {
      return res.status(404).json({ error: result.error });
    }
    if (result.resourceMismatch) {
      return res.status(400).json({ error: RESOURCE_MISMATCH_MESSAGE });
    }
    return res.status(400).json({ error: result.error });
  }

  return res.status(200).json(result.result);
});

app.put('/audit-logs/retention', (req: Request, res: Response) => {
  const body = req.body as { retentionDays?: number; operatorId?: string };

  if (typeof body.retentionDays !== 'number' || body.retentionDays < 0 || !Number.isInteger(body.retentionDays)) {
    return res.status(400).json({ error: 'retentionDays 必须是大于等于 0 的整数' });
  }

  const operatorId = body.operatorId || 'SYSTEM';
  const sourceIp = getClientIp(req);

  updateRetentionConfig(body.retentionDays, operatorId, sourceIp);

  const config = getRetentionConfig();
  return res.status(200).json(config);
});

app.get('/audit-logs/retention', (_req: Request, res: Response) => {
  const config = getRetentionConfig();
  return res.status(200).json(config);
});

app.all('/audit-logs/:id', (req: Request, res: Response) => {
  if (req.method === 'GET') {
    const id = parseInt(req.params.id, 10);
    if (isNaN(id)) {
      return res.status(400).json({ error: '无效的日志 ID' });
    }
    const log = getLogById(id);
    if (!log) {
      return res.status(404).json({ error: '日志不存在' });
    }
    return res.status(200).json(log);
  }

  return res.status(405).json({ error: METHOD_NOT_ALLOWED_MESSAGE });
});

app.all('/audit-logs', (req: Request, res: Response) => {
  if (req.method === 'POST' || req.method === 'GET') {
    return res.status(404).json({ error: '未找到' });
  }
  return res.status(405).json({ error: METHOD_NOT_ALLOWED_MESSAGE });
});

const CLEANUP_INTERVAL_MS = 60 * 60 * 1000;
setInterval(() => {
  runCleanup();
}, CLEANUP_INTERVAL_MS);

if (require.main === module) {
  app.listen(PORT, () => {
    console.log(`Audit log service listening on port ${PORT}`);
  });
}

export { app };
