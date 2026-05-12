import { db } from '../database';
import { BudgetAlert } from '../types';
import { generateId, getCurrentTime } from '../utils';
import { createAuditLog } from './auditService';
import { getProjectById } from './projectService';

const ALERT_THRESHOLD = 0.8;

export function checkBudgetAlert(projectId: string, operator: string): boolean {
  const project = getProjectById(projectId);
  if (!project) return false;

  const usageRate = project.usedBudget / project.budgetAmount;
  
  if (usageRate < ALERT_THRESHOLD) return false;

  const existingAlert = db.prepare(`
    SELECT * FROM budget_alerts WHERE project_id = ? ORDER BY created_at DESC LIMIT 1
  `).get(projectId) as BudgetAlert | undefined;

  if (existingAlert) {
    return true;
  }

  const id = generateId();
  const now = getCurrentTime();

  db.prepare(`
    INSERT INTO budget_alerts (id, project_id, usage_rate, created_at)
    VALUES (?, ?, ?, ?)
  `).run(id, projectId, usageRate, now);

  const alert: BudgetAlert = {
    id,
    projectId,
    usageRate,
    createdAt: now
  };

  createAuditLog(
    operator,
    '预算预警',
    `项目"${project.projectName}"预算使用已超过80%，当前使用率：${(usageRate * 100).toFixed(2)}%`,
    null,
    { project, alert }
  );

  return true;
}

export function getBudgetAlerts(projectId: string): BudgetAlert[] {
  return db.prepare(`
    SELECT * FROM budget_alerts WHERE project_id = ? ORDER BY created_at DESC
  `).all(projectId) as BudgetAlert[];
}

export function getAllBudgetAlerts(): BudgetAlert[] {
  return db.prepare(`
    SELECT * FROM budget_alerts ORDER BY created_at DESC
  `).all() as BudgetAlert[];
}
