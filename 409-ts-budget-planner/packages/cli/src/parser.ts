import { CliCommand } from './types';
import { BudgetCategory, Month, Quarter } from '@budget-planner/shared';

const VALID_CATEGORIES: BudgetCategory[] = ['TRAVEL', 'OFFICE', 'EQUIPMENT', 'TRAINING', 'OTHER'];
const VALID_MONTHS: Month[] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12];
const VALID_QUARTERS: Quarter[] = [1, 2, 3, 4];

function parseAsInt(value: string): number | null {
  const num = parseInt(value, 10);
  if (isNaN(num)) return null;
  return num;
}

function parseAsMonth(value: string): Month | null {
  const num = parseAsInt(value);
  if (num === null || !VALID_MONTHS.includes(num as Month)) return null;
  return num as Month;
}

function parseAsQuarter(value: string): Quarter | null {
  const num = parseAsInt(value);
  if (num === null || !VALID_QUARTERS.includes(num as Quarter)) return null;
  return num as Quarter;
}

function parseAsCategory(value: string): BudgetCategory | null {
  if (VALID_CATEGORIES.includes(value as BudgetCategory)) {
    return value as BudgetCategory;
  }
  return null;
}

function parseJsonArray<T>(value: string, validator: (item: unknown) => item is T): T[] {
  const parsed = JSON.parse(value);
  if (!Array.isArray(parsed)) {
    throw new Error('Expected JSON array');
  }
  for (const item of parsed) {
    if (!validator(item)) {
      throw new Error(`Invalid item in array: ${JSON.stringify(item)}`);
    }
  }
  return parsed;
}

function isAllocationItem(item: unknown): item is { month: Month; category: BudgetCategory; amount: number } {
  if (!item || typeof item !== 'object') return false;
  const obj = item as Record<string, unknown>;
  return (
    typeof obj.month === 'number' &&
    VALID_MONTHS.includes(obj.month as Month) &&
    typeof obj.category === 'string' &&
    VALID_CATEGORIES.includes(obj.category as BudgetCategory) &&
    typeof obj.amount === 'number' &&
    Number.isInteger(obj.amount) &&
    obj.amount >= 0
  );
}

function isAdjustmentItem(item: unknown): item is { month: Month; category: BudgetCategory; newAllocated: number } {
  if (!item || typeof item !== 'object') return false;
  const obj = item as Record<string, unknown>;
  return (
    typeof obj.month === 'number' &&
    VALID_MONTHS.includes(obj.month as Month) &&
    typeof obj.category === 'string' &&
    VALID_CATEGORIES.includes(obj.category as BudgetCategory) &&
    typeof obj.newAllocated === 'number' &&
    Number.isInteger(obj.newAllocated) &&
    obj.newAllocated >= 0
  );
}

function isReimbursementItem(item: unknown): item is { category: BudgetCategory; amount: number } {
  if (!item || typeof item !== 'object') return false;
  const obj = item as Record<string, unknown>;
  return (
    typeof obj.category === 'string' &&
    VALID_CATEGORIES.includes(obj.category as BudgetCategory) &&
    typeof obj.amount === 'number' &&
    Number.isInteger(obj.amount) &&
    obj.amount > 0
  );
}

interface ParsedArgs {
  positional: string[];
  options: Record<string, string>;
}

function parseArgs(args: string[]): ParsedArgs {
  const positional: string[] = [];
  const options: Record<string, string> = {};
  
  let i = 0;
  while (i < args.length) {
    const arg = args[i];
    
    if (arg.startsWith('--')) {
      const key = arg.slice(2);
      if (i + 1 < args.length && !args[i + 1].startsWith('--')) {
        options[key] = args[i + 1];
        i += 2;
      } else {
        options[key] = 'true';
        i += 1;
      }
    } else {
      positional.push(arg);
      i += 1;
    }
  }
  
  return { positional, options };
}

export function parseCommand(args: string[]): CliCommand | null {
  const { positional, options } = parseArgs(args);
  
  if (options.help !== undefined) {
    return null;
  }

  const serverUrl = options.server || 'http://localhost:3000';

  if (positional.length === 0) {
    return null;
  }

  const [command1, command2] = positional;

  if (command1 === 'department') {
    if (command2 === 'list') {
      return { type: 'department-list', serverUrl };
    }
    if (command2 === 'create' && positional.length >= 5) {
      return {
        type: 'department-create',
        serverUrl,
        id: positional[2],
        name: positional[3],
        managerId: positional[4],
      };
    }
  }

  if (command1 === 'budget') {
    if (command2 === 'get' && positional.length >= 4) {
      const year = parseAsInt(positional[3]);
      if (year !== null) {
        return {
          type: 'budget-get',
          serverUrl,
          departmentId: positional[2],
          year,
        };
      }
    }
    if (command2 === 'create' && positional.length >= 5 && options.allocations) {
      const year = parseAsInt(positional[3]);
      const totalAmount = parseAsInt(positional[4]);
      if (year !== null && totalAmount !== null) {
        const allocations = parseJsonArray(options.allocations, isAllocationItem);
        return {
          type: 'budget-create',
          serverUrl,
          departmentId: positional[2],
          year,
          totalAmount,
          allocations,
        };
      }
    }
    if (command2 === 'adjust' && positional.length >= 4 && options.adjustments) {
      const year = parseAsInt(positional[3]);
      if (year !== null) {
        const adjustments = parseJsonArray(options.adjustments, isAdjustmentItem);
        return {
          type: 'budget-adjust',
          serverUrl,
          departmentId: positional[2],
          year,
          adjustments,
          requestedBy: options.requestedBy || 'cli-user',
        };
      }
    }
    if (command2 === 'draft' && positional.length >= 4) {
      const newYear = parseAsInt(positional[2]);
      const sourceYear = parseAsInt(positional[3]);
      if (newYear !== null && sourceYear !== null) {
        return {
          type: 'budget-draft',
          serverUrl,
          newYear,
          sourceYear,
        };
      }
    }
    if (command2 === 'confirm' && positional.length >= 5) {
      const year = parseAsInt(positional[3]);
      if (year !== null) {
        return {
          type: 'budget-confirm',
          serverUrl,
          departmentId: positional[2],
          year,
          confirmedBy: positional[4],
        };
      }
    }
  }

  if (command1 === 'reimbursement') {
    if (command2 === 'submit' && positional.length >= 7 && options.items) {
      const year = parseAsInt(positional[3]);
      const month = parseAsMonth(positional[4]);
      if (year !== null && month !== null) {
        const items = parseJsonArray(options.items, isReimbursementItem);
        return {
          type: 'reimbursement-submit',
          serverUrl,
          departmentId: positional[2],
          year,
          month,
          description: positional[5],
          requestedBy: positional[6],
          items,
        };
      }
    }
  }

  if (command1 === 'carryover') {
    if (command2 === 'request' && positional.length >= 8) {
      const year = parseAsInt(positional[3]);
      const fromQuarter = parseAsQuarter(positional[4]);
      const toQuarter = parseAsQuarter(positional[5]);
      const amount = parseAsInt(positional[6]);
      if (year !== null && fromQuarter !== null && toQuarter !== null && amount !== null) {
        return {
          type: 'carryover-request',
          serverUrl,
          departmentId: positional[2],
          year,
          fromQuarter,
          toQuarter,
          amount,
          requestedBy: positional[7],
        };
      }
    }
  }

  if (command1 === 'stats') {
    if (command2 === 'usage') {
      const cmd: CliCommand = {
        type: 'stats-usage',
        serverUrl,
      };
      if (options.departmentId) cmd.departmentId = options.departmentId;
      if (options.category) {
        const cat = parseAsCategory(options.category);
        if (cat !== null) cmd.category = cat;
      }
      if (options.startYear) {
        const y = parseAsInt(options.startYear);
        if (y !== null) cmd.startYear = y;
      }
      if (options.startMonth) {
        const m = parseAsMonth(options.startMonth);
        if (m !== null) cmd.startMonth = m;
      }
      if (options.endYear) {
        const y = parseAsInt(options.endYear);
        if (y !== null) cmd.endYear = y;
      }
      if (options.endMonth) {
        const m = parseAsMonth(options.endMonth);
        if (m !== null) cmd.endMonth = m;
      }
      return cmd;
    }
    if (command2 === 'trend') {
      const cmd: CliCommand = {
        type: 'stats-trend',
        serverUrl,
      };
      if (options.departmentId) cmd.departmentId = options.departmentId;
      if (options.category) {
        const cat = parseAsCategory(options.category);
        if (cat !== null) cmd.category = cat;
      }
      if (options.startYear) {
        const y = parseAsInt(options.startYear);
        if (y !== null) cmd.startYear = y;
      }
      if (options.startMonth) {
        const m = parseAsMonth(options.startMonth);
        if (m !== null) cmd.startMonth = m;
      }
      if (options.endYear) {
        const y = parseAsInt(options.endYear);
        if (y !== null) cmd.endYear = y;
      }
      if (options.endMonth) {
        const m = parseAsMonth(options.endMonth);
        if (m !== null) cmd.endMonth = m;
      }
      if (options.interval === 'monthly' || options.interval === 'quarterly') {
        cmd.interval = options.interval;
      }
      return cmd;
    }
  }

  throw new Error(`Invalid command: ${positional.join(' ')}`);
}
