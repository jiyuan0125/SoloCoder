import db from '../database';
import { 
  ProjectStatus, 
  BudgetCategory, 
  ReviewResult, 
  InspectionResult,
  BudgetBreakdown 
} from '../types';
import { AppError } from '../middleware/errorHandler';

const MAX_RESUBMISSIONS = 3;

export const createProject = (data: {
  name: string;
  category: string;
  location: string;
  totalBudget: number;
  implementationPeriod: string;
  budgetBreakdown: BudgetBreakdown;
}) => {
  const { name, category, location, totalBudget, implementationPeriod, budgetBreakdown } = data;

  if (totalBudget <= 0) {
    throw new AppError('总预算必须大于零', 400);
  }

  const budgetSum = budgetBreakdown.construction + budgetBreakdown.equipment + 
                    budgetBreakdown.labor + budgetBreakdown.other;
  if (budgetSum !== totalBudget) {
    const diff = budgetSum - totalBudget;
    throw new AppError(`预算分项之和必须等于总预算，差异金额：${diff} 分`, 400);
  }

  if (budgetBreakdown.construction < 0 || budgetBreakdown.equipment < 0 ||
      budgetBreakdown.labor < 0 || budgetBreakdown.other < 0) {
    throw new AppError('预算分项不能为负数', 400);
  }

  const now = new Date().toISOString();
  const transaction = db.transaction(() => {
    const existing = db.prepare('SELECT id FROM projects WHERE name = ?').get(name);
    if (existing) {
      throw new AppError('项目名称已存在', 409);
    }

    const insertProject = db.prepare(`
      INSERT INTO projects (name, category, location, total_budget, implementation_period, 
                            status, resubmission_count, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    const result = insertProject.run(
      name, category, location, totalBudget, implementationPeriod,
      ProjectStatus.TOWNSHIP_REVIEW_PENDING, 0, now, now
    );
    const projectId = result.lastInsertRowid as number;

    const insertBudget = db.prepare(`
      INSERT INTO project_budgets (project_id, construction, equipment, labor, other, created_at)
      VALUES (?, ?, ?, ?, ?, ?)
    `);
    insertBudget.run(
      projectId,
      budgetBreakdown.construction,
      budgetBreakdown.equipment,
      budgetBreakdown.labor,
      budgetBreakdown.other,
      now
    );

    return projectId;
  });

  return transaction();
};

export const updateProject = (projectId: number, data: {
  name?: string;
  category?: string;
  location?: string;
  totalBudget?: number;
  implementationPeriod?: string;
  budgetBreakdown?: BudgetBreakdown;
}) => {
  const project = db.prepare(`
    SELECT * FROM projects WHERE id = ?
  `).get(projectId) as any;

  if (!project) {
    throw new AppError('项目不存在', 404);
  }

  if (project.status !== ProjectStatus.TOWNSHIP_REVIEW_REJECTED &&
      project.status !== ProjectStatus.COUNTY_REVIEW_REJECTED) {
    throw new AppError('只有审核不通过的项目才能修改', 400);
  }

  if (project.resubmission_count >= MAX_RESUBMISSIONS) {
    throw new AppError(`每个项目最多允许重新提交 ${MAX_RESUBMISSIONS} 次`, 400);
  }

  const { name, category, location, totalBudget, implementationPeriod, budgetBreakdown } = data;

  if (name && name !== project.name) {
    const existing = db.prepare('SELECT id FROM projects WHERE name = ? AND id != ?').get(name, projectId);
    if (existing) {
      throw new AppError('项目名称已存在', 409);
    }
  }

  if (totalBudget !== undefined && totalBudget <= 0) {
    throw new AppError('总预算必须大于零', 400);
  }

  const finalTotal = totalBudget ?? project.total_budget;

  if (budgetBreakdown) {
    const budgetSum = budgetBreakdown.construction + budgetBreakdown.equipment + 
                      budgetBreakdown.labor + budgetBreakdown.other;
    if (budgetSum !== finalTotal) {
      const diff = budgetSum - finalTotal;
      throw new AppError(`预算分项之和必须等于总预算，差异金额：${diff} 分`, 400);
    }

    if (budgetBreakdown.construction < 0 || budgetBreakdown.equipment < 0 ||
        budgetBreakdown.labor < 0 || budgetBreakdown.other < 0) {
      throw new AppError('预算分项不能为负数', 400);
    }
  }

  const now = new Date().toISOString();
  const nextStatus = project.status === ProjectStatus.TOWNSHIP_REVIEW_REJECTED 
    ? ProjectStatus.TOWNSHIP_REVIEW_PENDING 
    : ProjectStatus.COUNTY_REVIEW_PENDING;

  const transaction = db.transaction(() => {
    const updateProject = db.prepare(`
      UPDATE projects 
      SET name = COALESCE(?, name),
          category = COALESCE(?, category),
          location = COALESCE(?, location),
          total_budget = COALESCE(?, total_budget),
          implementation_period = COALESCE(?, implementation_period),
          status = ?,
          resubmission_count = resubmission_count + 1,
          updated_at = ?
      WHERE id = ?
    `);
    updateProject.run(name, category, location, totalBudget, implementationPeriod, nextStatus, now, projectId);

    if (budgetBreakdown) {
      const updateBudget = db.prepare(`
        UPDATE project_budgets 
        SET construction = ?, equipment = ?, labor = ?, other = ?
        WHERE project_id = ?
      `);
      updateBudget.run(
        budgetBreakdown.construction,
        budgetBreakdown.equipment,
        budgetBreakdown.labor,
        budgetBreakdown.other,
        projectId
      );
    }
  });

  transaction();
  return true;
};

export const submitTownshipReview = (projectId: number) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);
  if (project.status !== ProjectStatus.SUBMITTED) {
    throw new AppError('只有已提交的项目才能进入乡镇初审', 400);
  }

  const now = new Date().toISOString();
  db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
    .run(ProjectStatus.TOWNSHIP_REVIEW_PENDING, now, projectId);
  return true;
};

export const processTownshipReview = (projectId: number, result: string, comments: string | null, reviewer: string) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);
  if (project.status !== ProjectStatus.TOWNSHIP_REVIEW_PENDING) {
    throw new AppError('项目不在乡镇初审状态', 400);
  }

  const isValidResult = result === ReviewResult.PASSED || result === ReviewResult.REJECTED;
  if (!isValidResult) {
    throw new AppError('审核结果无效', 400);
  }

  const now = new Date().toISOString();
  const nextStatus = result === ReviewResult.PASSED 
    ? ProjectStatus.COUNTY_REVIEW_PENDING 
    : ProjectStatus.TOWNSHIP_REVIEW_REJECTED;

  const transaction = db.transaction(() => {
    db.prepare(`
      INSERT INTO review_records (project_id, review_level, result, comments, reviewer, reviewed_at, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `).run(projectId, 'township', result, comments, reviewer, now, now);

    db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
      .run(nextStatus, now, projectId);
  });

  transaction();
  return true;
};

export const processCountyReview = (projectId: number, result: string, comments: string | null, reviewer: string) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);
  if (project.status !== ProjectStatus.COUNTY_REVIEW_PENDING) {
    throw new AppError('项目不在县级复审状态', 400);
  }

  const isValidResult = result === ReviewResult.PASSED || result === ReviewResult.REJECTED;
  if (!isValidResult) {
    throw new AppError('审核结果无效', 400);
  }

  const now = new Date().toISOString();
  const nextStatus = result === ReviewResult.PASSED 
    ? ProjectStatus.APPROVED 
    : ProjectStatus.COUNTY_REVIEW_REJECTED;

  const transaction = db.transaction(() => {
    db.prepare(`
      INSERT INTO review_records (project_id, review_level, result, comments, reviewer, reviewed_at, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `).run(projectId, 'county', result, comments, reviewer, now, now);

    db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
      .run(nextStatus, now, projectId);
  });

  transaction();
  return true;
};

export const getAllProjects = () => {
  const projects = db.prepare(`
    SELECT 
      id, name, category, location, total_budget as totalBudget, 
      implementation_period as implementationPeriod, status, 
      resubmission_count as resubmissionCount, created_at as createdAt,
      updated_at as updatedAt
    FROM projects
  `).all();
  return projects;
};

export const getProjectById = (projectId: number) => {
  const project = db.prepare(`
    SELECT 
      id, name, category, location, total_budget as totalBudget, 
      implementation_period as implementationPeriod, status, 
      resubmission_count as resubmissionCount, created_at as createdAt,
      updated_at as updatedAt
    FROM projects WHERE id = ?
  `).get(projectId);

  if (!project) throw new AppError('项目不存在', 404);

  const budget = db.prepare(`
    SELECT construction, equipment, labor, other
    FROM project_budgets WHERE project_id = ?
  `).get(projectId);

  const allocations = db.prepare(`
    SELECT id, stage, amount, percentage, allocation_date as allocationDate,
           payment_method as paymentMethod, receiving_account as receivingAccount
    FROM fund_allocations WHERE project_id = ?
  `).all(projectId);

  return { ...project as any, budget, allocations };
};
