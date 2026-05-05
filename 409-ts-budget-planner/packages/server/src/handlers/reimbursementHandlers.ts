import {
  ReimbursementRecord, SubmitReimbursementRequest, MonthlyCategoryBudget,
  createSuccessResponse, createErrorResponse, ErrorCode
} from '@budget-planner/shared';
import { ParsedRequest, RouteResult } from '../router';
import { 
  getBudget, getMonthlyBudget, updateMonthlyBudgets, updateBudget, addReimbursement
} from '../store/dataStore';
import { 
  isValidMonth, isValidCategory, isValidAmount, isValidString, isValidYear
} from '../utils/validation';
import { cloneAnnualBudget, getRemainingBudget } from '../utils/budgetUtils';

function generateId(): string {
  return `reimb_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

function isReimbursementItem(item: unknown): item is { category: string; amount: number } {
  if (!item || typeof item !== 'object') return false;
  const obj = item as Record<string, unknown>;
  return (
    isValidCategory(obj.category) &&
    isValidAmount(obj.amount) &&
    obj.amount > 0
  );
}

function validateSubmitReimbursementRequest(body: unknown): body is SubmitReimbursementRequest {
  if (!body || typeof body !== 'object') return false;
  const obj = body as Record<string, unknown>;
  
  if (
    !isValidString(obj.departmentId) ||
    !isValidYear(obj.year) ||
    !isValidMonth(obj.month) ||
    !isValidString(obj.description) ||
    !isValidString(obj.requestedBy)
  ) {
    return false;
  }

  if (!Array.isArray(obj.items) || obj.items.length === 0) {
    return false;
  }

  for (const item of obj.items) {
    if (!isReimbursementItem(item)) {
      return false;
    }
  }

  return true;
}

export function handlePostReimbursements(req: ParsedRequest): RouteResult {
  if (!validateSubmitReimbursementRequest(req.body)) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid reimbursement request: required fields: departmentId, year, month, items, description, requestedBy'
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

  const insufficientBudgets: Array<{
    category: string;
    remaining: number;
    requested: number;
    deficit: number;
  }> = [];

  for (const item of req.body.items) {
    const monthlyBudget = getMonthlyBudget(budget, req.body.month, item.category);
    if (!monthlyBudget) {
      continue;
    }

    const remaining = getRemainingBudget(monthlyBudget);
    if (item.amount > remaining) {
      insufficientBudgets.push({
        category: item.category,
        remaining,
        requested: item.amount,
        deficit: item.amount - remaining,
      });
    }
  }

  if (insufficientBudgets.length > 0) {
    const record: ReimbursementRecord = {
      id: generateId(),
      departmentId: req.body.departmentId,
      year: req.body.year,
      month: req.body.month,
      items: req.body.items,
      description: req.body.description,
      requestedBy: req.body.requestedBy,
      requestedAt: new Date().toISOString(),
      status: 'REJECTED',
      rejectedReason: 'Insufficient budget',
      processedAt: new Date().toISOString(),
    };

    addReimbursement(record);

    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INSUFFICIENT_BUDGET,
        'Insufficient budget for one or more categories',
        {
          insufficientBudgets,
          reimbursementId: record.id,
        }
      ),
    };
  }

  const clonedBudget = cloneAnnualBudget(budget);
  const updatedBudgets: MonthlyCategoryBudget[] = [];

  for (const item of req.body.items) {
    const monthlyBudget = getMonthlyBudget(clonedBudget, req.body.month, item.category);
    if (monthlyBudget) {
      updatedBudgets.push({
        ...monthlyBudget,
        used: monthlyBudget.used + item.amount,
      });
    }
  }

  const updatedBudget = updateMonthlyBudgets(clonedBudget, updatedBudgets);
  updateBudget(updatedBudget);

  const record: ReimbursementRecord = {
    id: generateId(),
    departmentId: req.body.departmentId,
    year: req.body.year,
    month: req.body.month,
    items: req.body.items,
    description: req.body.description,
    requestedBy: req.body.requestedBy,
    requestedAt: new Date().toISOString(),
    status: 'APPROVED',
    processedAt: new Date().toISOString(),
  };

  addReimbursement(record);

  return {
    statusCode: 200,
    body: createSuccessResponse({
      reimbursementId: record.id,
      status: 'APPROVED' as const,
    }),
  };
}
