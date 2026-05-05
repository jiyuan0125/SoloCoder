import {
  EXPENSE_CATEGORIES,
  EXPENSE_STATUSES,
  DINING_LIMIT_CENTS,
  ACCOMMODATION_NIGHT_LIMIT_CENTS,
  LARGE_AMOUNT_THRESHOLD_CENTS,
} from '@expense-report/shared';
import type { ExpenseCategory, ExpenseStatus } from '@expense-report/shared';

export function generateId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 10);
  return `${timestamp}-${random}`;
}

export function getCurrentTime(): string {
  return new Date().toISOString();
}

export function isValidDate(dateString: string): boolean {
  const date = new Date(dateString);
  return !isNaN(date.getTime());
}

export function isValidAmount(amount: unknown): amount is number {
  return typeof amount === 'number' && Number.isInteger(amount) && amount >= 0;
}

export function isValidCategory(category: string): category is ExpenseCategory {
  return EXPENSE_CATEGORIES.includes(category as ExpenseCategory);
}

export function isValidStatus(status: string): status is ExpenseStatus {
  return EXPENSE_STATUSES.includes(status as ExpenseStatus);
}

export function isLargeAmount(amountCents: number): boolean {
  return amountCents > LARGE_AMOUNT_THRESHOLD_CENTS;
}

export function checkExpenseLimit(
  category: ExpenseCategory,
  amountCents: number
): { exceedsLimit: boolean; limitCents: number; needsSupplementary: boolean } {
  let limitCents = Number.MAX_SAFE_INTEGER;
  let needsSupplementary = false;

  if (category === 'dining') {
    limitCents = DINING_LIMIT_CENTS;
    if (amountCents > limitCents) {
      needsSupplementary = true;
    }
  } else if (category === 'accommodation') {
    limitCents = ACCOMMODATION_NIGHT_LIMIT_CENTS;
    if (amountCents > limitCents) {
      needsSupplementary = true;
    }
  }

  return {
    exceedsLimit: amountCents > limitCents,
    limitCents,
    needsSupplementary,
  };
}

export function formatAmountCents(amountCents: number): string {
  const yuan = Math.floor(amountCents / 100);
  const fen = amountCents % 100;
  return `${yuan}.${fen.toString().padStart(2, '0')} 元`;
}

export function parseAmountToCents(amountStr: string): number | null {
  const trimmed = amountStr.trim();
  const match = trimmed.match(/^(\d+)(?:\.(\d{1,2}))?$/);
  if (!match) {
    return null;
  }
  const yuan = parseInt(match[1] ?? '0', 10);
  const fenStr = (match[2] ?? '0').padEnd(2, '0');
  const fen = parseInt(fenStr, 10);
  return yuan * 100 + fen;
}

export function isDateInRange(
  dateStr: string,
  startDate?: string,
  endDate?: string
): boolean {
  const date = new Date(dateStr);
  const dateTime = date.setHours(0, 0, 0, 0);

  if (startDate) {
    const start = new Date(startDate);
    const startTime = start.setHours(0, 0, 0, 0);
    if (dateTime < startTime) {
      return false;
    }
  }

  if (endDate) {
    const end = new Date(endDate);
    const endTime = end.setHours(23, 59, 59, 999);
    if (dateTime > endTime) {
      return false;
    }
  }

  return true;
}
