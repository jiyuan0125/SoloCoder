import {
  BudgetCategory, Month, Quarter,
} from '@budget-planner/shared';

export type CliCommand =
  | DepartmentListCommand
  | DepartmentCreateCommand
  | BudgetGetCommand
  | BudgetCreateCommand
  | BudgetAdjustCommand
  | BudgetDraftCommand
  | BudgetConfirmCommand
  | ReimbursementSubmitCommand
  | CarryoverRequestCommand
  | StatsUsageCommand
  | StatsTrendCommand;

export interface BaseCommand {
  serverUrl: string;
}

export interface DepartmentListCommand extends BaseCommand {
  type: 'department-list';
}

export interface DepartmentCreateCommand extends BaseCommand {
  type: 'department-create';
  id: string;
  name: string;
  managerId: string;
}

export interface BudgetGetCommand extends BaseCommand {
  type: 'budget-get';
  departmentId: string;
  year: number;
}

export interface BudgetCreateCommand extends BaseCommand {
  type: 'budget-create';
  departmentId: string;
  year: number;
  totalAmount: number;
  allocations: Array<{
    month: Month;
    category: BudgetCategory;
    amount: number;
  }>;
}

export interface BudgetAdjustCommand extends BaseCommand {
  type: 'budget-adjust';
  departmentId: string;
  year: number;
  adjustments: Array<{
    month: Month;
    category: BudgetCategory;
    newAllocated: number;
  }>;
  requestedBy: string;
}

export interface BudgetDraftCommand extends BaseCommand {
  type: 'budget-draft';
  newYear: number;
  sourceYear: number;
}

export interface BudgetConfirmCommand extends BaseCommand {
  type: 'budget-confirm';
  departmentId: string;
  year: number;
  confirmedBy: string;
}

export interface ReimbursementSubmitCommand extends BaseCommand {
  type: 'reimbursement-submit';
  departmentId: string;
  year: number;
  month: Month;
  description: string;
  requestedBy: string;
  items: Array<{
    category: BudgetCategory;
    amount: number;
  }>;
}

export interface CarryoverRequestCommand extends BaseCommand {
  type: 'carryover-request';
  departmentId: string;
  year: number;
  fromQuarter: Quarter;
  toQuarter: Quarter;
  amount: number;
  requestedBy: string;
}

export interface StatsUsageCommand extends BaseCommand {
  type: 'stats-usage';
  departmentId?: string;
  category?: BudgetCategory;
  startYear?: number;
  startMonth?: Month;
  endYear?: number;
  endMonth?: Month;
}

export interface StatsTrendCommand extends BaseCommand {
  type: 'stats-trend';
  departmentId?: string;
  category?: BudgetCategory;
  startYear?: number;
  startMonth?: Month;
  endYear?: number;
  endMonth?: Month;
  interval?: 'monthly' | 'quarterly';
}
