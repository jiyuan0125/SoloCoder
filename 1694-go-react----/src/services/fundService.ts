import db from '../database';
import { ProjectStatus, InspectionResult, BudgetCategory } from '../types';
import { AppError } from '../middleware/errorHandler';

const MAX_PERCENTAGE = 100;

const getExistingAllocations = (projectId: number) => {
  return db.prepare(`
    SELECT stage, percentage FROM fund_allocations WHERE project_id = ?
  `).all(projectId) as any[];
};

const getTotalAllocatedPercentage = (projectId: number) => {
  const result = db.prepare(`
    SELECT COALESCE(SUM(percentage), 0) as total FROM fund_allocations WHERE project_id = ?
  `).get(projectId) as any;
  return result.total as number;
};

export const allocateFund = (
  projectId: number,
  stage: 'initial' | 'mid_term' | 'final',
  percentage: number,
  amount: number,
  allocationDate: string,
  paymentMethod: string,
  receivingAccount: string
) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);

  if (stage === 'initial') {
    if (project.status !== ProjectStatus.APPROVED) {
      throw new AppError('只有已立项的项目才能拨付启动资金', 400);
    }
    if (percentage !== 30) {
      throw new AppError('启动资金比例必须为30%', 400);
    }
  } else if (stage === 'mid_term') {
    if (project.status !== ProjectStatus.MID_TERM_INSPECTION_PASSED) {
      throw new AppError('只有中期验收通过才能拨付中期资金', 400);
    }
    if (percentage !== 40) {
      throw new AppError('中期资金比例必须为40%', 400);
    }
  } else if (stage === 'final') {
    if (project.status !== ProjectStatus.FINAL_INSPECTION_PASSED) {
      throw new AppError('只有终期验收通过才能拨付终期资金', 400);
    }
    if (percentage !== 30) {
      throw new AppError('终期资金比例必须为30%', 400);
    }
  }

  const existing = getExistingAllocations(projectId);
  const hasStage = existing.some((a: any) => a.stage === stage);
  if (hasStage) {
    throw new AppError(`该阶段资金已拨付`, 400);
  }

  const currentTotal = getTotalAllocatedPercentage(projectId);
  if (currentTotal + percentage > MAX_PERCENTAGE) {
    throw new AppError(`拨付比例之和不能超过100%，当前已拨付${currentTotal}%`, 400);
  }

  if (amount <= 0) {
    throw new AppError('拨付金额必须大于零', 400);
  }

  const expectedAmount = Math.round(project.total_budget * percentage / 100);
  if (amount !== expectedAmount) {
    throw new AppError(`拨付金额不正确，应为总预算的${percentage}%（${expectedAmount}分）`, 400);
  }

  const now = new Date().toISOString();
  const transaction = db.transaction(() => {
    db.prepare(`
      INSERT INTO fund_allocations (project_id, stage, amount, percentage, allocation_date, 
                                     payment_method, receiving_account, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `).run(projectId, stage, amount, percentage, allocationDate, paymentMethod, receivingAccount, now);

    let nextStatus = project.status;
    if (stage === 'initial') {
      nextStatus = ProjectStatus.INITIAL_FUND_ALLOCATED;
    } else if (stage === 'mid_term') {
      nextStatus = ProjectStatus.MID_TERM_FUND_ALLOCATED;
    } else if (stage === 'final') {
      nextStatus = ProjectStatus.FINAL_FUND_ALLOCATED;
    }

    db.prepare('UPDATE projects SET status = ?, updated_at = ? WHERE id = ?')
      .run(nextStatus, now, projectId);
  });

  transaction();
  return true;
};

export const recordExpenditure = (
  projectId: number,
  category: BudgetCategory,
  amount: number,
  description: string,
  expenditureDate: string
) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);

  if (project.status === ProjectStatus.TERMINATED) {
    throw new AppError('项目已终止', 400);
  }

  if (amount <= 0) {
    throw new AppError('支出金额必须大于零', 400);
  }

  const budget = db.prepare(`
    SELECT construction, equipment, labor, other FROM project_budgets WHERE project_id = ?
  `).get(projectId) as any;

  if (!budget) {
    throw new AppError('项目预算不存在', 400);
  }

  const categoryBudget = budget[category];
  const totalSpentResult = db.prepare(`
    SELECT COALESCE(SUM(amount), 0) as total FROM project_expenditures 
    WHERE project_id = ? AND category = ?
  `).get(projectId, category) as any;
  const totalSpent = totalSpentResult.total as number;

  if (totalSpent + amount > categoryBudget) {
    const remaining = categoryBudget - totalSpent;
    throw new AppError(`支出超出预算，该类别剩余预算为${remaining}分`, 400);
  }

  const now = new Date().toISOString();
  db.prepare(`
    INSERT INTO project_expenditures (project_id, category, amount, description, expenditure_date, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
  `).run(projectId, category, amount, description, expenditureDate, now);

  return true;
};

export const submitBudgetAdjustment = (
  projectId: number,
  fromCategory: BudgetCategory,
  toCategory: BudgetCategory,
  amount: number,
  reason: string
) => {
  const project = db.prepare('SELECT * FROM projects WHERE id = ?').get(projectId) as any;
  if (!project) throw new AppError('项目不存在', 404);

  if (project.status === ProjectStatus.TERMINATED) {
    throw new AppError('项目已终止', 400);
  }

  if (fromCategory === toCategory) {
    throw new AppError('调出和调入类别不能相同', 400);
  }

  if (amount <= 0) {
    throw new AppError('调剂金额必须大于零', 400);
  }

  const budget = db.prepare(`
    SELECT construction, equipment, labor, other FROM project_budgets WHERE project_id = ?
  `).get(projectId) as any;

  if (!budget) {
    throw new AppError('项目预算不存在', 400);
  }

  const fromBudget = budget[fromCategory];
  const totalSpentResult = db.prepare(`
    SELECT COALESCE(SUM(amount), 0) as total FROM project_expenditures 
    WHERE project_id = ? AND category = ?
  `).get(projectId, fromCategory) as any;
  const totalSpent = totalSpentResult.total as number;

  const available = fromBudget - totalSpent;
  if (amount > available) {
    throw new AppError(`调出金额超出可用余额，该类别可用余额为${available}分`, 400);
  }

  const now = new Date().toISOString();
  db.prepare(`
    INSERT INTO budget_adjustment_requests (project_id, from_category, to_category, amount, 
                                            reason, status, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `).run(projectId, fromCategory, toCategory, amount, reason, 'pending', now);

  return true;
};

export const processBudgetAdjustment = (
  requestId: number,
  status: 'approved' | 'rejected',
  approver: string
) => {
  const request = db.prepare(`
    SELECT * FROM budget_adjustment_requests WHERE id = ?
  `).get(requestId) as any;

  if (!request) throw new AppError('调剂申请不存在', 404);
  if (request.status !== 'pending') {
    throw new AppError('该申请已处理', 400);
  }

  const now = new Date().toISOString();

  if (status === 'approved') {
    const transaction = db.transaction(() => {
      const budget = db.prepare(`
        SELECT construction, equipment, labor, other FROM project_budgets WHERE project_id = ?
      `).get(request.project_id) as any;

      const fromKey = request.from_category;
      const toKey = request.to_category;
      const newFrom = budget[fromKey] - request.amount;
      const newTo = budget[toKey] + request.amount;

      if (newFrom < 0 || newTo < 0) {
        throw new AppError('调剂后预算不能为负', 400);
      }

      const updateBudget = db.prepare(`
        UPDATE project_budgets SET 
          construction = CASE WHEN 'construction' = ? THEN ? ELSE construction END,
          equipment = CASE WHEN 'equipment' = ? THEN ? ELSE equipment END,
          labor = CASE WHEN 'labor' = ? THEN ? ELSE labor END,
          other = CASE WHEN 'other' = ? THEN ? ELSE other END
        WHERE project_id = ?
      `);

      updateBudget.run(fromKey, newFrom, fromKey, newFrom, fromKey, newFrom, fromKey, newFrom, request.project_id);
      updateBudget.run(toKey, newTo, toKey, newTo, toKey, newTo, toKey, newTo, request.project_id);

      db.prepare(`
        UPDATE budget_adjustment_requests 
        SET status = ?, approver = ?, approved_at = ?
        WHERE id = ?
      `).run('approved', approver, now, requestId);
    });

    transaction();
  } else {
    db.prepare(`
      UPDATE budget_adjustment_requests 
      SET status = ?, approver = ?, approved_at = ?
      WHERE id = ?
    `).run('rejected', approver, now, requestId);
  }

  return true;
};
