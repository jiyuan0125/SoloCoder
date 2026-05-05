export type BudgetCategory = 
  | 'TRAVEL'
  | 'OFFICE'
  | 'EQUIPMENT'
  | 'TRAINING'
  | 'OTHER';

export type Month = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12;

export type Quarter = 1 | 2 | 3 | 4;

export type BudgetStatus = 'DRAFT' | 'ACTIVE' | 'ARCHIVED';

export interface MonthlyCategoryBudget {
  month: Month;
  category: BudgetCategory;
  allocated: number;
  used: number;
}

export interface Department {
  id: string;
  name: string;
  managerId: string;
}

export interface AnnualBudget {
  id: string;
  departmentId: string;
  year: number;
  totalAmount: number;
  monthlyBudgets: MonthlyCategoryBudget[];
  status: BudgetStatus;
  createdAt: string;
  confirmedAt?: string;
}

export interface ReimbursementItem {
  category: BudgetCategory;
  amount: number;
}

export interface ReimbursementRequest {
  id: string;
  departmentId: string;
  year: number;
  month: Month;
  items: ReimbursementItem[];
  description: string;
  requestedBy: string;
  requestedAt: string;
}

export interface ReimbursementRecord extends ReimbursementRequest {
  status: 'APPROVED' | 'REJECTED';
  rejectedReason?: string;
  processedAt: string;
}

export interface CarryoverRequest {
  id: string;
  departmentId: string;
  year: number;
  fromQuarter: Quarter;
  toQuarter: Quarter;
  amount: number;
  requestedBy: string;
  requestedAt: string;
}

export interface CarryoverRecord extends CarryoverRequest {
  status: 'APPROVED' | 'REJECTED';
  rejectedReason?: string;
  processedAt: string;
}

export interface BudgetAdjustmentRequest {
  id: string;
  departmentId: string;
  year: number;
  adjustments: Array<{
    month: Month;
    category: BudgetCategory;
    newAllocated: number;
  }>;
  requestedBy: string;
  requestedAt: string;
}

export interface UsageStatistic {
  departmentId: string;
  departmentName: string;
  category: BudgetCategory;
  year: number;
  month: Month;
  allocated: number;
  used: number;
  usageRate: number;
}

export interface TrendDataPoint {
  period: string;
  departmentId: string;
  departmentName: string;
  category: BudgetCategory;
  allocated: number;
  used: number;
  usageRate: number;
}
