import db from '../db';
import { Quota, Alert } from '../types';
import { acquireLock, releaseLock } from '../utils/lock';

export function getCurrentQuota(tenantId: number): Quota | null {
  const tenant = db.prepare('SELECT * FROM tenants WHERE id = ?').get(tenantId) as any;
  if (!tenant) return null;

  const pkg = db.prepare('SELECT * FROM packages WHERE id = ?').get(tenant.package_id) as any;
  if (!pkg) return null;

  const baseQuota: Quota = {
    max_users: pkg.max_users,
    max_storage: pkg.max_storage,
    max_api_calls: pkg.max_api_calls,
  };

  const now = new Date().toISOString();
  const tempQuota = db.prepare('SELECT * FROM tenant_quotas WHERE tenant_id = ? AND is_temporary = 1 AND expires_at > ? ORDER BY created_at DESC LIMIT 1').get(tenantId, now) as any;

  if (tempQuota) {
    return {
      max_users: tempQuota.max_users,
      max_storage: tempQuota.max_storage,
      max_api_calls: tempQuota.max_api_calls,
    };
  }

  return baseQuota;
}

export function getBaseQuota(tenantId: number): Quota | null {
  const tenant = db.prepare('SELECT * FROM tenants WHERE id = ?').get(tenantId) as any;
  if (!tenant) return null;

  const pkg = db.prepare('SELECT * FROM packages WHERE id = ?').get(tenant.package_id) as any;
  if (!pkg) return null;

  return {
    max_users: pkg.max_users,
    max_storage: pkg.max_storage,
    max_api_calls: pkg.max_api_calls,
  };
}

export function getUsage(tenantId: number): any | null {
  return db.prepare('SELECT * FROM usage WHERE tenant_id = ?').get(tenantId) as any;
}

export function ensureUsage(tenantId: number): any {
  let usage = getUsage(tenantId);
  if (!usage) {
    const result = db.prepare('INSERT INTO usage (tenant_id) VALUES (?)').run(tenantId);
    usage = db.prepare('SELECT * FROM usage WHERE id = ?').get(result.lastInsertRowid as number) as any;
  }

  const now = new Date();
  const lastReset = new Date(usage.last_api_reset_at);

  if (now.getMonth() !== lastReset.getMonth() || now.getFullYear() !== lastReset.getFullYear()) {
    db.prepare('UPDATE usage SET api_calls_this_month = 0, last_api_reset_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?').run(usage.id);
    usage = getUsage(tenantId) as any;
  }

  return usage;
}

export function checkQuota(tenantId: number, resourceType: 'user' | 'storage' | 'api', amount: number = 1): { allowed: boolean; message?: string; resource?: string } {
  const quota = getCurrentQuota(tenantId);
  if (!quota) return { allowed: false, message: '租户不存在', resource: 'tenant' };

  const usage = ensureUsage(tenantId);
  let current = 0;
  let max = 0;
  let resourceName = '';

  switch (resourceType) {
    case 'user':
      current = usage.user_count;
      max = quota.max_users;
      resourceName = '用户数';
      break;
    case 'storage':
      current = usage.storage_used;
      max = quota.max_storage;
      resourceName = '存储空间';
      break;
    case 'api':
      current = usage.api_calls_this_month;
      max = quota.max_api_calls;
      resourceName = 'API调用量';
      break;
  }

  if (current + amount > max) {
    return { allowed: false, message: `${resourceName}已达上限，无法继续操作`, resource: resourceType };
  }

  return { allowed: true };
}

export function recordUsage(tenantId: number, resourceType: 'user' | 'storage' | 'api', amount: number): void {
  ensureUsage(tenantId);
  
  switch (resourceType) {
    case 'user':
      db.prepare('UPDATE usage SET user_count = user_count + ?, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = ?').run(amount, tenantId);
      break;
    case 'storage':
      db.prepare('UPDATE usage SET storage_used = storage_used + ?, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = ?').run(amount, tenantId);
      break;
    case 'api':
      db.prepare('UPDATE usage SET api_calls_this_month = api_calls_this_month + ?, updated_at = CURRENT_TIMESTAMP WHERE tenant_id = ?').run(amount, tenantId);
      break;
  }

  checkAndCreateAlert(tenantId);
}

function checkAndCreateAlert(tenantId: number): void {
  const quota = getCurrentQuota(tenantId);
  if (!quota) return;

  const usage = getUsage(tenantId);
  if (!usage) return;

  const resources: Array<{ name: string; usage: number; max: number }> = [
    { name: '用户数', usage: usage.user_count, max: quota.max_users },
    { name: '存储空间', usage: usage.storage_used, max: quota.max_storage },
    { name: 'API调用量', usage: usage.api_calls_this_month, max: quota.max_api_calls },
  ];

  for (const r of resources) {
    const percentage = Math.round((r.usage / r.max) * 100);
    
    if (percentage >= 80) {
      const existing = db.prepare('SELECT * FROM alerts WHERE tenant_id = ? AND resource_type = ? AND handled = 0 ORDER BY created_at DESC LIMIT 1').get(tenantId, r.name) as any;

      if (!existing) {
        db.prepare('INSERT INTO alerts (tenant_id, resource_type, usage_percentage, message) VALUES (?, ?, ?, ?)').run(tenantId, r.name, percentage, `${r.name}使用量已达${percentage}%`);
      }
    }
  }
}

export function applyTemporaryQuota(
  tenantId: number,
  newQuota: Partial<Quota>,
  reason: string,
  validDays: number
): { success: boolean; message: string; code?: number } {
  const locked = acquireLock(tenantId);
  if (!locked) {
    return { success: false, message: '配额变更操作正在进行中，请稍后重试', code: 409 };
  }

  try {
    const baseQuota = getBaseQuota(tenantId);
    if (!baseQuota) {
      return { success: false, message: '租户不存在', code: 404 };
    }

    const currentQuota = getCurrentQuota(tenantId)!;

    const effectiveNewQuota: Quota = {
      max_users: newQuota.max_users ?? currentQuota.max_users,
      max_storage: newQuota.max_storage ?? currentQuota.max_storage,
      max_api_calls: newQuota.max_api_calls ?? currentQuota.max_api_calls,
    };

    if (
      effectiveNewQuota.max_users <= baseQuota.max_users &&
      effectiveNewQuota.max_storage <= baseQuota.max_storage &&
      effectiveNewQuota.max_api_calls <= baseQuota.max_api_calls
    ) {
      return { success: false, message: '临时配额值必须大于等于原配额', code: 400 };
    }

    const now = new Date();
    const expiresAt = new Date(now.getTime() + validDays * 24 * 60 * 60 * 1000);

    db.prepare(
      `INSERT INTO tenant_quotas (tenant_id, max_users, max_storage, max_api_calls, is_temporary, expires_at, reason) VALUES (?, ?, ?, ?, 1, ?, ?)`
    ).run(
      tenantId,
      effectiveNewQuota.max_users,
      effectiveNewQuota.max_storage,
      effectiveNewQuota.max_api_calls,
      expiresAt.toISOString(),
      reason
    );

    return { success: true, message: '临时配额设置成功' };
  } finally {
    releaseLock(tenantId);
  }
}

export function checkTenantDeletion(tenantId: number): { canDelete: boolean; blockers: string[] } {
  const blockers: string[] = [];

  const activeSubscriptions = db.prepare('SELECT COUNT(*) as count FROM subscriptions WHERE tenant_id = ? AND status = ?').get(tenantId, 'active') as any;
  if (activeSubscriptions.count > 0) {
    blockers.push(`${activeSubscriptions.count}个活跃订阅`);
  }

  const unhandledAlerts = db.prepare('SELECT COUNT(*) as count FROM alerts WHERE tenant_id = ? AND handled = 0').get(tenantId) as any;
  if (unhandledAlerts.count > 0) {
    blockers.push(`${unhandledAlerts.count}个未处理告警`);
  }

  return { canDelete: blockers.length === 0, blockers };
}
