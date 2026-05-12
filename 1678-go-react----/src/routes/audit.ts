import { Router, Request, Response } from 'express';
import { getAuditLogs, getAuditLogById } from '../services/auditService';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const logs = getAuditLogs();
  res.json(logs);
});

router.get('/:id', (req: Request, res: Response) => {
  const log = getAuditLogById(req.params.id);
  if (!log) {
    return res.status(404).json({ error: '审计日志不存在' });
  }
  res.json(log);
});

router.put('/:id', (req: Request, res: Response) => {
  res.status(405).json({ error: '审计日志不可修改' });
});

router.patch('/:id', (req: Request, res: Response) => {
  res.status(405).json({ error: '审计日志不可修改' });
});

router.delete('/:id', (req: Request, res: Response) => {
  res.status(405).json({ error: '审计日志不可删除' });
});

router.post('/', (req: Request, res: Response) => {
  res.status(405).json({ error: '审计日志只能通过系统操作自动创建，不可直接创建' });
});

export default router;
