import { db } from '../database';
import { Project, ProjectStatus, ProjectApplication, Todo, FundAllocation } from '../types';
import {
  generateId,
  getCurrentTime,
  isValidStatusTransition,
  validateCreditCode,
  validateBudget,
  addDays
} from '../utils';
import { createAuditLog } from './auditService';
import { createTodo, checkOverdueTodos } from './todoService';
import { checkBudgetAlert } from './budgetService';

const MAX_SUBMIT_COUNT = 3;

interface DBProject {
  id: string;
  organization_name: string;
  credit_code: string;
  contact_person: string;
  project_name: string;
  category: string;
  budget_amount: number;
  used_budget: number;
  implementation_period: number;
  status: string;
  submit_count: number;
  created_at: string;
  updated_at: string;
}

function mapProject(dbProject: DBProject): Project {
  return {
    id: dbProject.id,
    organizationName: dbProject.organization_name,
    creditCode: dbProject.credit_code,
    contactPerson: dbProject.contact_person,
    projectName: dbProject.project_name,
    category: dbProject.category,
    budgetAmount: dbProject.budget_amount,
    usedBudget: dbProject.used_budget,
    implementationPeriod: dbProject.implementation_period,
    status: dbProject.status as ProjectStatus,
    submitCount: dbProject.submit_count,
    createdAt: dbProject.created_at,
    updatedAt: dbProject.updated_at
  };
}

export function getProjectById(id: string): Project | null {
  const dbProject = db.prepare('SELECT * FROM projects WHERE id = ?').get(id) as DBProject | undefined;
  return dbProject ? mapProject(dbProject) : null;
}

export function getProjectByCreditCode(creditCode: string): Project | null {
  const dbProject = db.prepare('SELECT * FROM projects WHERE credit_code = ? ORDER BY created_at DESC').get(creditCode) as DBProject | undefined;
  return dbProject ? mapProject(dbProject) : null;
}

export function getAllProjects(): Project[] {
  const dbProjects = db.prepare('SELECT * FROM projects ORDER BY created_at DESC').all() as DBProject[];
  return dbProjects.map(mapProject);
}

export function createProject(app: ProjectApplication, operator: string): { success: boolean; message: string; project?: Project } {
  if (!validateCreditCode(app.creditCode)) {
    return { success: false, message: '统一社会信用代码格式不正确' };
  }

  if (!validateBudget(app.budgetAmount)) {
    return { success: false, message: '预算金额必须大于0' };
  }

  const existingProject = getProjectByCreditCode(app.creditCode);
  if (existingProject && existingProject.submitCount >= MAX_SUBMIT_COUNT) {
    return { success: false, message: '已达最大提交次数' };
  }

  const id = generateId();
  const now = getCurrentTime();
  const submitCount = existingProject ? existingProject.submitCount + 1 : 1;

  const insertStmt = db.prepare(`
    INSERT INTO projects (
      id, organization_name, credit_code, contact_person, project_name,
      category, budget_amount, used_budget, implementation_period, status,
      submit_count, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, '申请中', ?, ?, ?)
  `);

  insertStmt.run(
    id,
    app.organizationName,
    app.creditCode,
    app.contactPerson,
    app.projectName,
    app.category,
    app.budgetAmount,
    app.implementationPeriod,
    submitCount,
    now,
    now
  );

  const project = getProjectById(id)!;

  createAuditLog(
    operator,
    '项目申请',
    `提交项目申请：${app.projectName}`,
    null,
    project
  );

  createTodo(
    id,
    '项目初审',
    `对项目"${app.projectName}"进行初审`,
    addDays(now, 7),
    operator
  );

  return { success: true, message: '申请提交成功', project };
}

export function updateProjectStatus(
  projectId: string,
  newStatus: ProjectStatus,
  operator: string,
  reason?: string
): { success: boolean; message: string; project?: Project } {
  const project = getProjectById(projectId);
  if (!project) {
    return { success: false, message: '项目不存在' };
  }

  if (!isValidStatusTransition(project.status, newStatus)) {
    return { success: false, message: `非法状态跳转：${project.status} -> ${newStatus}` };
  }

  const beforeSnapshot = { ...project };
  const now = getCurrentTime();

  const updateStmt = db.prepare(`
    UPDATE projects SET status = ?, updated_at = ? WHERE id = ?
  `);

  const result = updateStmt.run(newStatus, now, projectId);
  
  if (result.changes === 0) {
    return { success: false, message: '状态更新失败' };
  }

  const updatedProject = getProjectById(projectId)!;

  const content = reason 
    ? `状态变更：${project.status} -> ${newStatus}。原因：${reason}`
    : `状态变更：${project.status} -> ${newStatus}`;

  createAuditLog(
    operator,
    '状态变更',
    content,
    beforeSnapshot,
    updatedProject
  );

  if (newStatus === '初审') {
    createTodo(
      projectId,
      '专家评审',
      `对项目"${updatedProject.projectName}"进行专家评审`,
      addDays(now, 14),
      operator
    );
  } else if (newStatus === '整改') {
    createTodo(
      projectId,
      '项目整改',
      `根据评审意见对项目"${updatedProject.projectName}"进行整改`,
      addDays(now, 30),
      operator
    );
  }

  return { success: true, message: '状态更新成功', project: updatedProject };
}

export function allocateFunds(
  projectId: string,
  amount: number,
  operator: string
): { success: boolean; message: string; allocation?: FundAllocation } {
  const project = getProjectById(projectId);
  if (!project) {
    return { success: false, message: '项目不存在' };
  }

  if (project.status !== '批准') {
    return { success: false, message: '只有已批准的项目才能拨付资金' };
  }

  const remaining = project.budgetAmount - project.usedBudget;
  if (amount > remaining) {
    return { success: false, message: `拨付金额超过剩余预算，剩余：${remaining}` };
  }

  const beforeSnapshot = { ...project };
  const now = getCurrentTime();
  const allocationId = generateId();
  const newUsedBudget = project.usedBudget + amount;

  const updateProject = db.prepare(`
    UPDATE projects SET used_budget = ?, updated_at = ? WHERE id = ?
  `);

  const insertAllocation = db.prepare(`
    INSERT INTO fund_allocations (id, project_id, amount, operator, created_at)
    VALUES (?, ?, ?, ?, ?)
  `);

  const transaction = db.transaction(() => {
    updateProject.run(newUsedBudget, now, projectId);
    insertAllocation.run(allocationId, projectId, amount, operator, now);
  });

  transaction();

  const updatedProject = getProjectById(projectId)!;

  createAuditLog(
    operator,
    '资金拨付',
    `向项目"${project.projectName}"拨付资金：${amount}元`,
    beforeSnapshot,
    updatedProject
  );

  checkBudgetAlert(projectId, operator);

  const allocation: FundAllocation = {
    id: allocationId,
    projectId,
    amount,
    operator,
    createdAt: now
  };

  return { success: true, message: '资金拨付成功', allocation };
}

export function getFundAllocations(projectId: string): FundAllocation[] {
  return db.prepare(`
    SELECT * FROM fund_allocations WHERE project_id = ? ORDER BY created_at DESC
  `).all(projectId) as FundAllocation[];
}

export function processOverdueTasks(): number {
  return checkOverdueTodos();
}
