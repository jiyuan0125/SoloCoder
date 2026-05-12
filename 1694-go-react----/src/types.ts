export enum ProjectStatus {
  SUBMITTED = 'SUBMITTED',
  TOWNSHIP_REVIEW_PENDING = 'TOWNSHIP_REVIEW_PENDING',
  TOWNSHIP_REVIEW_PASSED = 'TOWNSHIP_REVIEW_PASSED',
  TOWNSHIP_REVIEW_REJECTED = 'TOWNSHIP_REVIEW_REJECTED',
  COUNTY_REVIEW_PENDING = 'COUNTY_REVIEW_PENDING',
  COUNTY_REVIEW_PASSED = 'COUNTY_REVIEW_PASSED',
  COUNTY_REVIEW_REJECTED = 'COUNTY_REVIEW_REJECTED',
  APPROVED = 'APPROVED',
  INITIAL_FUND_ALLOCATED = 'INITIAL_FUND_ALLOCATED',
  MID_TERM_INSPECTION_PENDING = 'MID_TERM_INSPECTION_PENDING',
  MID_TERM_INSPECTION_PASSED = 'MID_TERM_INSPECTION_PASSED',
  MID_TERM_RECTIFICATION_REQUIRED = 'MID_TERM_RECTIFICATION_REQUIRED',
  MID_TERM_INSPECTION_FAILED = 'MID_TERM_INSPECTION_FAILED',
  MID_TERM_FUND_ALLOCATED = 'MID_TERM_FUND_ALLOCATED',
  FINAL_INSPECTION_PENDING = 'FINAL_INSPECTION_PENDING',
  FINAL_INSPECTION_PASSED = 'FINAL_INSPECTION_PASSED',
  FINAL_RECTIFICATION_REQUIRED = 'FINAL_RECTIFICATION_REQUIRED',
  FINAL_INSPECTION_FAILED = 'FINAL_INSPECTION_FAILED',
  FINAL_FUND_ALLOCATED = 'FINAL_FUND_ALLOCATED',
  COMPLETED = 'COMPLETED',
  TERMINATED = 'TERMINATED'
}

export enum ReviewResult {
  PASSED = 'PASSED',
  REJECTED = 'REJECTED'
}

export enum InspectionResult {
  PASSED = 'PASSED',
  PASSED_AFTER_RECTIFICATION = 'PASSED_AFTER_RECTIFICATION',
  FAILED = 'FAILED'
}

export enum BudgetCategory {
  CONSTRUCTION = 'construction',
  EQUIPMENT = 'equipment',
  LABOR = 'labor',
  OTHER = 'other'
}

export interface BudgetBreakdown {
  construction: number;
  equipment: number;
  labor: number;
  other: number;
}

export interface Project {
  id: number;
  name: string;
  category: string;
  location: string;
  totalBudget: number;
  implementationPeriod: string;
  status: ProjectStatus;
  resubmissionCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectBudget {
  id: number;
  projectId: number;
  construction: number;
  equipment: number;
  labor: number;
  other: number;
  createdAt: string;
}

export interface ProjectExpenditure {
  id: number;
  projectId: number;
  category: BudgetCategory;
  amount: number;
  description: string;
  expenditureDate: string;
  createdAt: string;
}

export interface FundAllocation {
  id: number;
  projectId: number;
  stage: 'initial' | 'mid_term' | 'final';
  amount: number;
  percentage: number;
  allocationDate: string;
  paymentMethod: string;
  receivingAccount: string;
  createdAt: string;
}

export interface ReviewRecord {
  id: number;
  projectId: number;
  reviewLevel: 'township' | 'county';
  result: ReviewResult;
  comments: string | null;
  reviewer: string;
  reviewedAt: string;
  createdAt: string;
}

export interface InspectionRecord {
  id: number;
  projectId: number;
  stage: 'mid_term' | 'final';
  result: InspectionResult;
  comments: string | null;
  rectificationDeadline: string | null;
  inspector: string;
  inspectedAt: string;
  createdAt: string;
}

export interface BudgetAdjustmentRequest {
  id: number;
  projectId: number;
  fromCategory: BudgetCategory;
  toCategory: BudgetCategory;
  amount: number;
  reason: string;
  status: 'pending' | 'approved' | 'rejected';
  approver: string | null;
  approvedAt: string | null;
  createdAt: string;
}
