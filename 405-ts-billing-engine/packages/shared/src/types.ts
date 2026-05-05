export type PlanType = 'free' | 'basic' | 'pro';
export type BillingCycle = 'monthly' | 'yearly';
export type BillStatus = 'draft' | 'pending' | 'paid' | 'overdue' | 'reviewing' | 'reviewed';
export type CustomerStatus = 'active' | 'frozen' | 'trial';
export type PaymentStatus = 'pending' | 'completed' | 'failed';
export type RefundStatus = 'pending' | 'approved' | 'rejected';

export interface Plan {
  type: PlanType;
  name: string;
  monthlyFee: number;
  yearlyFee: number;
  monthlyQuota: number;
  overageFeePerUnit: number;
}

export interface Customer {
  id: string;
  name: string;
  email: string;
  status: CustomerStatus;
  trialEndAt: string | null;
  frozenAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Subscription {
  id: string;
  customerId: string;
  planType: PlanType;
  billingCycle: BillingCycle;
  startsAt: string;
  endsAt: string | null;
  isActive: boolean;
  isYearlyPaid: boolean;
  yearlyPaymentAmount: number;
  yearlyPaymentDate: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface UsageRecord {
  id: string;
  customerId: string;
  units: number;
  recordedAt: string;
  billingPeriodStart: string;
  billingPeriodEnd: string;
  isFreeQuota: boolean;
}

export interface Bill {
  id: string;
  customerId: string;
  subscriptionId: string;
  periodStart: string;
  periodEnd: string;
  baseFee: number;
  overageFee: number;
  discount: number;
  totalAmount: number;
  status: BillStatus;
  dueAt: string;
  paidAt: string | null;
  reviewRequestedAt: string | null;
  reviewDeadline: string;
  notes: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface BillItem {
  id: string;
  billId: string;
  type: 'base_fee' | 'overage' | 'discount' | 'frozen_usage';
  description: string;
  units: number;
  unitPrice: number;
  amount: number;
}

export interface Payment {
  id: string;
  customerId: string;
  billId: string | null;
  amount: number;
  status: PaymentStatus;
  transactionId: string | null;
  paidAt: string | null;
  createdAt: string;
}

export interface Refund {
  id: string;
  customerId: string;
  subscriptionId: string;
  amount: number;
  refundedMonths: number;
  status: RefundStatus;
  reason: string;
  processedAt: string | null;
  createdAt: string;
}

export interface MonthlyUsage {
  customerId: string;
  subscriptionId: string;
  periodStart: string;
  periodEnd: string;
  totalUnits: number;
  quotaUsed: number;
  overageUnits: number;
}

export interface UpgradeProration {
  fromPlan: PlanType;
  toPlan: PlanType;
  upgradeDate: string;
  periodStart: string;
  periodEnd: string;
  daysInPeriod: number;
  daysOnOldPlan: number;
  daysOnNewPlan: number;
  oldPlanDailyRate: number;
  newPlanDailyRate: number;
  oldPlanProratedFee: number;
  newPlanProratedFee: number;
  totalBaseFee: number;
}

export const PLANS: Readonly<Record<PlanType, Plan>> = {
  free: {
    type: 'free',
    name: '免费版',
    monthlyFee: 0,
    yearlyFee: 0,
    monthlyQuota: 1000,
    overageFeePerUnit: 0,
  },
  basic: {
    type: 'basic',
    name: '基础版',
    monthlyFee: 9900,
    yearlyFee: 9900 * 12,
    monthlyQuota: 10000,
    overageFeePerUnit: 1,
  },
  pro: {
    type: 'pro',
    name: '专业版',
    monthlyFee: 29900,
    yearlyFee: 29900 * 12,
    monthlyQuota: 100000,
    overageFeePerUnit: 1,
  },
} as const;

export const TRIAL_DAYS = 7;
export const BILL_DUE_DAYS = 15;
export const REVIEW_WINDOW_DAYS = 15;
export const FREEZE_AFTER_MONTHS_OVERDUE = 2;
export const YEARLY_DISCOUNT_PERCENT = 10;
export const REFUND_FEE_PERCENT = 20;
export const REFUND_ELIGIBLE_MONTH = 10;
