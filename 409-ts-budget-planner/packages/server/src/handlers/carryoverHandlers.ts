import {
  CarryoverRecord, RequestCarryoverRequest,
  createSuccessResponse, createErrorResponse, ErrorCode
} from '@budget-planner/shared';
import { ParsedRequest, RouteResult } from '../router';
import { 
  getBudget, updateBudget, addCarryover
} from '../store/dataStore';
import { 
  isValidQuarter, isValidAmount, isValidString, isValidYear
} from '../utils/validation';
import { 
  calculateQuarterTotal, calculateQuarterCarryoverLimit, 
  addCarryoverToBudget, cloneAnnualBudget, getMonthsInQuarter, getRemainingBudget
} from '../utils/budgetUtils';

function generateId(): string {
  return `carryover_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

function validateRequestCarryoverRequest(body: unknown): body is RequestCarryoverRequest {
  if (!body || typeof body !== 'object') return false;
  const obj = body as Record<string, unknown>;
  return (
    isValidString(obj.departmentId) &&
    isValidYear(obj.year) &&
    isValidQuarter(obj.fromQuarter) &&
    isValidQuarter(obj.toQuarter) &&
    isValidAmount(obj.amount) &&
    (obj.amount as number) > 0 &&
    isValidString(obj.requestedBy)
  );
}

export function handlePostCarryovers(req: ParsedRequest): RouteResult {
  if (!validateRequestCarryoverRequest(req.body)) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid carryover request: required fields: departmentId, year, fromQuarter, toQuarter, amount, requestedBy'
      ),
    };
  }

  const { departmentId, year, fromQuarter, toQuarter, amount, requestedBy } = req.body;

  if (toQuarter !== fromQuarter + 1) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_QUARTER_TRANSITION,
        'Carryover can only be to the next quarter',
        {
          fromQuarter,
          toQuarter,
        }
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

  if (budget.status !== 'ACTIVE') {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.BUDGET_NOT_ACTIVE,
        'Budget is not active'
      ),
    };
  }

  const { allocated: quarterAllocated, used: quarterUsed } = calculateQuarterTotal(budget, fromQuarter);
  const quarterRemaining = quarterAllocated - quarterUsed;

  if (amount > quarterRemaining) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INSUFFICIENT_BUDGET,
        'Requested carryover amount exceeds remaining budget in the quarter',
        {
          requested: amount,
          available: quarterRemaining,
          deficit: amount - quarterRemaining,
        }
      ),
    };
  }

  const carryoverLimit = calculateQuarterCarryoverLimit(budget, fromQuarter);
  if (amount > carryoverLimit) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.CARRYOVER_EXCEEDS_LIMIT,
        'Carryover amount exceeds 20% of quarterly total',
        {
          requested: amount,
          limit: carryoverLimit,
          quarterlyTotal: quarterAllocated,
        }
      ),
    };
  }

  const fromMonths = getMonthsInQuarter(fromQuarter);
  let totalAvailableToCarryover = 0;
  
  for (const mb of budget.monthlyBudgets) {
    if (fromMonths.includes(mb.month)) {
      totalAvailableToCarryover += getRemainingBudget(mb);
    }
  }

  if (amount > totalAvailableToCarryover) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INSUFFICIENT_BUDGET,
        'Insufficient remaining budget to carry over',
        {
          requested: amount,
          available: totalAvailableToCarryover,
        }
      ),
    };
  }

  const clonedBudget = cloneAnnualBudget(budget);
  const updatedBudget = addCarryoverToBudget(clonedBudget, fromQuarter, toQuarter, amount);
  
  updateBudget(updatedBudget);

  const record: CarryoverRecord = {
    id: generateId(),
    departmentId,
    year,
    fromQuarter,
    toQuarter,
    amount,
    requestedBy,
    requestedAt: new Date().toISOString(),
    status: 'APPROVED',
    processedAt: new Date().toISOString(),
  };

  addCarryover(record);

  return {
    statusCode: 200,
    body: createSuccessResponse({
      carryoverId: record.id,
      status: 'APPROVED' as const,
    }),
  };
}
