import {
  AnnualBudget, MonthlyCategoryBudget, Month, Quarter
} from '@budget-planner/shared';

export function getMonthsInQuarter(quarter: Quarter): Month[] {
  switch (quarter) {
    case 1: return [1, 2, 3];
    case 2: return [4, 5, 6];
    case 3: return [7, 8, 9];
    case 4: return [10, 11, 12];
  }
}

export function getQuarterForMonth(month: Month): Quarter {
  if (month <= 3) return 1;
  if (month <= 6) return 2;
  if (month <= 9) return 3;
  return 4;
}

export function calculateQuarterTotal(
  budget: AnnualBudget,
  quarter: Quarter
): { allocated: number; used: number } {
  const months = getMonthsInQuarter(quarter);
  let totalAllocated = 0;
  let totalUsed = 0;

  for (const monthlyBudget of budget.monthlyBudgets) {
    if (months.includes(monthlyBudget.month)) {
      totalAllocated += monthlyBudget.allocated;
      totalUsed += monthlyBudget.used;
    }
  }

  return { allocated: totalAllocated, used: totalUsed };
}

export function calculateQuarterCarryoverLimit(
  budget: AnnualBudget,
  quarter: Quarter
): number {
  const { allocated } = calculateQuarterTotal(budget, quarter);
  return Math.floor(allocated * 20 / 100);
}

export function calculateTotalAllocated(budget: AnnualBudget): number {
  return budget.monthlyBudgets.reduce((sum, b) => sum + b.allocated, 0);
}

export function calculateTotalUsed(budget: AnnualBudget): number {
  return budget.monthlyBudgets.reduce((sum, b) => sum + b.used, 0);
}

export function getRemainingBudget(monthlyBudget: MonthlyCategoryBudget): number {
  return monthlyBudget.allocated - monthlyBudget.used;
}

export function cloneMonthlyBudget(mb: MonthlyCategoryBudget): MonthlyCategoryBudget {
  return { ...mb };
}

export function cloneAnnualBudget(budget: AnnualBudget): AnnualBudget {
  return {
    ...budget,
    monthlyBudgets: budget.monthlyBudgets.map(cloneMonthlyBudget),
  };
}

export function addCarryoverToBudget(
  budget: AnnualBudget,
  fromQuarter: Quarter,
  toQuarter: Quarter,
  amount: number
): AnnualBudget {
  const fromMonths = getMonthsInQuarter(fromQuarter);
  const toMonths = getMonthsInQuarter(toQuarter);
  
  let remaining = amount;
  
  const newMonthlyBudgets = budget.monthlyBudgets.map((mb) => {
    if (fromMonths.includes(mb.month) && remaining > 0) {
      const currentRemaining = getRemainingBudget(mb);
      const toDeduct = Math.min(currentRemaining, remaining);
      
      if (toDeduct > 0) {
        remaining -= toDeduct;
        return {
          ...mb,
          allocated: mb.allocated - toDeduct,
        };
      }
    }
    return mb;
  });

  let remainingToAdd = amount;
  const finalBudgets = newMonthlyBudgets.map((mb) => {
    if (toMonths.includes(mb.month) && remainingToAdd > 0) {
      const toAdd = Math.min(remainingToAdd, remainingToAdd);
      remainingToAdd -= toAdd;
      return {
        ...mb,
        allocated: mb.allocated + toAdd,
      };
    }
    return mb;
  });

  return {
    ...budget,
    monthlyBudgets: finalBudgets,
  };
}
