import { v4 as uuidv4 } from 'uuid';
import { db } from '../database';
import {
  Project,
  ProjectStatus,
  ProjectCategory,
  RewardTier,
  SupportRecord,
  CreateProjectRequest,
} from '../types';

const VALID_CATEGORIES: ProjectCategory[] = ['film', 'music', 'publishing', 'game', 'design'];

const STATUS_TRANSITIONS: Record<ProjectStatus, ProjectStatus[]> = {
  preparing: ['funding', 'cancelled'],
  funding: ['succeeded', 'failed', 'cancelled'],
  succeeded: [],
  failed: [],
  cancelled: [],
};

function validateCategory(category: string): category is ProjectCategory {
  return VALID_CATEGORIES.includes(category as ProjectCategory);
}

function validateStatus(status: string): status is ProjectStatus {
  return ['preparing', 'funding', 'succeeded', 'failed', 'cancelled'].includes(status);
}

export class ProjectError extends Error {
  statusCode: number;
  constructor(message: string, statusCode: number) {
    super(message);
    this.statusCode = statusCode;
  }
}

function mapProjectRow(row: any): Project {
  return {
    id: row.id,
    name: row.name,
    category: row.category as ProjectCategory,
    targetAmount: row.target_amount,
    raisedAmount: row.raised_amount,
    deadline: row.deadline,
    description: row.description,
    status: row.status as ProjectStatus,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
    succeededAt: row.succeeded_at || undefined,
  };
}

function mapRewardTierRow(row: any): RewardTier {
  return {
    id: row.id,
    projectId: row.project_id,
    amount: row.amount,
    description: row.description,
    createdAt: row.created_at,
  };
}

function mapSupportRow(row: any): SupportRecord {
  return {
    id: row.id,
    projectId: row.project_id,
    rewardTierId: row.reward_tier_id,
    supporterId: row.supporter_id,
    amount: row.amount,
    status: row.status,
    createdAt: row.created_at,
    refundedAt: row.refunded_at || undefined,
  };
}

function getNowISO(): string {
  return new Date().toISOString();
}

export function createProject(req: CreateProjectRequest): { project: Project; rewardTiers: RewardTier[] } {
  if (!req.name || req.name.trim().length === 0) {
    throw new ProjectError('项目名称不能为空', 400);
  }

  if (!validateCategory(req.category)) {
    throw new ProjectError('无效的项目类别', 400);
  }

  if (!req.targetAmount || req.targetAmount <= 0) {
    throw new ProjectError('目标金额必须大于0', 400);
  }

  const deadlineDate = new Date(req.deadline);
  if (isNaN(deadlineDate.getTime())) {
    throw new ProjectError('无效的截止日期', 400);
  }

  if (deadlineDate.getTime() <= Date.now()) {
    throw new ProjectError('截止日期必须晚于当前时间', 400);
  }

  if (!req.rewardTiers || req.rewardTiers.length === 0) {
    throw new ProjectError('至少需要设置一个回报档位', 400);
  }

  for (const tier of req.rewardTiers) {
    if (!tier.amount || tier.amount <= 0) {
      throw new ProjectError('回报档位金额必须大于0', 400);
    }
  }

  const existing = db.prepare('SELECT id FROM projects WHERE name = ?').get(req.name) as any;
  if (existing) {
    throw new ProjectError('项目名称已存在', 409);
  }

  const now = getNowISO();
  const projectId = uuidv4();

  db.prepare(`
    INSERT INTO projects (id, name, category, target_amount, raised_amount, deadline, description, status, created_at, updated_at)
    VALUES (?, ?, ?, ?, 0, ?, ?, 'preparing', ?, ?)
  `).run(
    projectId,
    req.name,
    req.category,
    req.targetAmount,
    req.deadline,
    req.description,
    now,
    now
  );

  const rewardTiers: RewardTier[] = [];
  const insertTier = db.prepare(`
    INSERT INTO reward_tiers (id, project_id, amount, description, created_at)
    VALUES (?, ?, ?, ?, ?)
  `);

  for (const tier of req.rewardTiers) {
    const tierId = uuidv4();
    insertTier.run(tierId, projectId, tier.amount, tier.description, now);
    rewardTiers.push({
      id: tierId,
      projectId,
      amount: tier.amount,
      description: tier.description,
      createdAt: now,
    });
  }

  const projectRow = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  return { project: mapProjectRow(projectRow), rewardTiers };
}

export function getProjectById(id: string): Project | null {
  const row = db.prepare('SELECT * FROM projects WHERE id = ?').get(id) as any;
  return row ? mapProjectRow(row) : null;
}

export function getProjectByName(name: string): Project | null {
  const row = db.prepare('SELECT * FROM projects WHERE name = ?').get(name) as any;
  return row ? mapProjectRow(row) : null;
}

export function listProjects(): Project[] {
  const rows = db.prepare('SELECT * FROM projects ORDER BY created_at DESC').all();
  return rows.map(mapProjectRow);
}

export function getRewardTiersByProjectId(projectId: string): RewardTier[] {
  const rows = db.prepare('SELECT * FROM reward_tiers WHERE project_id = ? ORDER BY amount ASC').all(projectId);
  return rows.map(mapRewardTierRow);
}

export function getRewardTierById(id: string): RewardTier | null {
  const row = db.prepare('SELECT * FROM reward_tiers WHERE id = ?').get(id) as any;
  return row ? mapRewardTierRow(row) : null;
}

export function canTransitionStatus(currentStatus: ProjectStatus, newStatus: ProjectStatus): boolean {
  const allowed = STATUS_TRANSITIONS[currentStatus];
  return allowed ? allowed.includes(newStatus) : false;
}

export function updateProjectStatus(projectId: string, newStatus: ProjectStatus): Project {
  const projectRow = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!projectRow) {
    throw new ProjectError('项目不存在', 404);
  }

  const currentStatus = projectRow.status as ProjectStatus;
  if (!canTransitionStatus(currentStatus, newStatus)) {
    throw new ProjectError(`非法状态跳转: ${currentStatus} -> ${newStatus}`, 400);
  }

  const now = getNowISO();
  const updateData: any = {
    status: newStatus,
    updatedAt: now,
  };

  if (newStatus === 'succeeded') {
    db.prepare(`
      UPDATE projects
      SET status = ?, updated_at = ?, succeeded_at = ?
      WHERE id = ?
    `).run(newStatus, now, now, projectId);
  } else {
    db.prepare(`
      UPDATE projects
      SET status = ?, updated_at = ?
      WHERE id = ?
    `).run(newStatus, now, projectId);
  }

  const updatedRow = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  return mapProjectRow(updatedRow);
}

export function checkAndUpdateExpiredProject(projectId: string): Project {
  const projectRow = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!projectRow) {
    throw new ProjectError('项目不存在', 404);
  }

  const project = mapProjectRow(projectRow);
  if (project.status !== 'funding') {
    return project;
  }

  const deadline = new Date(project.deadline).getTime();
  const now = Date.now();

  if (now >= deadline) {
    const newStatus = project.raisedAmount >= project.targetAmount ? 'succeeded' : 'failed';
    return updateProjectStatus(projectId, newStatus);
  }

  return project;
}

export function supportProject(rewardTierId: string, supporterId: string): SupportRecord {
  const rewardTier = getRewardTierById(rewardTierId);
  if (!rewardTier) {
    throw new ProjectError('回报档位不存在', 404);
  }

  const projectRow = db.prepare('SELECT * FROM projects WHERE id = ?').get(rewardTier.projectId) as any;
  if (!projectRow) {
    throw new ProjectError('项目不存在', 404);
  }

  let project = mapProjectRow(projectRow);
  project = checkAndUpdateExpiredProject(project.id);

  if (project.status !== 'funding') {
    throw new ProjectError('项目不处于筹款中状态', 400);
  }

  const existingSupport = db.prepare(
    'SELECT id FROM support_records WHERE supporter_id = ? AND reward_tier_id = ? AND status = ?'
  ).get(supporterId, rewardTierId, 'active') as any;

  if (existingSupport) {
    throw new ProjectError('已经支持过该档位', 409);
  }

  const now = getNowISO();
  const supportId = uuidv4();

  db.prepare(`
    INSERT INTO support_records (id, project_id, reward_tier_id, supporter_id, amount, status, created_at)
    VALUES (?, ?, ?, ?, ?, 'active', ?)
  `).run(supportId, rewardTier.projectId, rewardTierId, supporterId, rewardTier.amount, now);

  const newRaised = project.raisedAmount + rewardTier.amount;
  db.prepare('UPDATE projects SET raised_amount = ?, updated_at = ? WHERE id = ?').run(newRaised, now, project.id);

  if (newRaised >= project.targetAmount && canTransitionStatus('funding', 'succeeded')) {
    db.prepare(`
      UPDATE projects
      SET status = 'succeeded', updated_at = ?, succeeded_at = ?
      WHERE id = ?
    `).run(now, now, project.id);
  }

  const supportRow = db.prepare('SELECT * FROM support_records WHERE id = ?').get(supportId) as any;
  return mapSupportRow(supportRow);
}

export function getSupportById(supportId: string): SupportRecord | null {
  const row = db.prepare('SELECT * FROM support_records WHERE id = ?').get(supportId) as any;
  return row ? mapSupportRow(row) : null;
}

export function getSupportsByProjectId(projectId: string): SupportRecord[] {
  const rows = db.prepare(
    "SELECT * FROM support_records WHERE project_id = ? ORDER BY created_at DESC"
  ).all(projectId);
  return rows.map(mapSupportRow);
}

export function getActiveSupportsByProjectId(projectId: string): SupportRecord[] {
  const rows = db.prepare(
    "SELECT * FROM support_records WHERE project_id = ? AND status = 'active' ORDER BY created_at DESC"
  ).all(projectId);
  return rows.map(mapSupportRow);
}

export interface RefundResult {
  support: SupportRecord;
  project: Project;
  message?: string;
}

export function refundSupport(supportId: string): RefundResult {
  const support = getSupportById(supportId);
  if (!support) {
    throw new ProjectError('支持记录不存在', 404);
  }

  if (support.status === 'refunded') {
    throw new ProjectError('该支持已退款', 400);
  }

  const projectRow = db.prepare('SELECT * FROM projects WHERE id = ?').get(support.projectId) as any;
  if (!projectRow) {
    throw new ProjectError('项目不存在', 404);
  }

  const project = mapProjectRow(projectRow);

  if (project.status === 'preparing' || project.status === 'cancelled') {
    throw new ProjectError('该项目状态不允许退款', 400);
  }

  if (project.status === 'succeeded') {
    if (!project.succeededAt) {
      throw new ProjectError('项目状态异常', 500);
    }
    const succeededTime = new Date(project.succeededAt).getTime();
    const now = Date.now();
    const hoursDiff = (now - succeededTime) / (1000 * 60 * 60);

    if (hoursDiff > 72) {
      throw new ProjectError('超出72小时退款期', 400);
    }
  }

  const now = getNowISO();
  const newRaised = project.raisedAmount - support.amount;
  let message: string | undefined;
  let newProjectStatus = project.status;

  db.prepare(`
    UPDATE support_records
    SET status = 'refunded', refunded_at = ?
    WHERE id = ?
  `).run(now, supportId);

  db.prepare('UPDATE projects SET raised_amount = ?, updated_at = ? WHERE id = ?').run(newRaised, now, project.id);

  if (project.status === 'succeeded' && newRaised < project.targetAmount) {
    db.prepare(`
      UPDATE projects
      SET status = 'funding', updated_at = ?, succeeded_at = NULL
      WHERE id = ?
    `).run(now, project.id);
    newProjectStatus = 'funding';
    message = '退款成功，项目因未达到目标金额已回退为筹款中状态';
  } else {
    message = '退款成功';
  }

  const updatedProjectRow = db.prepare('SELECT * FROM projects WHERE id = ?').get(project.id) as any;
  const updatedSupportRow = db.prepare('SELECT * FROM support_records WHERE id = ?').get(supportId) as any;

  return {
    support: mapSupportRow(updatedSupportRow),
    project: mapProjectRow(updatedProjectRow),
    message,
  };
}

export function getSupportsBySupporterAndProject(supporterId: string, projectId: string): SupportRecord[] {
  const rows = db.prepare(
    "SELECT * FROM support_records WHERE supporter_id = ? AND project_id = ? ORDER BY created_at DESC"
  ).all(supporterId, projectId);
  return rows.map(mapSupportRow);
}
