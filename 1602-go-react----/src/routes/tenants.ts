import { Router, Request, Response } from 'express';
import db from '../db';
import { checkTenantDeletion, getCurrentQuota, getBaseQuota, applyTemporaryQuota } from '../services/quotaService';

const router = Router();

router.get('/', (_req: Request, res: Response) => {
  const tenants = db.prepare('SELECT * FROM tenants ORDER BY created_at DESC').all();
  res.json(tenants);
});

router.get('/:id', (req: Request, res: Response) => {
  const tenant = db.prepare('SELECT * FROM tenants WHERE id = ?').get(Number(req.params.id));
  if (!tenant) {
    return res.status(404).json({ message: '租户不存在' });
  }
  res.json(tenant);
});

router.get('/:id/quota', (req: Request, res: Response) => {
  const tenantId = Number(req.params.id);
  const tenant = db.prepare('SELECT * FROM tenants WHERE id = ?').get(tenantId);
  if (!tenant) {
    return res.status(404).json({ message: '租户不存在' });
  }

  const currentQuota = getCurrentQuota(tenantId);
  const baseQuota = getBaseQuota(tenantId);

  res.json({
    base: baseQuota,
    current: currentQuota,
  });
});

router.get('/:id/quotas-history', (req: Request, res: Response) => {
  const tenantId = Number(req.params.id);
  const quotas = db.prepare('SELECT * FROM tenant_quotas WHERE tenant_id = ? ORDER BY created_at DESC').all(tenantId);
  res.json(quotas);
});

router.get('/:id/alerts', (req: Request, res: Response) => {
  const tenantId = Number(req.params.id);
  const alerts = db.prepare('SELECT * FROM alerts WHERE tenant_id = ? ORDER BY created_at DESC').all(tenantId);
  res.json(alerts);
});

router.post('/', (req: Request, res: Response) => {
  const { enterprise_name, contact_person, contact_phone, package_id } = req.body;

  if (!enterprise_name || !contact_person || !contact_phone || !package_id) {
    return res.status(400).json({ message: '缺少必填字段' });
  }

  const existingName = db.prepare('SELECT * FROM tenants WHERE enterprise_name = ?').get(enterprise_name);
  if (existingName) {
    return res.status(409).json({ message: '企业名称已存在' });
  }

  const pkg = db.prepare('SELECT * FROM packages WHERE id = ?').get(package_id);
  if (!pkg) {
    return res.status(400).json({ message: '套餐不存在' });
  }

  const result = db.prepare(
    'INSERT INTO tenants (enterprise_name, contact_person, contact_phone, package_id) VALUES (?, ?, ?, ?)'
  ).run(enterprise_name, contact_person, contact_phone, package_id);

  db.prepare('INSERT INTO usage (tenant_id) VALUES (?)').run(result.lastInsertRowid as number);

  const tenant = db.prepare('SELECT * FROM tenants WHERE id = ?').get(result.lastInsertRowid as number);
  res.status(201).json(tenant);
});

router.post('/:id/quota', (req: Request, res: Response) => {
  const tenantId = Number(req.params.id);
  const { max_users, max_storage, max_api_calls, reason, valid_days } = req.body;

  if (!reason || valid_days === undefined) {
    return res.status(400).json({ message: '缺少必填字段：reason 和 valid_days' });
  }

  const newQuota = {
    max_users,
    max_storage,
    max_api_calls,
  };

  const result = applyTemporaryQuota(tenantId, newQuota, reason, valid_days);

  if (!result.success) {
    return res.status(result.code || 400).json({ message: result.message });
  }

  res.json({ message: result.message });
});

router.put('/:id', (req: Request, res: Response) => {
  const id = Number(req.params.id);
  const { enterprise_name, contact_person, contact_phone, package_id } = req.body;

  const existing = db.prepare('SELECT * FROM tenants WHERE id = ?').get(id);
  if (!existing) {
    return res.status(404).json({ message: '租户不存在' });
  }

  if (enterprise_name) {
    const nameConflict = db.prepare('SELECT * FROM tenants WHERE enterprise_name = ? AND id != ?').get(enterprise_name, id);
    if (nameConflict) {
      return res.status(409).json({ message: '企业名称已存在' });
    }
  }

  if (package_id) {
    const pkg = db.prepare('SELECT * FROM packages WHERE id = ?').get(package_id);
    if (!pkg) {
      return res.status(400).json({ message: '套餐不存在' });
    }
  }

  db.prepare(
    'UPDATE tenants SET enterprise_name = COALESCE(?, enterprise_name), contact_person = COALESCE(?, contact_person), contact_phone = COALESCE(?, contact_phone), package_id = COALESCE(?, package_id), updated_at = CURRENT_TIMESTAMP WHERE id = ?'
  ).run(enterprise_name, contact_person, contact_phone, package_id, id);

  const updated = db.prepare('SELECT * FROM tenants WHERE id = ?').get(id);
  res.json(updated);
});

router.delete('/:id', (req: Request, res: Response) => {
  const id = Number(req.params.id);

  const existing = db.prepare('SELECT * FROM tenants WHERE id = ?').get(id);
  if (!existing) {
    return res.status(404).json({ message: '租户不存在' });
  }

  const { canDelete, blockers } = checkTenantDeletion(id);

  if (!canDelete) {
    return res.status(400).json({ 
      message: '无法删除租户，存在以下阻断资源',
      blockers 
    });
  }

  db.prepare('DELETE FROM usage WHERE tenant_id = ?').run(id);
  db.prepare('DELETE FROM alerts WHERE tenant_id = ?').run(id);
  db.prepare('DELETE FROM tenant_quotas WHERE tenant_id = ?').run(id);
  db.prepare('DELETE FROM subscriptions WHERE tenant_id = ?').run(id);
  db.prepare('DELETE FROM tenants WHERE id = ?').run(id);

  res.json({ message: '租户已删除' });
});

export default router;
