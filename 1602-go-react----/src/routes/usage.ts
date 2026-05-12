import { Router, Request, Response } from 'express';
import db from '../db';
import { checkQuota, recordUsage } from '../services/quotaService';

const router = Router();

router.get('/:tenantId', (req: Request, res: Response) => {
  const tenantId = Number(req.params.tenantId);
  
  const tenant = db.prepare('SELECT * FROM tenants WHERE id = ?').get(tenantId);
  if (!tenant) {
    return res.status(404).json({ message: '租户不存在' });
  }

  const usage = db.prepare('SELECT * FROM usage WHERE tenant_id = ?').get(tenantId);
  res.json(usage);
});

router.post('/:tenantId/consume', (req: Request, res: Response) => {
  const tenantId = Number(req.params.tenantId);
  const { resource_type, amount } = req.body;

  if (!resource_type || !['user', 'storage', 'api'].includes(resource_type)) {
    return res.status(400).json({ message: '无效的 resource_type，必须是 user, storage 或 api' });
  }

  if (amount === undefined || amount <= 0) {
    return res.status(400).json({ message: 'amount 必须是大于0的数字' });
  }

  const result = checkQuota(tenantId, resource_type, amount);
  
  if (!result.allowed) {
    return res.status(403).json({ 
      message: result.message, 
      code: 'QUOTA_EXCEEDED',
      resource: result.resource
    });
  }

  recordUsage(tenantId, resource_type, amount);

  const usage = db.prepare('SELECT * FROM usage WHERE tenant_id = ?').get(tenantId);
  
  res.json({ 
    message: '资源消耗成功', 
    resource: resource_type,
    amount,
    usage 
  });
});

export default router;
