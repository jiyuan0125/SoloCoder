import {
  AnnualBudget, MonthlyCategoryBudget, CreateBudgetRequest, AdjustBudgetRequest,
  CreateDraftBudgetRequest, ConfirmBudgetRequest, BudgetCategory, Month,
  createSuccessResponse, createErrorResponse, ErrorCode
} from '@budget-planner/shared';
import { ParsedRequest, RouteResult } from '../router';
import { 
  addBudget, getBudget, getDepartment, updateBudget, getMonthlyBudget,
  updateMonthlyBudgets, updateBudgetStatus, getBudgetsByYear
} from '../store/dataStore';
import { 
  isValidMonth, isValidCategory, isValidAmount, isValidYear, isValidString,
  parseAsNumber
} from '../utils/validation';
import { calculateTotalAllocated, cloneAnnualBudget } from '../utils/budgetUtils';

function generateId(): string {
  return `budget_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

function isAllocationItem(item: unknown): item is { month: Month; category: BudgetCategory; amount: number } {
  if (!item || typeof item !== 'object') return false;
  const obj = item as Record<string, unknown>;
  return (
    isValidMonth(obj.month) &&
    isValidCategory(obj.category) &&
    isValidAmount(obj.amount)
  );
}

function isAdjustmentItem(item: unknown): item is { month: Month; category: BudgetCategory; newAllocated: number } {
  if (!item || typeof item !== 'object') return false;
  const obj = item as Record<string, unknown>;
  return (
    isValidMonth(obj.month) &&
    isValidCategory(obj.category) &&
    isValidAmount(obj.newAllocated)
  );
}

function validateCreateBudgetRequest(body: unknown): body is CreateBudgetRequest {
  if (!body || typeof body !== 'object') return false;
  const obj = body as Record<string, unknown>;
  
  if (!isValidString(obj.departmentId) || !isValidYear(obj.year) || !isValidAmount(obj.totalAmount)) {
    return false;
  }

  if (!Array.isArray(obj.allocations)) {
    return false;
  }

  for (const alloc of obj.allocations) {
    if (!isAllocationItem(alloc)) {
      return false;
    }
  }

  return true;
}

function validateAdjustBudgetRequest(body: unknown): body is AdjustBudgetRequest {
  if (!body || typeof body !== 'object') return false;
  const obj = body as Record<string, unknown>;
  
  if (!isValidString(obj.departmentId) || !isValidYear(obj.year) || !isValidString(obj.requestedBy)) {
    return false;
  }

  if (!Array.isArray(obj.adjustments)) {
    return false;
  }

  for (const adj of obj.adjustments) {
    if (!isAdjustmentItem(adj)) {
      return false;
    }
  }

  return true;
}

function validateCreateDraftBudgetRequest(body: unknown): body is CreateDraftBudgetRequest {
  if (!body || typeof body !== 'object') return false;
  const obj = body as Record<string, unknown>;
  return isValidYear(obj.newYear) && isValidYear(obj.sourceYear);
}

function validateConfirmBudgetRequest(body: unknown): body is ConfirmBudgetRequest {
  if (!body || typeof body !== 'object') return false;
  const obj = body as Record<string, unknown>;
  return (
    isValidString(obj.departmentId) &&
    isValidYear(obj.year) &&
    isValidString(obj.confirmedBy)
  );
}

export function handleGetBudgets(req: ParsedRequest): RouteResult {
  const departmentId = req.query.departmentId;
  const yearStr = req.query.year;

  if (!departmentId || !yearStr) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Missing required query parameters: departmentId, year'
      ),
    };
  }

  const year = parseAsNumber(yearStr);
  if (year === null || !isValidYear(year)) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid year parameter'
      ),
    };
  }

  const budget = getBudget(departmentId, year);
  if (!budget) {
    return {
      statusCode: 404,
      body: createErrorResponse(
        ErrorCode.BUDGET_NOT_FOUND,
        'Budget not found'
      ),
    };
  }

  return {
    statusCode: 200,
    body: createSuccessResponse(budget),
  };
}

export function handlePostBudgets(req: ParsedRequest): RouteResult {
  if (!validateCreateBudgetRequest(req.body)) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid budget request: required fields: departmentId, year, totalAmount, allocations'
      ),
    };
  }

  const department = getDepartment(req.body.departmentId);
  if (!department) {
    return {
      statusCode: 404,
      body: createErrorResponse(
        ErrorCode.DEPARTMENT_NOT_FOUND,
        'Department not found'
      ),
    };
  }

  const existingBudget = getBudget(req.body.departmentId, req.body.year);
  if (existingBudget) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.BUDGET_ALREADY_EXISTS,
        'Budget already exists for this department and year'
      ),
    };
  }

  const categories: BudgetCategory[] = ['TRAVEL', 'OFFICE', 'EQUIPMENT', 'TRAINING', 'OTHER'];
  const months: Month[] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12];

  const monthlyBudgets: MonthlyCategoryBudget[] = [];
  
  for (const month of months) {
    for (const category of categories) {
      const allocation = req.body.allocations.find(
        (a) => a.month === month && a.category === category
      );
      monthlyBudgets.push({
        month,
        category,
        allocated: allocation ? allocation.amount : 0,
        used: 0,
      });
    }
  }

  const budget: AnnualBudget = {
    id: generateId(),
    departmentId: req.body.departmentId,
    year: req.body.year,
    totalAmount: req.body.totalAmount,
    monthlyBudgets,
    status: 'ACTIVE',
    createdAt: new Date().toISOString(),
  };

  addBudget(budget);

  return {
    statusCode: 201,
    body: createSuccessResponse(budget),
  };
}

export function handlePostBudgetsAdjust(req: ParsedRequest): RouteResult {
  if (!validateAdjustBudgetRequest(req.body)) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid adjustment request'
      ),
    };
  }

  const budget = getBudget(req.body.departmentId, req.body.year);
  if (!budget) {
    return {
      statusCode: 404,
      body: createErrorResponse(
        ErrorCode.BUDGET_NOT_FOUND,
        'Budget not found'
      ),
    };
  }

  if (budget.status !== 'ACTIVE') {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.BUDGET_NOT_ACTIVE,
        'Budget is not active'
      ),
    };
  }

  for (const adjustment of req.body.adjustments) {
    const monthlyBudget = getMonthlyBudget(budget, adjustment.month, adjustment.category);
    if (monthlyBudget && monthlyBudget.used > adjustment.newAllocated) {
      return {
        statusCode: 400,
        body: createErrorResponse(
          ErrorCode.BUDGET_LOWER_THAN_USED,
          'New budget is lower than already used amount',
          {
            month: adjustment.month,
            category: adjustment.category,
            used: monthlyBudget.used,
            newAllocated: adjustment.newAllocated,
            deficit: monthlyBudget.used - adjustment.newAllocated,
          }
        ),
      };
    }
  }

  const clonedBudget = cloneAnnualBudget(budget);
  const updatedMonthlyBudgets = req.body.adjustments.map((adj) => {
    const existing = getMonthlyBudget(clonedBudget, adj.month, adj.category);
    if (!existing) {
      return {
        month: adj.month,
        category: adj.category,
        allocated: adj.newAllocated,
        used: 0,
      } as MonthlyCategoryBudget;
    }
    return {
      ...existing,
      allocated: adj.newAllocated,
    };
  });

  const updatedBudget = updateMonthlyBudgets(clonedBudget, updatedMonthlyBudgets);

  const newTotalAllocated = calculateTotalAllocated(updatedBudget);
  if (newTotalAllocated > budget.totalAmount) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.TOTAL_EXCEEDS_ANNUAL,
        'Total allocated exceeds annual budget',
        {
          currentTotal: newTotalAllocated,
          annualTotal: budget.totalAmount,
        }
      ),
    };
  }

  updateBudget(updatedBudget);

  return {
    statusCode: 200,
    body: createSuccessResponse(updatedBudget),
  };
}

export function handlePostBudgetsDraft(req: ParsedRequest): RouteResult {
  if (!validateCreateDraftBudgetRequest(req.body)) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid draft budget request: required fields: newYear, sourceYear'
      ),
    };
  }

  const { newYear, sourceYear } = req.body;

  if (newYear <= sourceYear) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'New year must be greater than source year'
      ),
    };
  }

  const sourceBudgets = getBudgetsByYear(sourceYear);
  if (sourceBudgets.length === 0) {
    return {
      statusCode: 404,
      body: createErrorResponse(
        ErrorCode.BUDGET_NOT_FOUND,
        'No budgets found for source year'
      ),
    };
  }

  const draftBudgets: AnnualBudget[] = [];

  for (const sourceBudget of sourceBudgets) {
    const existingNewBudget = getBudget(sourceBudget.departmentId, newYear);
    if (existingNewBudget) {
      continue;
    }

    const draftBudget: AnnualBudget = {
      id: generateId(),
      departmentId: sourceBudget.departmentId,
      year: newYear,
      totalAmount: sourceBudget.totalAmount,
      monthlyBudgets: sourceBudget.monthlyBudgets.map((mb) => ({
        ...mb,
        used: 0,
      })),
      status: 'DRAFT',
      createdAt: new Date().toISOString(),
    };

    addBudget(draftBudget);
    draftBudgets.push(draftBudget);
  }

  return {
    statusCode: 201,
    body: createSuccessResponse(draftBudgets),
  };
}

export function handlePostBudgetsConfirm(req: ParsedRequest): RouteResult {
  if (!validateConfirmBudgetRequest(req.body)) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid confirm budget request: required fields: departmentId, year, confirmedBy'
      ),
    };
  }

  const budget = getBudget(req.body.departmentId, req.body.year);
  if (!budget) {
    return {
      statusCode: 404,
      body: createErrorResponse(
        ErrorCode.BUDGET_NOT_FOUND,
        'Budget not found'
      ),
    };
  }

  if (budget.status !== 'DRAFT') {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.BUDGET_NOT_DRAFT,
        'Budget is not in draft status'
      ),
    };
  }

  const confirmedBudget = updateBudgetStatus(
    budget,
    'ACTIVE',
    new Date().toISOString()
  );

  updateBudget(confirmedBudget);

  return {
    statusCode: 200,
    body: createSuccessResponse(confirmedBudget),
  };
}
