import {
  BudgetCategory, Month, Quarter,
} from '@budget-planner/shared';

const VALID_CATEGORIES: BudgetCategory[] = ['TRAVEL', 'OFFICE', 'EQUIPMENT', 'TRAINING', 'OTHER'];
const VALID_MONTHS: Month[] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12];
const VALID_QUARTERS: Quarter[] = [1, 2, 3, 4];

export function isValidCategory(category: unknown): category is BudgetCategory {
  return VALID_CATEGORIES.includes(category as BudgetCategory);
}

export function isValidMonth(month: unknown): month is Month {
  return typeof month === 'number' && VALID_MONTHS.includes(month as Month);
}

export function isValidQuarter(quarter: unknown): quarter is Quarter {
  return typeof quarter === 'number' && VALID_QUARTERS.includes(quarter as Quarter);
}

export function isValidAmount(amount: unknown): amount is number {
  return typeof amount === 'number' && Number.isInteger(amount) && amount >= 0;
}

export function isValidYear(year: unknown): year is number {
  return typeof year === 'number' && Number.isInteger(year) && year > 0;
}

export function isValidString(value: unknown): value is string {
  return typeof value === 'string' && value.trim().length > 0;
}

export function parseAsMonth(value: unknown): Month | null {
  if (typeof value === 'string') {
    const num = parseInt(value, 10);
    if (!isNaN(num) && isValidMonth(num)) {
      return num;
    }
  }
  if (isValidMonth(value)) {
    return value;
  }
  return null;
}

export function parseAsQuarter(value: unknown): Quarter | null {
  if (typeof value === 'string') {
    const num = parseInt(value, 10);
    if (!isNaN(num) && isValidQuarter(num)) {
      return num;
    }
  }
  if (isValidQuarter(value)) {
    return value;
  }
  return null;
}

export function parseAsNumber(value: unknown): number | null {
  if (typeof value === 'number') {
    return value;
  }
  if (typeof value === 'string') {
    const num = parseInt(value, 10);
    if (!isNaN(num)) {
      return num;
    }
  }
  return null;
}

export function parseAsCategory(value: unknown): BudgetCategory | null {
  if (isValidCategory(value)) {
    return value;
  }
  return null;
}
