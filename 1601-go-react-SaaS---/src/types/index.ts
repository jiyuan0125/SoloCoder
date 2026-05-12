export type PlanId = 'basic' | 'pro' | 'enterprise';

export interface Plan {
  id: PlanId;
  name: string;
  priceMonthlyCents: number;
  maxUsers: number | null;
  apiCallsLimit: number;
  storageLimitGB: number;
  overageApiCallCostCents: number;
  overageStorageCostCentsPerGB: number;
}

export interface Subscription {
  id: string;
  customerId: string;
  planId: PlanId;
  status: 'active' | 'paused' | 'cancelled';
  startDate: Date;
  planChangeHistory: PlanChangeRecord[];
}

export interface PlanChangeRecord {
  fromPlanId: PlanId | null;
  toPlanId: PlanId;
  effectiveDate: Date;
  reason: string;
}

export interface Usage {
  id: string;
  subscriptionId: string;
  periodStart: Date;
  periodEnd: Date;
  apiCalls: number;
  storageGB: number;
  billed: boolean;
}

export interface Bill {
  id: string;
  subscriptionId: string;
  periodStart: Date;
  periodEnd: Date;
  baseFeeCents: number;
  usageFeeCents: number;
  totalAmountCents: number;
  status: 'pending' | 'paid' | 'overdue';
  generatedAt: Date;
  dueDate: Date;
  paidAt: Date | null;
}

export interface Customer {
  id: string;
  name: string;
  email: string;
}
