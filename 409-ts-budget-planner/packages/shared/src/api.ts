import {
  Department, AnnualBudget, UsageStatistic, TrendDataPoint, BudgetCategory, Month, Quarter } from './types';

export interface CreateDepartmentRequest {
  id: string;
  name: string;
  managerId: string;
}

export interface CreateBudgetRequest {
  departmentId: string;
  year: number;
  totalAmount: number;
  allocations: Array<{
    month: Month;
    category: BudgetCategory;
    amount: number;
  }>;
}

export interface SubmitReimbursementRequest {
  departmentId: string;
  year: number;
  month: Month;
  items: Array<{
    category: BudgetCategory;
    amount: number;
  }>;
  description: string;
  requestedBy: string;
}

export interface AdjustBudgetRequest {
  departmentId: string;
  year: number;
  adjustments: Array<{
    month: Month;
    category: BudgetCategory;
    newAllocated: number;
  }>;
  requestedBy: string;
}

export interface RequestCarryoverRequest {
  departmentId: string;
  year: number;
  fromQuarter: Quarter;
  toQuarter: Quarter;
  amount: number;
  requestedBy: string;
}

export interface QueryUsageRequest {
  departmentId?: string;
  category?: BudgetCategory;
  startYear?: number;
  startMonth?: Month;
  endYear?: number;
  endMonth?: Month;
}

export interface QueryTrendRequest {
  departmentId?: string;
  category?: BudgetCategory;
  startYear?: number;
  startMonth?: Month;
  endYear?: number;
  endMonth?: Month;
  interval?: 'monthly' | 'quarterly';
}

export interface CreateDraftBudgetRequest {
  newYear: number;
  sourceYear: number;
}

export interface ConfirmBudgetRequest {
  departmentId: string;
  year: number;
  confirmedBy: string;
}

export type ApiEndpoints = {
  'POST /departments': {
    request: CreateDepartmentRequest;
    response: Department;
  };
  'GET /departments': {
    request: void;
    response: Department[];
  };
  'POST /budgets': {
    request: CreateBudgetRequest;
    response: AnnualBudget;
  };
  'GET /budgets': {
    request: { departmentId: string; year: number };
    response: AnnualBudget;
  };
  'POST /reimbursements': {
    request: SubmitReimbursementRequest;
    response: { reimbursementId: string; status: 'APPROVED' | 'REJECTED' };
  };
  'POST /budgets/adjust': {
    request: AdjustBudgetRequest;
    response: AnnualBudget;
  };
  'POST /carryovers': {
    request: RequestCarryoverRequest;
    response: { carryoverId: string; status: 'APPROVED' | 'REJECTED' };
  };
  'GET /statistics/usage': {
    request: QueryUsageRequest;
    response: UsageStatistic[];
  };
  'GET /statistics/trend': {
    request: QueryTrendRequest;
    response: TrendDataPoint[];
  };
  'POST /budgets/draft': {
    request: CreateDraftBudgetRequest;
    response: AnnualBudget[];
  };
  'POST /budgets/confirm': {
    request: ConfirmBudgetRequest;
    response: AnnualBudget;
  };
};
