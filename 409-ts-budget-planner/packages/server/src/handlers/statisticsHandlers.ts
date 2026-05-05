import {
  UsageStatistic, TrendDataPoint, QueryUsageRequest, QueryTrendRequest,
  Month, BudgetCategory, Quarter,
  createSuccessResponse
} from '@budget-planner/shared';
import { ParsedRequest, RouteResult } from '../router';
import { 
  getAllBudgets, getDepartment
} from '../store/dataStore';
import { 
  parseAsNumber, parseAsCategory, isValidYear, parseAsMonth
} from '../utils/validation';
import { 
  getMonthsInQuarter
} from '../utils/budgetUtils';

function validateQueryUsageRequest(query: Record<string, string | undefined>): QueryUsageRequest {
  const result: QueryUsageRequest = {};

  if (query.departmentId) {
    result.departmentId = query.departmentId;
  }

  if (query.category) {
    const cat = parseAsCategory(query.category);
    if (cat !== null) {
      result.category = cat;
    }
  }

  if (query.startYear) {
    const year = parseAsNumber(query.startYear);
    if (year !== null && isValidYear(year)) {
      result.startYear = year;
    }
  }

  if (query.startMonth) {
    const month = parseAsMonth(query.startMonth);
    if (month !== null) {
      result.startMonth = month;
    }
  }

  if (query.endYear) {
    const year = parseAsNumber(query.endYear);
    if (year !== null && isValidYear(year)) {
      result.endYear = year;
    }
  }

  if (query.endMonth) {
    const month = parseAsMonth(query.endMonth);
    if (month !== null) {
      result.endMonth = month;
    }
  }

  return result;
}

function validateQueryTrendRequest(query: Record<string, string | undefined>): QueryTrendRequest {
  const result: QueryTrendRequest = {};

  if (query.departmentId) {
    result.departmentId = query.departmentId;
  }

  if (query.category) {
    const cat = parseAsCategory(query.category);
    if (cat !== null) {
      result.category = cat;
    }
  }

  if (query.startYear) {
    const year = parseAsNumber(query.startYear);
    if (year !== null && isValidYear(year)) {
      result.startYear = year;
    }
  }

  if (query.startMonth) {
    const month = parseAsMonth(query.startMonth);
    if (month !== null) {
      result.startMonth = month;
    }
  }

  if (query.endYear) {
    const year = parseAsNumber(query.endYear);
    if (year !== null && isValidYear(year)) {
      result.endYear = year;
    }
  }

  if (query.endMonth) {
    const month = parseAsMonth(query.endMonth);
    if (month !== null) {
      result.endMonth = month;
    }
  }

  if (query.interval === 'monthly' || query.interval === 'quarterly') {
    result.interval = query.interval;
  }

  return result;
}

function isInRange(
  year: number,
  month: Month,
  startYear?: number,
  startMonth?: Month,
  endYear?: number,
  endMonth?: Month
): boolean {
  if (startYear !== undefined) {
    if (year < startYear) return false;
    if (year === startYear && startMonth !== undefined && month < startMonth) return false;
  }

  if (endYear !== undefined) {
    if (year > endYear) return false;
    if (year === endYear && endMonth !== undefined && month > endMonth) return false;
  }

  return true;
}

export function handleGetStatisticsUsage(req: ParsedRequest): RouteResult {
  const query = validateQueryUsageRequest(req.query);

  const budgets = getAllBudgets();
  const results: UsageStatistic[] = [];

  for (const budget of budgets) {
    if (query.departmentId && budget.departmentId !== query.departmentId) {
      continue;
    }

    const department = getDepartment(budget.departmentId);
    if (!department) {
      continue;
    }

    for (const mb of budget.monthlyBudgets) {
      if (query.category && mb.category !== query.category) {
        continue;
      }

      if (!isInRange(
        budget.year,
        mb.month,
        query.startYear,
        query.startMonth,
        query.endYear,
        query.endMonth
      )) {
        continue;
      }

      const usageRate = mb.allocated > 0 ? Math.floor((mb.used * 10000) / mb.allocated) : 0;

      results.push({
        departmentId: budget.departmentId,
        departmentName: department.name,
        category: mb.category,
        year: budget.year,
        month: mb.month,
        allocated: mb.allocated,
        used: mb.used,
        usageRate,
      });
    }
  }

  return {
    statusCode: 200,
    body: createSuccessResponse(results),
  };
}

export function handleGetStatisticsTrend(req: ParsedRequest): RouteResult {
  const query = validateQueryTrendRequest(req.query);
  const interval = query.interval || 'monthly';

  const budgets = getAllBudgets();
  const results: TrendDataPoint[] = [];

  if (interval === 'monthly') {
    for (const budget of budgets) {
      if (query.departmentId && budget.departmentId !== query.departmentId) {
        continue;
      }

      const department = getDepartment(budget.departmentId);
      if (!department) {
        continue;
      }

      for (const mb of budget.monthlyBudgets) {
        if (query.category && mb.category !== query.category) {
          continue;
        }

        if (!isInRange(
          budget.year,
          mb.month,
          query.startYear,
          query.startMonth,
          query.endYear,
          query.endMonth
        )) {
          continue;
        }

        const usageRate = mb.allocated > 0 ? Math.floor((mb.used * 10000) / mb.allocated) : 0;

        results.push({
          period: `${budget.year}-${String(mb.month).padStart(2, '0')}`,
          departmentId: budget.departmentId,
          departmentName: department.name,
          category: mb.category,
          allocated: mb.allocated,
          used: mb.used,
          usageRate,
        });
      }
    }
  } else {
    for (const budget of budgets) {
      if (query.departmentId && budget.departmentId !== query.departmentId) {
        continue;
      }

      const department = getDepartment(budget.departmentId);
      if (!department) {
        continue;
      }

      const categories: BudgetCategory[] = ['TRAVEL', 'OFFICE', 'EQUIPMENT', 'TRAINING', 'OTHER'];
      const quarters: Quarter[] = [1, 2, 3, 4];

      for (const quarter of quarters) {
        const months = getMonthsInQuarter(quarter);
        
        if (months.length > 0 && !isInRange(
          budget.year,
          months[0],
          query.startYear,
          query.startMonth,
          query.endYear,
          query.endMonth
        )) {
          continue;
        }

        for (const category of categories) {
          if (query.category && category !== query.category) {
            continue;
          }

          let quarterlyAllocated = 0;
          let quarterlyUsed = 0;

          for (const mb of budget.monthlyBudgets) {
            if (months.includes(mb.month) && mb.category === category) {
              quarterlyAllocated += mb.allocated;
              quarterlyUsed += mb.used;
            }
          }

          const usageRate = quarterlyAllocated > 0 ? Math.floor((quarterlyUsed * 10000) / quarterlyAllocated) : 0;

          results.push({
            period: `${budget.year}-Q${quarter}`,
            departmentId: budget.departmentId,
            departmentName: department.name,
            category,
            allocated: quarterlyAllocated,
            used: quarterlyUsed,
            usageRate,
          });
        }
      }
    }
  }

  return {
    statusCode: 200,
    body: createSuccessResponse(results),
  };
}
