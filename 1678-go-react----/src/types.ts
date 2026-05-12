export type ProjectStatus = '申请中' | '初审' | '评审' | '批准' | '整改' | '不通过';

export type TodoStatus = '待办' | '已完成' | '逾期';

export interface Project {
  id: string;
  organizationName: string;
  creditCode: string;
  contactPerson: string;
  projectName: string;
  category: string;
  budgetAmount: number;
  usedBudget: number;
  implementationPeriod: number;
  status: ProjectStatus;
  submitCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface Expert {
  id: string;
  name: string;
  field: string;
  createdAt: string;
}

export interface Review {
  id: string;
  projectId: string;
  expertId: string;
  score: number;
  createdAt: string;
}

export interface AuditLog {
  id: string;
  operator: string;
  operationTime: string;
  operationType: string;
  content: string;
  beforeSnapshot: string | null;
  afterSnapshot: string | null;
  createdAt: string;
}

export interface Todo {
  id: string;
  projectId: string;
  title: string;
  description: string;
  deadline: string;
  status: TodoStatus;
  createdAt: string;
}

export interface BudgetAlert {
  id: string;
  projectId: string;
  usageRate: number;
  createdAt: string;
}

export interface FundAllocation {
  id: string;
  projectId: string;
  amount: number;
  operator: string;
  createdAt: string;
}

export interface ProjectApplication {
  organizationName: string;
  creditCode: string;
  contactPerson: string;
  projectName: string;
  category: string;
  budgetAmount: number;
  implementationPeriod: number;
}
